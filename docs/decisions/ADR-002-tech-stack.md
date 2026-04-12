# ADR-002: Technology Stack Selection

## Status

Accepted

## Date

2026-04-12

## Context

OathMesh requires a technology stack that:
- Supports the protocol (JWS/JWT with EdDSA)
- Enables fast MVP iteration
- Has strong security properties (no CVE surface from crypto libs)
- Is appropriate for infrastructure/security tooling

## Decision

| Concern | Choice | Justification |
|---------|--------|----------------|
| Language | Go 1.22+ | stdlib crypto, fast binary, infra ecosystem |
| Config DSL | Apple Pkl | Typed policies, IDE validation, pkl-gen-go |
| HTTP framework | chi/v5 | Lightweight, idiomatic, stdlib-compatible |
| Signing | crypto/ed25519 (stdlib) | No CVE surface from third-party JWT libs |
| Logging | log/slog (stdlib) | Structured, zero-dep, cloud-native |
| CLI | spf13/cobra | Industry standard |
| Config env vars | spf13/viper | Layered env var loading |

### Why Go over Rust

- `crypto/ed25519` is in the Go standard library
- Single static binary without FFI complexity
- Go is the lingua franca of infra/security tooling (Vault, Consul, Caddy)
- Faster MVP iteration for a solo founder
- `log/slog` (Go 1.21+) gives structured logging out of the box

### Why Pkl over YAML/JSON

- Type-safe authoring with IDE completion
- Schema violations caught at write time, not at deny time
- `pkl-gen-go` generates native Go structs from Pkl schemas
- Policy files become a typed interface, not freeform YAML

## Consequences

- **Positive**: Zero external deps in `internal/core`
- **Positive**: Fast build times
- **Negative**: Pkl requires tooling (`pkl-gen-go`)

## Review

Approved by Architect Agent on 2026-04-12.
