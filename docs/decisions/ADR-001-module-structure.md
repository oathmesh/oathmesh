# ADR-001: Module and Package Structure

## Status

Accepted

## Date

2026-04-12

## Context

OathMesh needs a Go module structure that:
- Has zero external dependencies in the core protocol types
- Supports multiple SDKs (Go, Python, Node.js) from a single repo
- Follows Go best practices for internal packages

## Decision

The module root is `github.com/oathmesh/oathmesh` with the following package structure:

```
oathmesh/
├── cmd/oathmesh/           # CLI entry point (cobra)
├── internal/
│   ├── core/               # Protocol types (zero external deps)
│   │   ├── errors.go       # OathMeshError type
│   │   ├── audit.go        # AuditEvent + AuditSink
│   │   ├── context.go      # VerifiedCallerContext
│   │   ├── replay.go       # ReplayCache interface
│   │   └── token.go        # OathToken, Header, Claims
│   ├── sign/               # Ed25519 signing (Security Agent owns)
│   ├── verify/             # 14-step verification (Security Agent owns)
│   ├── policy/             # Pkl-backed policy engine
│   ├── audit/              # Audit pipeline implementations
│   ├── issuer/             # HTTP issuer service (chi)
│   ├── gateway/            # Reverse proxy mode
│   └── config/             # Config loading (Pkl + env)
├── sdk/
│   ├── go/                 # Go middleware
│   ├── python/             # Python SDK
│   └── node/               # Node.js SDK
├── policy/                 # Pkl policy files
├── examples/               # Working examples
└── docs/                   # Documentation
```

### Package Dependencies

```
cmd/oathmesh → internal/* → internal/core
sdk/go → internal/verify, internal/core
internal/issuer → internal/sign, internal/core
internal/gateway → internal/verify, internal/policy, internal/audit
```

### Constraints

- `internal/core` has zero external imports (verified by `go list -deps`)
- SDK packages are sub-packages of the same Go module (no separate modules)
- No `go.work` workspace file

## Consequences

- **Positive**: Clean separation of concerns, Security Agent owns crypto code
- **Positive**: Single module simplifies dependency management
- **Negative**: Python/Node SDKs require separate package.json/pyproject.toml

## Review

Architect Agent approved package structure and dependency graph on 2026-04-12.
