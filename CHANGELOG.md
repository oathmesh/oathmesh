# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-04-27

### ⚠️ BREAKING CHANGES
- **Config Rename:** `ClockSkew` has been renamed to `ClockSkewLeeway` across Go, Node.js, and Python SDKs to accurately reflect its purpose. The default value has been increased from `10s` to `30s` to better accommodate NTP drift in distributed systems.
- **Environment Variables:** `OATHMESH_PRIVATE_KEY_FILE` is now deprecated. Please migrate to `OATHMESH_PRIVATE_KEY_PATH`. The old variable will emit a stern warning but will continue to function temporarily.

### 🔒 Security Hardening (Phase 1)
This release concludes Phase 1 of the OathMesh security roadmap, focusing on critical attack vector mitigation and audit preparation.

- **Cryptographic & Token Fixes:**
  - Updated verification pipeline to correctly apply `ClockSkewLeeway` to `exp`, `iat`, and `nbf` claims.
  - Added strict regression tests to guarantee `alg:none` and symmetric algorithm confusion attacks fail immediately.
  
- **Key Management:**
  - Introduced the `oathmesh keygen` CLI command to generate Ed25519 key pairs.
  - `oathmesh keygen` now enforces strict `0600` POSIX file permissions on generated private keys, preventing accidental host-level credential exposure.
  - Added deprecation warnings for plaintext key environment variables.

- **Pkl Policy Engine Sandboxing:**
  - Hardened the Pkl evaluation execution to neutralize SSRF vectors and arbitrary remote includes.
  - Applied strict `--allowed-modules="pkl:*"` and scoped `--allowed-resources="env:*,prop:*,file://<dir>/"` flags. Broad `file:///` access is no longer permitted.
  - Added regression tests ensuring external HTTP imports and local file escape (`/etc/...`) attempts fail securely.
  - Introduced a baseline `policy/policy.pkl.schema` for policy validation.

### 📚 Documentation
- Added `docs/security/AUDIT_SCOPE.md` defining boundaries for independent security reviews.
- Added `docs/security/SECURITY_ASSESSMENT.md` containing the evaluation checklist and remediation process.

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
