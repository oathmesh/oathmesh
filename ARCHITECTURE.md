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
  sign/                 → Ed25519 signing, JWS construction
  verify/               → 14-step verification pipeline
  policy/               → Pkl policy engine + hot-reload
  audit/                → NDJSON audit event pipeline
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
cmd/oathmesh ──▶ internal/issuer ──▶ internal/sign
                                  ──▶ internal/verify ──▶ internal/core
                                  ──▶ internal/audit   ──▶ internal/core
                                  ──▶ internal/config
                 internal/gateway ──▶ internal/verify
                                  ──▶ internal/core
                 internal/policy  ──▶ internal/core
                 sdk/go/middleware ──▶ internal/verify
                                  ──▶ internal/core
```

**Invariant:** `internal/core` has zero external dependencies. It defines protocol types only. If it imported anything external, every downstream package would transitively depend on it.

## Data Flow

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
  │  {token: "<om+jwt>"}   │
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
  │                         │  Step 05: Load JWKS (cached)
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
| Language | Go 1.22+ | Performance, stdlib crypto, single-binary deployment |
| HTTP framework | chi/v5 | Lightweight, stdlib-compatible, middleware-chain pattern |
| Signing | crypto/ed25519 (stdlib) | No third-party dependency on critical security path |
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

- **Issuer:** Primarily stateless for minting via shared keys, but acts as a control plane proxy bridging administrative revocation interactions dynamically securely to Redis.
- **Verifier:** Stateless, scales with request volume
- **Replay cache:** 
  - Dev: In-memory (single instance)
  - Prod: Redis (multi-instance, shared)
- **Policy:** Hot-reload from filesystem, no external dependency

## Decision Matrix

| Decision | Choice | Alternative Considered |
|----------|--------|----------------------|
| Signing algorithm | Ed25519 | RSA, ECDH |
| Token format | JWS (compact) | JWT, custom binary |
| TTL max | 300s | 1h, 24h |
| Policy DSL | Pkl | Rego (OPA), CEL |
| Audit format | NDJSON | JSON Lines, syslog |
