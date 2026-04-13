# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.6] - 2026-04-13

### Security
- Fixed SSRF vulnerability in JWKS fetch (CodeQL go/request-forgery)
  - Added `NewFixedJWKS()` with hardcoded URL - user input completely ignored
  - Pre-computed JWKS URLs in config - no string concatenation
  - Rate limiting added to test suite
- Updated Go toolchain to 1.26

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

## [Unreleased]

### Added
- Demo script health polling loop fixed
- SDK parity across Go, Node.js, and Python implementations

### Security
- Added request binding requirement (rqh claim) enforcement
- NetworkPolicy in Kubernetes deployment
- TLS enforcement for production issuer URLs
- Replay attack defense with in-memory and Redis-backed cache

## [0.1.0] - 2025-04-13

### Added
- Core Go implementation (issuer, gateway, CLI)
- Node.js SDK
- Python SDK
- Docker Compose deployment
- Kubernetes deployment guide
- TLS configuration guide
- Security documentation (threat model, key management, replay defense)