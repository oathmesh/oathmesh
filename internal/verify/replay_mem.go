package verify

import (
	"context"
	"hash/fnv"
	"sync"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/metrics"
)

const (
	numShards       = 256
	maxShardItems   = 19531 // ~5,000,000 total items
	cleanupInterval = 60 * time.Second
)

type cacheShard struct {
	sync.RWMutex
	entries map[string]time.Time
}

// MemoryReplayCache implements core.ReplayCache for single-instance deployments.
// Uses a 256-way sharded lock table to eliminate global mutex contention during bursts,
// while defending against OOM by enforcing a maximum tracked cardinality (maxShardItems).
type MemoryReplayCache struct {
	shards [numShards]*cacheShard
	done   chan struct{}
}

func NewMemoryReplayCache() *MemoryReplayCache {
	rc := &MemoryReplayCache{
		done: make(chan struct{}),
	}
	for i := 0; i < numShards; i++ {
		rc.shards[i] = &cacheShard{
			entries: make(map[string]time.Time),
		}
	}
	go rc.cleanup()
	return rc
}

func (rc *MemoryReplayCache) getShard(jti string) *cacheShard {
	h := fnv.New32()
	h.Write([]byte(jti))
	return rc.shards[h.Sum32()%numShards]
}

func (rc *MemoryReplayCache) Check(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	shard := rc.getShard(jti)

	shard.RLock()
	expiry, exists := shard.entries[jti]
	shard.RUnlock()

	if exists && time.Now().Before(expiry) {
		return true, nil
	}

	shard.Lock()
	defer shard.Unlock()

	// Double check
	if expiry, exists := shard.entries[jti]; exists && time.Now().Before(expiry) {
		return true, nil
	}

	if !exists {
		if len(shard.entries) >= maxShardItems {
			rc.inlinePrune(shard)
		}
		metrics.ReplayCacheSize.Inc()
	}

	if ttl < time.Second {
		ttl = time.Second
	}
	shard.entries[jti] = time.Now().Add(ttl)
	return false, nil
}

func (rc *MemoryReplayCache) inlinePrune(shard *cacheShard) {
	now := time.Now()
	evicted := 0
	for jti, exp := range shard.entries {
		if now.After(exp) {
			delete(shard.entries, jti)
			evicted++
		}
	}
	
	if len(shard.entries) >= maxShardItems {
		forceEvict := 100
		for jti := range shard.entries {
			delete(shard.entries, jti)
			evicted++
			forceEvict--
			if forceEvict <= 0 {
				break
			}
		}
	}
	if evicted > 0 {
		metrics.ReplayCacheSize.Sub(float64(evicted))
	}
}

func (rc *MemoryReplayCache) Close() {
	close(rc.done)
}

func (rc *MemoryReplayCache) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rc.done:
			return
		case <-ticker.C:
			now := time.Now()
			for i := 0; i < numShards; i++ {
				shard := rc.shards[i]
				shard.Lock()
				evicted := 0
				for jti, expiry := range shard.entries {
					if now.After(expiry) {
						delete(shard.entries, jti)
						evicted++
					}
				}
				shard.Unlock()
				if evicted > 0 {
					metrics.ReplayCacheSize.Sub(float64(evicted))
				}
			}
		}
	}
}

var _ core.ReplayCache = (*MemoryReplayCache)(nil)
