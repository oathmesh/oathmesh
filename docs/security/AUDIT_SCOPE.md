# OathMesh Security Audit Scope

## In Scope
- 14-step verification pipeline (`internal/verify/verify.go` and `internal/verify/steps.go`)
- Ed25519 signing/verification (`internal/crypto/ed25519.go` and `internal/sign/`)
- Token parsing and claim validation (`internal/token/parser.go` and `internal/sign/`)
- Redis replay cache implementation (`internal/verify/replay_redis.go`)
- Stateful revocation implementation (`internal/verify/revocation_redis.go`)
- Pkl policy sandboxing (`internal/policy/loader.go`)

## Out of Scope
- SDK bindings (Go, Node.js, Python)
- Gateway proxy mode
- CLI management commands (except `keygen` security posture)
