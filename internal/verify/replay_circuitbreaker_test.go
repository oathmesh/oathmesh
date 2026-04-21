package verify

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// failingReplayCache simulates a Redis cache that fails after a configured number of calls.
type failingReplayCache struct {
	mu          sync.Mutex
	callCount   int
	failAfter   int  // fail after this many successful calls (-1 = always fail)
	recovered   bool // if true, stop failing
	checkCalled int
}

func (f *failingReplayCache) Check(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.checkCalled++

	if f.recovered {
		return false, nil
	}
	if f.failAfter == -1 || f.callCount >= f.failAfter {
		return false, fmt.Errorf("redis connection refused")
	}
	f.callCount++
	return false, nil
}

func (f *failingReplayCache) setRecovered(v bool) {
	f.mu.Lock()
	f.recovered = v
	f.mu.Unlock()
}

func TestCircuitBreaker_ClosedState_RoutesToPrimary(t *testing.T) {
	primary := &failingReplayCache{failAfter: 100}
	cb := NewCircuitBreakerReplayCache(primary, CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  50 * time.Millisecond,
	})
	defer cb.Close()

	replayed, err := cb.Check(context.Background(), "jti-1", 30*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if replayed {
		t.Fatal("expected not replayed")
	}
	if cb.State() != CircuitClosed {
		t.Fatalf("expected CircuitClosed, got %d", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	primary := &failingReplayCache{failAfter: -1} // always fail
	cb := NewCircuitBreakerReplayCache(primary, CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  50 * time.Millisecond,
	})
	defer cb.Close()

	// 3 failures should open the circuit
	for i := 0; i < 3; i++ {
		_, err := cb.Check(context.Background(), fmt.Sprintf("jti-%d", i), 30*time.Second)
		if err != nil {
			t.Fatalf("expected fallback to succeed, got: %v", err)
		}
	}

	if cb.State() != CircuitOpen {
		t.Fatalf("expected CircuitOpen after %d failures, got %d", 3, cb.State())
	}
}

func TestCircuitBreaker_FallbackPreservesReplayProtection(t *testing.T) {
	primary := &failingReplayCache{failAfter: -1}
	cb := NewCircuitBreakerReplayCache(primary, CircuitBreakerConfig{
		FailureThreshold: 1,
		RecoveryTimeout:  1 * time.Hour, // don't probe during this test
	})
	defer cb.Close()

	// First call: triggers circuit open, falls back, records jti
	_, _ = cb.Check(context.Background(), "jti-unique", 30*time.Second)

	if cb.State() != CircuitOpen {
		t.Fatalf("expected CircuitOpen, got %d", cb.State())
	}

	// Second call with same jti: should detect replay via fallback
	replayed, err := cb.Check(context.Background(), "jti-unique", 30*time.Second)
	if err != nil {
		t.Fatalf("expected no error from fallback, got: %v", err)
	}
	if !replayed {
		t.Fatal("expected replay detected via fallback MemoryReplayCache")
	}
}

func TestCircuitBreaker_RecoveryProbe(t *testing.T) {
	primary := &failingReplayCache{failAfter: -1}
	cb := NewCircuitBreakerReplayCache(primary, CircuitBreakerConfig{
		FailureThreshold: 1,
		RecoveryTimeout:  50 * time.Millisecond,
	})
	defer cb.Close()

	// Trigger circuit open
	_, _ = cb.Check(context.Background(), "jti-1", 30*time.Second)
	if cb.State() != CircuitOpen {
		t.Fatalf("expected CircuitOpen, got %d", cb.State())
	}

	// Simulate Redis recovery
	primary.setRecovered(true)

	// Wait for recovery timeout
	time.Sleep(60 * time.Millisecond)

	// Next call should probe and recover
	replayed, err := cb.Check(context.Background(), "jti-probe", 30*time.Second)
	if err != nil {
		t.Fatalf("expected no error after recovery, got: %v", err)
	}
	if replayed {
		t.Fatal("expected not replayed on fresh jti")
	}
	if cb.State() != CircuitClosed {
		t.Fatalf("expected CircuitClosed after recovery, got %d", cb.State())
	}
}

func TestCircuitBreaker_FailedProbeStaysOpen(t *testing.T) {
	primary := &failingReplayCache{failAfter: -1}
	cb := NewCircuitBreakerReplayCache(primary, CircuitBreakerConfig{
		FailureThreshold: 1,
		RecoveryTimeout:  50 * time.Millisecond,
	})
	defer cb.Close()

	// Trigger circuit open
	_, _ = cb.Check(context.Background(), "jti-1", 30*time.Second)

	// Wait for recovery timeout
	time.Sleep(60 * time.Millisecond)

	// Probe — should fail and stay open
	_, err := cb.Check(context.Background(), "jti-probe-fail", 30*time.Second)
	if err != nil {
		t.Fatalf("expected fallback success, got: %v", err)
	}
	if cb.State() != CircuitOpen {
		t.Fatalf("expected CircuitOpen after failed probe, got %d", cb.State())
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	primary := &failingReplayCache{failAfter: -1}
	cb := NewCircuitBreakerReplayCache(primary, CircuitBreakerConfig{
		FailureThreshold: 1,
		RecoveryTimeout:  1 * time.Hour,
	})
	defer cb.Close()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			_, err := cb.Check(context.Background(), fmt.Sprintf("jti-concurrent-%d", id), 30*time.Second)
			if err != nil {
				t.Errorf("goroutine %d: unexpected error: %v", id, err)
			}
		}(i)
	}
	wg.Wait()
}
