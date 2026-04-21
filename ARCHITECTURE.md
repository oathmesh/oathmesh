# Architecture

> 📐 Technical design, system components, and data flows for OathMesh.

## System Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                        OathMesh System                              │
│                                                                     │
│  ┌──────────────┐     ┌──────────────┐     ┌────────────────────┐ │
│  │   Caller      │     │   Issuer      │     │    Receiver        │ │
│  │               │     │   Service     │     │                    │ │
│  │  agent://     │────▶│              │     │  ┌──────────────┐  │ │
│  │  svc://       │     │  POST /v1/   │     │  │  Verifier    │  │ │
│  │  job://       │◀────│   token      │     │  │  (14 steps)  │  │ │
│  │  tool://      │     │              │     │  └──────┬───────┘  │ │
│  └──────────────┘     │  GET /.well- │     │  ┌──────▼───────┐  │ │
│        │               │   known/jwks │     │  │  Policy      │  │ │
│        │               └──────────────┘     │  │  Engine      │  │ │
│        │                      │              │  └──────┬───────┘  │ │
│        │               ┌──────▼───────┐     │  ┌──────▼───────┐  │ │
│        └──────────────▶│   Gateway    │────▶│  │  Audit       │  │ │
│                        │   (optional) │     │  │  Pipeline    │  │ │
│                        └──────────────┘     │  └──────────────┘  │ │
│                                              └────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
cmd/oathmesh/           → CLI entry point (cobra)
internal/
  core/                 → Protocol types, zero external deps
  jwk/                  → JWK/JWKS types and operations (zero external deps)
  sign/                 → Ed25519 signing, JWS construction (delegates JWK types to internal/jwk)
  verify/               → 14-step verification pipeline (func-slice architecture)
    steps.go            → Individual step functions + pipeline definition
    verify.go           → Pipeline orchestrator (~40 lines)
    jwks.go             → JWKS cache with pre-warm support
    replay_mem.go       → In-memory sharded replay cache (batch cleanup, 16 shards/tick)
    replay_redis.go     → Redis-backed replay cache
    replay_circuitbreaker.go → Circuit-breaker: Redis → MemoryReplayCache failover
  policy/               → Pkl policy engine + hot-reload
  audit/                → NDJSON audit event pipeline + FanOutAuditSink
  metrics/              → Prometheus metrics (incl. oathmesh_clock_skew_rejections_total)
  issuer/               → HTTP issuer service (chi)
  gateway/              → Reverse proxy mode
  config/               → Configuration schema + env loading
sdk/
  go/middleware/        → Go chi middleware
  node/                 → Node.js/TypeScript Express middleware
  python/               → Python FastAPI/Flask middleware
```

## Dependency Graph

```
cmd/oathmesh ──▶ internal/issuer ──▶ internal/sign ──▶ internal/jwk
                                  ──▶ internal/verify ──▶ internal/core
                                  ──▶ internal/audit   ──▶ internal/core
                                  ──▶ internal/config
                 internal/gateway ──▶ internal/verify
                                  ──▶ internal/core
                 internal/policy  ──▶ internal/core
                 internal/sign    ──▶ internal/jwk
                 sdk/go/middleware ──▶ internal/verify
                                  ──▶ internal/core
```

**Invariants:**
- `internal/core` has zero external dependencies. It defines protocol types only.
- `internal/jwk` has zero external dependencies. JWK/JWKS types are stdlib-only.
- `internal/sign` re-exports `jwk.JWK`/`jwk.JWKS` as type aliases for backward compatibility.

## Data Flow

### Token Lifecycle Flow

```text
+--------+     1) POST /v1/token      +--------+     3) Authorization: OathMesh <token>     +----------+
| Caller |--------------------------->| Issuer |---------------------------------------------->| Receiver |
+--------+   {sub,aud,act,ttl hint}   +--------+                                               +----------+
    ^                                      |                                                         |
    | 2) { token, expires_in, token_type } | Validate + clamp TTL (1-300s)                         |
    +--------------------------------------+ Generate jti + sign JWS                                 |
                                                                                                      |
                                                                                                      v
                                                                                           4) Verify 14 steps
                                                                                                      |
                                                                                       +--------------+---------------+
                                                                                       |                              |
                                                                                       v                              v
                                                                              5a) allow + audit.allow       5b) deny + audit.deny
                                                                                  (request continues)         (401 + error taxonomy)
