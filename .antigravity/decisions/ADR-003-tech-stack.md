---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
---

# ADR-003: Technology Stack — Go Issuer, TypeScript Node SDK, Python SDK

**Status:** Accepted

## Context

OathMesh requires implementations across three roles:
1. **Issuer + Gateway + CLI** — server-side, performance-critical, single-binary deployment
2. **Node.js SDK** — middleware for Express/Koa/Hono receivers
3. **Python SDK** — middleware for FastAPI/Flask receivers

The language choice for each role has different constraints and trade-offs.

Reference: `oathmesh.txt` sections 6.1, 16.2

## Options Considered

### For Issuer / Gateway / CLI

#### Option A: Go

- Single static binary — no runtime dependencies
- Excellent crypto/ed25519 in standard library
- Built-in HTTP server, JSON handling, context propagation
- Container images under 15MB
- Mature ecosystem for JWT (lestrrat-go/jwx), YAML (gopkg.in/yaml.v3)
- Strong concurrency model
- Cons: more verbose than Node.js for rapid prototyping

#### Option B: Node.js (TypeScript)

- Fastest prototyping speed
- Shared ecosystem with the Node SDK
- Cons: runtime dependency (Node.js must be installed or bundled)
- Cons: larger container images
- Cons: weaker crypto performance for EdDSA than Go
- Cons: single-threaded — needs clustering for production

#### Option C: Rust

- Best performance, smallest binary
- Strong type safety
- Cons: slowest development speed
- Cons: steeper learning curve for contributors
- Cons: smaller JWT/JOSE ecosystem than Go

### For Node.js SDK

No choice needed — must be TypeScript/Node.js by definition.

### For Python SDK

No choice needed — must be Python by definition.

## Decision

We will use:

| Component | Language | Rationale |
|---|---|---|
| Issuer Service | **Go** | Single binary, stdlib crypto, sub-15MB images, excellent for infra services |
| Gateway | **Go** | Shares codebase with issuer, native reverse proxy support (net/http/httputil) |
| CLI | **Go** | Same binary as issuer; cobra for CLI framework |
| Node.js SDK | **TypeScript** (strict mode) | Target ecosystem; jose library for JWT |
| Python SDK | **Python 3.11+** | Target ecosystem; PyJWT + cryptography for JWT |

### Key Libraries

| Component | Core Libraries |
|---|---|
| Go Issuer | stdlib `crypto/ed25519`, `lestrrat-go/jwx/v2` for JWK/JWS, `gopkg.in/yaml.v3` for policy, `log/slog` for structured logging |
| Go CLI | `spf13/cobra` for CLI framework |
| Node SDK | `jose` for JWT/JWK, `zod` for validation, `yaml` for policy parsing |
| Python SDK | `PyJWT` + `cryptography` for JWT, `pydantic` for validation, `pyyaml` for policy, `httpx` for async JWKS fetch |

### Build and Distribution

| Component | Build | Distribution |
|---|---|---|
| Go Issuer/Gateway/CLI | `go build` → single static binary | Docker image (distroless base), GitHub Releases |
| Node SDK | `tsc` + bundler | npm as `@oathmesh/node` |
| Python SDK | `build` + `twine` | PyPI as `oathmesh` |

## Consequences

### Positive
- Go issuer produces tiny, dependency-free Docker images (meets < 30MB budget)
- Single binary simplifies deployment — no runtime to install
- TypeScript and Python SDKs are idiomatic for their target audiences
- All three languages have mature, well-audited JWT/EdDSA libraries

### Negative
- Two languages in the monorepo (Go + TypeScript + Python) increases toolchain complexity
- Contributors need Go knowledge for core protocol work
- Go's error handling verbosity may slow initial development

### Risks
- `lestrrat-go/jwx` maintenance risk: mitigated by its CNCF-adjacent status and large user base
- `jose` npm package maintenance risk: mitigated by being the most popular JOSE library for Node
- Cross-language consistency: mitigated by shared JSON Schemas in `/spec/` and integration tests

## References

- `oathmesh.txt` sections 6.1 (v1 scope), 16.2 (MVP scope)
- `rules/coding-standards.md` (language-specific conventions)
- `rules/core.md` § Dependency Rules
- ADR-001 (token format — informs library requirements)
