package verify

import (
	"context"
	"fmt"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
)

// RedisReplayCache implements core.ReplayCache using Redis.
// Production-grade: uses SET jti EX <remaining_ttl> NX for atomic check-and-set.
//
// TODO(phase3): Implement with github.com/redis/go-redis/v9.
// This is a placeholder that satisfies the interface for Phase 2.
type RedisReplayCache struct{}

func NewRedisReplayCache(redisURL string) (*RedisReplayCache, error) {
	return nil, fmt.Errorf("redis replay cache not yet implemented; use MemoryReplayCache for dev")
}

func (rc *RedisReplayCache) Check(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	return false, fmt.Errorf("redis replay cache not yet implemented")
}

// Compile-time check that RedisReplayCache implements core.ReplayCache.
var _ core.ReplayCache = (*RedisReplayCache)(nil)
