package issuer

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*clientState
	rate    int
	burst   int
	window  time.Duration
}

type clientState struct {
	tokens    int
	lastReset time.Time
}

func NewRateLimiter(rate, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientState),
		rate:    rate,
		burst:   burst,
		window:  time.Minute,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[ip]

	if !exists || now.Sub(client.lastReset) >= rl.window {
		rl.clients[ip] = &clientState{
			tokens:    rl.burst,
			lastReset: now,
		}
		return true
	}

	if client.tokens > 0 {
		client.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, client := range rl.clients {
			if now.Sub(client.lastReset) > 10*rl.window {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}
