package verify

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRevocationList implements core.RevocationList using a Redis hash map.
type RedisRevocationList struct {
	client *redis.Client
	key    string
}

// NewRedisRevocationList creates a new Redis-backed revocation list.
func NewRedisRevocationList(redisURL string) (*RedisRevocationList, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	return &RedisRevocationList{
		client: client,
		key:    "oathmesh:revocations",
	}, nil
}

// IsRevoked checks if a subject has been revoked.
// OathMesh revocation policy: if a subject is revoked, ALL tokens for that subject
// are invalid regardless of when they were issued (iat). Revocation invalidates
// all existing tokens and prevents any future tokens for that subject.
func (rl *RedisRevocationList) IsRevoked(ctx context.Context, subject string) (bool, error) {
	// Fast path check using HGET
	val, err := rl.client.HGet(ctx, rl.key, subject).Result()
	if err == redis.Nil {
		return false, nil // Not revoked
	}
	if err != nil {
		return false, fmt.Errorf("redis check failed: %w", err)
	}

	// Subject exists in revocation list
	_ = val // Parse result but don't use it; subject existence is what matters
	return true, nil
}

// Revoke adds a subject to the revocation list.
func (rl *RedisRevocationList) Revoke(ctx context.Context, subject string) error {
	return rl.client.HSet(ctx, rl.key, subject, time.Now().Format(time.RFC3339)).Err()
}

// Unrevoke removes a subject from the revocation list.
func (rl *RedisRevocationList) Unrevoke(ctx context.Context, subject string) error {
	return rl.client.HDel(ctx, rl.key, subject).Err()
}

// List returns all revoked subjects and their revocation times.
func (rl *RedisRevocationList) List(ctx context.Context) ([]Revocation, error) {
	data, err := rl.client.HGetAll(ctx, rl.key).Result()
	if err != nil {
		return nil, err
	}

	var out []Revocation
	for sub, val := range data {
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			out = append(out, Revocation{
				Subject:   sub,
				RevokedAt: t,
			})
		}
	}
	return out, nil
}

// Close closes the Redis client connection.
func (rl *RedisRevocationList) Close() error {
	return rl.client.Close()
}

var _ RevocationList = (*RedisRevocationList)(nil)
