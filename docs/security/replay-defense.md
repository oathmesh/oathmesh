# Replay Defense

## The Problem

Without replay protection, an intercepted Oath Token could be reused within its TTL window to make unauthorized requests. Even with short TTLs (≤300s), a 5-minute replay window is enough for an attacker to cause damage.

## Defense Layers

### Layer 1: Unique Token IDs (`jti`)

Every Oath Token contains a `jti` (JWT ID) claim—a UUID generated via `uuid.New()`. The UUID is cryptographically random, never sequential, never predictable.

### Layer 2: Replay Cache (Step 13)

At verification time, the receiver checks whether the `jti` has been seen before:

1. If `jti` is in the cache → **reject** with `replay_detected`
2. If `jti` is not in the cache → **record** the `jti` with TTL = remaining token lifetime, then proceed

The cache entry automatically expires when the token would have expired, preventing unbounded memory growth.

### Implementation Options

**MemoryReplayCache** (development / single-instance):
- In-process `map[string]time.Time` protected by `sync.RWMutex`
- Reads under `RLock`, writes under `Lock`
- Periodic cleanup goroutine removes expired entries
- Not shared across instances — use Redis for multi-instance deployments

**RedisReplayCache** (production):
- `SET jti EX <remaining_ttl> NX` — atomic check-and-set
- `NX` ensures the key is only set if it doesn't exist (race-free)
- `EX` sets automatic expiry matching the token's remaining lifetime
- Shared across all instances — consistent replay detection in distributed deployments

### Fail Behavior

| Scenario | Behavior |
|---|---|
| `MemoryReplayCache` unavailable | N/A (in-process, always available) |
| Redis unavailable | **Fail closed** (default). Returns `ErrCacheUnavailable`, which blocks the request. Configurable. |
| Redis returns error | Treated as cache unavailable. Fail-closed. |

### Layer 3: Request Hash Binding (`rqh`)

For high-security scenarios, callers can bind a token to a specific request body:

1. Compute `sha256(canonical_request_body)`
2. Include the hash as the `rqh` claim: `"rqh": "sha256:e3b0c44298fc1c149afb"`
3. At verification (Step 12), the receiver computes the same hash and compares

This prevents a replayed token from being used with a different request body.

**When to use `rqh`:**
- Write/mutate operations where replay with a different body would be dangerous
- Financial transactions
- State-changing operations in critical systems

**When `rqh` is not needed:**
- Read-only operations
- Idempotent operations where replay is harmless

## TTL as Defense

Short TTLs are themselves a replay defense:

| TTL | Replay Window |
|---|---|
| 60s (write) | Attacker has <60s to intercept and replay |
| 120s (default) | Attacker has <120s |
| 300s (maximum) | Attacker has <300s |

Combined with TLS (preventing interception) and the replay cache (preventing reuse), the effective replay risk is near zero.
