package verify

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/MustafaMahmoudAtta111/oathmesh/internal/core"
)

// ErrCacheUnavailable is returned when the replay cache backend is unreachable.
// This is distinct from ErrReplayDetected so operators can distinguish a Redis
// outage from an actual replay attack.
var ErrCacheUnavailable = errors.New("replay cache backend unavailable")

// RedisReplayCache implements core.ReplayCache using Redis.
// Production-grade: uses SET jti EX <remaining_ttl> NX for atomic check-and-set.
//
// NX ensures only the first call succeeds for a given jti. If SET returns false
// (key already exists), a replay is detected.
//
// Default behavior is fail-closed: if Redis is unreachable, the request is rejected.
// This follows the OathMesh security doctrine — fail-open would allow replays
// during an outage.
type RedisReplayCache struct {
	client     *redis.Client
	keyPrefix  string
	failClosed bool // true = reject on Redis error; false = allow (fail-open)
}

// RedisReplayCacheConfig holds configuration for the Redis replay cache.
type RedisReplayCacheConfig struct {
	// RedisURL is the Redis connection URL (e.g., "redis://localhost:6379/0").
	RedisURL string

	// KeyPrefix is prepended to jti values when storing in Redis.
	// Default: "oathmesh:replay:".
	KeyPrefix string

	// FailClosed controls behavior when Redis is unreachable.
	// true (default): reject the request — security-first.
	// false: allow the request — availability-first (NOT recommended).
	FailClosed bool
}

// NewRedisReplayCache creates a new Redis-backed replay cache.
func NewRedisReplayCache(cfg RedisReplayCacheConfig) (*RedisReplayCache, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	keyPrefix := cfg.KeyPrefix
	if keyPrefix == "" {
		keyPrefix = "oathmesh:replay:"
	}

	return &RedisReplayCache{
		client:     client,
		keyPrefix:  keyPrefix,
		failClosed: cfg.FailClosed,
	}, nil
}

// Check returns true if jti has been seen before (replay detected).
// Uses SET key EX ttl NX — atomic check-and-set:
//   - NX: only set if the key does NOT exist
//   - If SET succeeds (key was new): return false (not a replay)
//   - If SET fails (key exists): return true (replay detected)
//
// On Redis error with fail-closed (default): returns error wrapping ErrCacheUnavailable.
func (rc *RedisReplayCache) Check(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	key := rc.keyPrefix + jti

	if ttl < time.Second {
		ttl = time.Second
	}

	// SET key "1" EX ttl NX
	// Returns true if the key was set (new jti), false if it already existed (replay)
	result, err := rc.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		if rc.failClosed {
			return false, fmt.Errorf("%w: redis SET NX failed: %v", ErrCacheUnavailable, err)
		}
		// Fail-open: allow the request despite Redis error (NOT recommended)
		return false, nil
	}

	// SetNX returns true if the key was set (not a replay), false if it existed (replay)
	return !result, nil
}

// Ping checks connectivity to Redis.
func (rc *RedisReplayCache) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// Close closes the Redis client connection.
func (rc *RedisReplayCache) Close() error {
	return rc.client.Close()
}

// Compile-time check that RedisReplayCache implements core.ReplayCache.
var _ core.ReplayCache = (*RedisReplayCache)(nil)
