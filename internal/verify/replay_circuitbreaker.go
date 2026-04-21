package verify

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
)

// CircuitState represents the current state of the circuit breaker.
type CircuitState int32

const (
	// CircuitClosed means Redis is healthy — all requests go to Redis.
	CircuitClosed CircuitState = iota
	// CircuitOpen means Redis is unhealthy — all requests go to the fallback MemoryReplayCache.
	CircuitOpen
	// CircuitHalfOpen means we're probing Redis to see if it recovered.
	CircuitHalfOpen
)

const (
	// defaultFailureThreshold is how many consecutive Redis errors before the circuit opens.
	defaultFailureThreshold = 3
	// defaultRecoveryTimeout is how long to wait before probing Redis again.
	defaultRecoveryTimeout = 30 * time.Second
)

// CircuitBreakerReplayCache wraps a primary ReplayCache (typically Redis) with
// circuit-breaker logic. When the primary fails, it falls over to an in-process
// MemoryReplayCache to preserve replay protection with single-instance coverage.
//
// This is strictly better than fail-open (which would allow replays during outage)
// or fail-closed (which would deny all requests during outage).
//
// On failover, an audit-level log is emitted with "replay_cache": "degraded" so
// operators know the circuit breaker is active.
type CircuitBreakerReplayCache struct {
	primary  core.ReplayCache
	fallback *MemoryReplayCache

	state            atomic.Int32 // CircuitState
	consecutiveFails atomic.Int64
	lastFailure      atomic.Int64 // Unix nano

	failureThreshold int
	recoveryTimeout  time.Duration

	mu             sync.Mutex // protects probe transitions
	degradedLogged atomic.Bool
}

// CircuitBreakerConfig configures the circuit breaker behavior.
type CircuitBreakerConfig struct {
	// FailureThreshold is how many consecutive Redis errors trigger the circuit to open.
	// Default: 3.
	FailureThreshold int

	// RecoveryTimeout is how long to wait in the open state before probing Redis again.
	// Default: 30s.
	RecoveryTimeout time.Duration
}

// NewCircuitBreakerReplayCache wraps a primary replay cache (e.g. Redis) with
// circuit-breaker failover to an in-process MemoryReplayCache.
func NewCircuitBreakerReplayCache(primary core.ReplayCache, cfg CircuitBreakerConfig) *CircuitBreakerReplayCache {
	threshold := cfg.FailureThreshold
	if threshold <= 0 {
		threshold = defaultFailureThreshold
	}
	timeout := cfg.RecoveryTimeout
	if timeout <= 0 {
		timeout = defaultRecoveryTimeout
	}

	return &CircuitBreakerReplayCache{
		primary:          primary,
		fallback:         NewMemoryReplayCache(),
		failureThreshold: threshold,
		recoveryTimeout:  timeout,
	}
}

// Check implements core.ReplayCache. Routes to primary or fallback based on circuit state.
func (cb *CircuitBreakerReplayCache) Check(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	switch CircuitState(cb.state.Load()) {
	case CircuitClosed:
		return cb.checkPrimary(ctx, jti, ttl)

	case CircuitOpen:
		// Check if recovery timeout has elapsed — if so, attempt a probe.
		if cb.shouldProbe() {
			return cb.checkProbe(ctx, jti, ttl)
		}
		// Still in open state — use fallback.
		return cb.checkFallback(ctx, jti, ttl)

	case CircuitHalfOpen:
		// Another goroutine is already probing — use fallback.
		return cb.checkFallback(ctx, jti, ttl)

	default:
		return cb.checkFallback(ctx, jti, ttl)
	}
}

// checkPrimary routes to the primary cache and tracks failures.
func (cb *CircuitBreakerReplayCache) checkPrimary(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	replayed, err := cb.primary.Check(ctx, jti, ttl)
	if err != nil {
		cb.recordFailure()
		// Fall through to fallback for this request.
		return cb.checkFallback(ctx, jti, ttl)
	}
	cb.recordSuccess()
	return replayed, nil
}

// checkProbe attempts a single request to the primary to test recovery.
func (cb *CircuitBreakerReplayCache) checkProbe(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	cb.mu.Lock()
	// Double-check: another goroutine may have already transitioned.
	if CircuitState(cb.state.Load()) != CircuitOpen {
		cb.mu.Unlock()
		return cb.Check(ctx, jti, ttl) // Re-route through main logic.
	}
	cb.state.Store(int32(CircuitHalfOpen))
	cb.mu.Unlock()

	replayed, err := cb.primary.Check(ctx, jti, ttl)
	if err != nil {
		// Probe failed — back to open.
		cb.state.Store(int32(CircuitOpen))
		cb.lastFailure.Store(time.Now().UnixNano())
		return cb.checkFallback(ctx, jti, ttl)
	}

	// Probe succeeded — close the circuit.
	cb.state.Store(int32(CircuitClosed))
	cb.consecutiveFails.Store(0)
	cb.degradedLogged.Store(false)

	slog.Info("oathmesh: replay cache circuit breaker recovered",
		"replay_cache", "recovered",
		"component", "circuit_breaker",
	)

	return replayed, nil
}

// checkFallback routes to the in-process MemoryReplayCache.
func (cb *CircuitBreakerReplayCache) checkFallback(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	if !cb.degradedLogged.Swap(true) {
		slog.Warn("oathmesh: replay cache degraded — using in-process fallback",
			"replay_cache", "degraded",
			"component", "circuit_breaker",
			"reason", "primary replay cache unreachable",
		)
	}
	return cb.fallback.Check(ctx, jti, ttl)
}

// recordFailure increments the consecutive failure counter and opens the circuit
// if the threshold is reached.
func (cb *CircuitBreakerReplayCache) recordFailure() {
	fails := cb.consecutiveFails.Add(1)
	cb.lastFailure.Store(time.Now().UnixNano())

	if int(fails) >= cb.failureThreshold {
		cb.state.Store(int32(CircuitOpen))
	}
}

// recordSuccess resets the failure counter on a successful primary call.
func (cb *CircuitBreakerReplayCache) recordSuccess() {
	cb.consecutiveFails.Store(0)
}

// shouldProbe returns true if enough time has passed since the last failure
// to attempt a recovery probe.
func (cb *CircuitBreakerReplayCache) shouldProbe() bool {
	lastFail := time.Unix(0, cb.lastFailure.Load())
	return time.Since(lastFail) >= cb.recoveryTimeout
}

// State returns the current circuit state (for metrics/testing).
func (cb *CircuitBreakerReplayCache) State() CircuitState {
	return CircuitState(cb.state.Load())
}

// Close shuts down the fallback MemoryReplayCache cleanup goroutine.
func (cb *CircuitBreakerReplayCache) Close() {
	cb.fallback.Close()
}

// Compile-time check that CircuitBreakerReplayCache implements core.ReplayCache.
var _ core.ReplayCache = (*CircuitBreakerReplayCache)(nil)