```

### Verification Pipeline Overview (14 Steps)

The pipeline is implemented as a **func slice** (`[]pipelineStep` in `steps.go`).
Each step is an individually testable function operating on a shared `vctx` struct.
The orchestrator in `verify.go` is a ~40-line loop that stops on first error.

Every error includes a `Step` field (1-14) so operators can immediately identify
which pipeline stage rejected a token.

```text
START: Authorization header present and token extracted
  |
  +--> [01] Parse 3 segments
  +--> [02] Validate header (typ/alg, reject alg=none)
  +--> [03] Decode payload, read iss
  +--> [04] Check trusted issuer list
  +--> [11] Check required claims (moved before JWKS to fail fast)
  +--> [11.5] Verify subject format (svc://, agent://, etc.)
  +--> [05] Load JWKS (cache; mapped endpoint or {iss}/.well-known/jwks.json)
  +--> [06] Verify signature + alg confusion check
  +--> [07] Re-check iss after signature
  +--> [08] Check exp (10s skew) — emits oathmesh_clock_skew_rejections_total
  +--> [09] Check iat/nbf (not future, 10s skew) — emits oathmesh_clock_skew_rejections_total
  +--> [10] Check aud (exact match)
  +--> [12] Verify rqh binding (if present)
  +--> [13] Check replay cache (jti) + revocation list
  +--> [14] Evaluate policy (first match wins, default deny)
            |
            +--> ALLOW => 200 path + audit.allow
            +--> DENY  => 401 path + audit.deny
```

### Policy Evaluation Flow

```text
Verification context ready
  |
  v
Load policy rules (ordered)
  |
  v
+-------------------------------+
| Rule 1 match?                 |-- yes --> apply effect (allow/deny) --> emit audit
+-------------------------------+
  |
  no
  v
+-------------------------------+
| Rule 2 match?                 |-- yes --> apply effect (allow/deny) --> emit audit
+-------------------------------+
  |
  no
  v
... continue through Rule N ...
  |
  v
No rule matched
  |
  v
Default deny -> 401 + audit.deny
```

### Token Minting

```
Caller                    Issuer
  │                         │
  │  POST /v1/token         │
  │  Authorization: Bearer  │
  │  {sub, aud, act, ttl}   │
  │────────────────────────▶│
  │                         │ 1. Validate request
  │                         │ 2. Clamp TTL to max(1, min(hint, 300))
  │                         │ 3. Set iat = now, exp = iat + clamped_ttl
  │                         │ 4. Generate jti (uuid.New())
  │                         │ 5. Sign with Ed25519 private key
  │  {token: "<om+jwt>", expires_in: 120, token_type: "OathMesh"} │
  │◀────────────────────────│
```

### Token Verification

```
Caller                 Receiver (or Gateway)
  │                         │
  │  Authorization:         │
  │  OathMesh <token>       │
  │────────────────────────▶│
  │                         │  Step 01: Parse structure (3 segments)
  │                         │  Step 02: Validate header (typ, alg)
  │                         │  Step 03: Decode payload, extract iss
  │                         │  Step 04: Check iss against trusted list
  │                         │  Step 05: Load JWKS (mapped endpoint or issuer-derived URL, cached)
  │                         │  Step 06: Verify JWS signature
  │                         │  Step 07: Re-verify iss after sig check
  │                         │  Step 08: Check exp (+ 10s skew)
  │                         │  Step 09: Check iat (not future)
  │                         │  Step 10: Check aud (exact match)
  │                         │  Step 11: Check required claims
  │                         │  Step 12: Check rqh binding (if present)
  │                         │  Step 13: Check replay cache
  │                         │  Step 14: Evaluate policy → audit event
  │                         │
  │  200 + caller context   │  (if allowed)
  │◀────────────────────────│
  │  401 + error taxonomy   │  (if denied)
  │◀────────────────────────│
```

### Gateway Proxy Flow

```
Caller          Gateway              Upstream
  │                │                     │
  │  OathMesh      │                     │
  │  <token>       │                     │
  │───────────────▶│                     │
  │                │ Verify (14 steps)   │
  │                │ Policy evaluate     │
  │                │ Audit emit          │
  │                │                     │
  │                │ Strip Authorization │
  │                │ Inject X-OathMesh-* │
  │                │ headers             │
  │                │────────────────────▶│
  │                │                     │ (no raw token)
  │                │◀────────────────────│
  │◀───────────────│                     │
```

## Technology Stack

| Concern | Choice | Rationale |
|---|---|---|
| Language | Go 1.26.2+ | Performance, stdlib crypto, single-binary deployment |
| HTTP framework | chi/v5 | Lightweight, stdlib-compatible, middleware-chain pattern |
| Signing | `crypto/ed25519` or `AWS KMS` | Secure hardware isolation mapped natively out-of-band |
| Config DSL | Apple Pkl | Schema validation at load time, type safety |
| Audit | NDJSON to stdout/file | Cloud-native, parseable, appendable |
| Replay cache | In-memory / Redis | Memory for dev, Redis for production multi-instance |

## Performance Characteristics

> ⚡ Benchmarks coming soon. These are preliminary estimates.
> Run `make bench` to generate real numbers for your hardware.

| Operation | Latency (p99) | Notes |
|-----------|---------------|-------|
| Token mint | < 5ms | Ed25519 sign, in-memory key |
| Token verify | < 2ms | 14-step pipeline, cached JWKS |
| JWKS cache hit | < 0.1ms | In-memory LRU cache |
| Policy eval | < 1ms | Pkl hot-reload, compiled rules |

### Scaling Considerations

- **Issuer:** Primarily stateless for token minting, with optional Redis-backed revocation coordination.
- **Verifier:** Stateless, scales with request volume
- **Replay cache:** 
  - Dev: In-memory (256-shard, batch cleanup 16 shards/tick)
  - Prod: Redis via `CircuitBreakerReplayCache` — automatic failover to in-process `MemoryReplayCache` during Redis outages (never fails open)
- **JWKS cache:** Pre-warmable at startup via `JWKSCache.PreWarm(ctx)` to eliminate cold-start latency
- **Audit:** Composable via `FanOutAuditSink` (dispatches to N sinks: file, stdout, metrics, tracing)
- **Policy:** Hot-reload from filesystem, no external dependency

## Decision Matrix

| Decision | Choice | Alternative Considered |
|----------|--------|----------------------|
| Signing algorithm | Ed25519 | RSA, ECDH |
| Token format | JWS (compact) | JWT, custom binary |
| TTL max | 300s | 1h, 24h |
| Policy DSL | Pkl | Rego (OPA), CEL |
| Audit format | NDJSON | JSON Lines, syslog |
