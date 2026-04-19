package middleware

import (
	"context"
	"sync"
	"time"
)

// GRPCMiddlewareConfig holds configuration for gRPC interceptors.
type GRPCMiddlewareConfig struct {
	// RateLimitPerMinute is the number of tokens allowed per minute per subject.
	// Default: 1000
	RateLimitPerMinute int

	// TokenExtractor is used to extract tokens from gRPC metadata.
	// If nil, DefaultTokenExtractor is used.
	TokenExtractor TokenExtractor
}

// TokenExtractor defines how to extract a token from gRPC metadata context.
type TokenExtractor interface {
	ExtractToken(ctx context.Context) (string, error)
}

// RateLimiter enforces rate limits on a per-subject basis.
type RateLimiter interface {
	// Allow returns true if the subject is within the rate limit.
	// Returns false and an error message if rate limit is exceeded.
	Allow(subject string) (bool, string)
}

// SimpleRateLimiter is a thread-safe in-memory rate limiter.
type SimpleRateLimiter struct {
	mu              sync.RWMutex
	limitPerMinute  int
	requestTimes    map[string][]time.Time
	cleanupInterval time.Duration
}

// NewSimpleRateLimiter creates a new in-memory rate limiter.
func NewSimpleRateLimiter(limitPerMinute int) *SimpleRateLimiter {
	rl := &SimpleRateLimiter{
		limitPerMinute:  limitPerMinute,
		requestTimes:    make(map[string][]time.Time),
		cleanupInterval: 1 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a subject is within rate limits.
func (rl *SimpleRateLimiter) Allow(subject string) (bool, string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	oneMinuteAgo := now.Add(-1 * time.Minute)

	// Clean old entries for this subject
	times := rl.requestTimes[subject]
	var validTimes []time.Time
	for _, t := range times {
		if t.After(oneMinuteAgo) {
			validTimes = append(validTimes, t)
		}
	}

	// Check limit
	if len(validTimes) >= rl.limitPerMinute {
		return false, "rate limit exceeded"
	}

	// Add current request
	validTimes = append(validTimes, now)
	rl.requestTimes[subject] = validTimes

	return true, ""
}

// cleanup periodically removes old entries to prevent unbounded memory growth.
func (rl *SimpleRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		tenMinutesAgo := now.Add(-10 * time.Minute)

		for subject, times := range rl.requestTimes {
			var validTimes []time.Time
			for _, t := range times {
				if t.After(tenMinutesAgo) {
					validTimes = append(validTimes, t)
				}
			}

			if len(validTimes) == 0 {
				delete(rl.requestTimes, subject)
			} else {
				rl.requestTimes[subject] = validTimes
			}
		}
		rl.mu.Unlock()
	}
}
