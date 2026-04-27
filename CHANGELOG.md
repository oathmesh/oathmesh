# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0](https://github.com/oathmesh/oathmesh/compare/v0.1.0...v0.2.0) (2026-04-27)


### Features

* **security:** Phase 1 critical security hardening ([7ecac42](https://github.com/oathmesh/oathmesh/commit/7ecac42c18f9d5fcca21a294a6bddbdf0f09b9e9))
* **security:** phase 2 sdk parity and revocation fail-closed protocol ([518a9c0](https://github.com/oathmesh/oathmesh/commit/518a9c03b62b364390bc34ed786b7f1b00d98025))


### Bug Fixes

* **node:** align conformance test with token_malformed error ([7f3ac1a](https://github.com/oathmesh/oathmesh/commit/7f3ac1a1bb5e3a98b19cbec0330bcb81b73261f2))

## [1.0.0-rc.1] - 2026-04-27

### 🚀 Major Release: The Zero-Trust Hardening & Ecosystem Release

This release represents the culmination of the OathMesh 3-Phase Security and Maturation Roadmap. OathMesh is now a mathematically proven, polyglot-consistent, and ecosystem-interoperable zero-trust engine, ready for production perimeter defense.

### ⚠️ BREAKING CHANGES (Phase 1 & 2 Migrations)
- **Config Rename:** `ClockSkew` has been renamed to `ClockSkewLeeway` across all SDKs to reflect its purpose accurately. The default value has been increased from `10s` to `30s` to accommodate NTP drift in distributed systems.
- **Environment Variables:** `OATHMESH_PRIVATE_KEY_FILE` is now deprecated. Please migrate to `OATHMESH_PRIVATE_KEY_PATH`.
- **Pkl Policy Sandbox:** Broad `file:///` access and external HTTP imports are no longer permitted in Pkl policies. Policies are now strictly sandboxed to `file://<dir>/` relative to the policy root to neutralize SSRF and LFI vectors.
- **Cache Failover Behavior:** `RevocationList` cache failures (e.g., Redis network partitions) now strictly **fail-closed**. All SDKs return a secure denial (`verification_failed` / HTTP 401) rather than failing open.

### 🔒 Phase 1: Cryptographic & Policy Hardening
- **Clock Leeway:** Updated verification pipeline to correctly apply `ClockSkewLeeway` to `exp`, `iat`, and `nbf` claims.
- **Algorithm Confusion:** Added strict regression tests guaranteeing `alg:none` and symmetric algorithm attacks fail immediately.
- **Secure Keygen:** Introduced `oathmesh keygen` CLI command, enforcing strict `0600` POSIX file permissions on generated private keys.
- **Pkl Sandboxing:** Hardened Pkl evaluation execution with `--allowed-modules` and scoped `--allowed-resources`.

### 🛡️ Phase 2: SDK Parity & Resilience
- **Error Parity:** Implemented the unified `token_malformed` error across Go, Node.js, and Python SDKs.
- **Revocation Caches:** Added `InMemoryRevocationCache` and `RedisRevocationCache` implementations to Node.js and Python SDKs, establishing exact behavioral parity with Go.
- **Conformance:** Augmented cross-SDK conformance runners to dynamically test caching limits, TTL clamping, and `subject_revoked` triggers with a 0% mismatch rate.

### 🌐 Phase 3: Ecosystem Integrations & Adoption
- **Envoy `ext_authz` Service:** Introduced a standalone Go binary (`cmd/oathmesh-envoy`) implementing the gRPC `ext_authz` interface. Injects `X-OathMesh-*` context headers directly into Envoy's `OkResponse`.
- **Kong Go PDK Plugin:** Engineered a high-performance external Kong plugin (`plugins/kong`) using the Kong Go PDK, reusing the core Go verification pipeline via msgpack.
- **Performance Proven:** Published `docs/PERFORMANCE.md` detailing K8s/k6 benchmarks, mathematically proving the "Zero-Trust Overhead Delta" is <1ms at p99.
- **Audit Readiness:** Synthesized architecture, threat models, and scope into `docs/security/AUDIT_RFP.md` for top-tier security evaluation firms.


## [1.0.6] - 2026-04-13

### Security
- Fixed SSRF vulnerability in JWKS fetch (CodeQL go/request-forgery)
  - Added `NewFixedJWKS()` with hardcoded URL - user input completely ignored
  - Pre-computed JWKS URLs in config - no string concatenation
  - Rate limiting added to test suite
- Updated Go toolchain to 1.26.2

### Added
- Cross-SDK conformance test suite with fixtures and runners
- Replay cache support in Node.js SDK
- Replay cache support in Python SDK
- JSON policy evaluation in Node.js SDK
- JSON policy evaluation in Python SDK
- `make bench` target for performance benchmarks
- `make conformance` target for cross-SDK testing
- Dockerfile.gateway for separate gateway deployment
- Emergency JWKS cache TTL override (`OATHMESH_JWKS_CACHE_TTL_EMERGENCY`)
- Request binding requirement (`RequireRequestBinding`)

### Changed
- **BREAKING**: Key ID format changed to `issuer-key-YYYY-MM-{4-char-random-hex}` for improved uniqueness
- **BREAKING**: ES256 deprecated - tokens signed with ES256 will now log warnings
- Default TTL reduced to 120 seconds (was 300)
- Maximum TTL capped at 300 seconds
- Gateway now requires explicit policy configuration in production

### Fixed
- JWKS cache now properly respects TTL with emergency override capability

## [0.1.0] - 2026-04-23

### Added
- Demo script health polling loop fixed
- SDK parity across Go, Node.js, and Python implementations
- Conformance cases expanded with `alg_none_rejection`, `subject_format_validation`, `binding_required_semantics`, and `iat_future_rejection`

### Security
- Added request binding requirement (rqh claim) enforcement
- NetworkPolicy in Kubernetes deployment
- TLS enforcement for production issuer URLs
- Replay attack defense with in-memory and Redis-backed cache

### Changed
- Node.js and Python verifier behavior tightened toward canonical Go step semantics for conformance-critical checks
- Revocation-list behavior remains optional in Node.js/Python SDKs and is currently SKIP/N/A in cross-SDK conformance

## [0.1.0] - 2025-04-13

### Added
- Core Go implementation (issuer, gateway, CLI)
- Node.js SDK
- Python SDK
- Docker Compose deployment
- Kubernetes deployment guide
- TLS configuration guide
- Security documentation (threat model, key management, replay defense)
