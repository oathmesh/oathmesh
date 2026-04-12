package verify

import (
	"context"
	"sync"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
)

// MemoryReplayCache implements core.ReplayCache for dev and single-instance deployments.
// Protected by sync.RWMutex — reads under RLock, writes under Lock.
// A background goroutine periodically cleans up expired entries.
type MemoryReplayCache struct {
	mu      sync.RWMutex
	entries map[string]time.Time // jti → expiration time
	done    chan struct{}
}

// NewMemoryReplayCache creates a new in-memory replay cache.
// Starts a background cleanup goroutine that runs every 60 seconds.
func NewMemoryReplayCache() *MemoryReplayCache {
	rc := &MemoryReplayCache{
		entries: make(map[string]time.Time),
		done:    make(chan struct{}),
	}
	go rc.cleanup()
	return rc
}

// Check returns true if jti has been seen before (replay detected).
// If not seen, records the jti with the given TTL and returns false.
// Thread-safe: reads under RLock, writes under Lock.
func (rc *MemoryReplayCache) Check(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	// Fast path: read-only check
	rc.mu.RLock()
	expiry, exists := rc.entries[jti]
	rc.mu.RUnlock()

	if exists && time.Now().Before(expiry) {
		return true, nil // replay detected
	}

	// Slow path: write to record the jti
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Double-check after acquiring write lock
	if expiry, exists := rc.entries[jti]; exists && time.Now().Before(expiry) {
		return true, nil // replay detected
	}

	if ttl < time.Second {
		ttl = time.Second
	}

	rc.entries[jti] = time.Now().Add(ttl)
	return false, nil
}

// Close stops the background cleanup goroutine.
func (rc *MemoryReplayCache) Close() {
	close(rc.done)
}

func (rc *MemoryReplayCache) cleanup() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-rc.done:
			return
		case <-ticker.C:
			rc.mu.Lock()
			now := time.Now()
			for jti, expiry := range rc.entries {
				if now.After(expiry) {
					delete(rc.entries, jti)
				}
			}
			rc.mu.Unlock()
		}
	}
}

// Compile-time check that MemoryReplayCache implements core.ReplayCache.
var _ core.ReplayCache = (*MemoryReplayCache)(nil)
