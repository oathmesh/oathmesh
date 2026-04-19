# Copilot instructions for OathMesh

## Build, test, and lint commands
Run from repository root unless stated otherwise.

### Go (CLI + core services)
- Build CLI: `make build` or `CGO_ENABLED=0 go build -o bin/oathmesh ./cmd/oathmesh`
- Test (all Go packages): `make test` or `go test -v ./...`
- Test with race detector: `make race` or `go test -race ./...`
- Lint: `make lint` (runs `golangci-lint run ./...` when installed)
- Vet (CI): `go vet ./...`
- Local quality gate from README: `make quality-local`
- Single Go test examples:
  - `go test ./internal/verify -run '^TestConformance_issuer_check_untrusted$' -count=1`
  - `go test ./sdk/go/middleware -run '^TestConformance_middleware_auth_header_handling_semantics$' -count=1`

### Node SDK (`sdk/node`)
- Install deps: `cd sdk/node && npm install`
- Build: `cd sdk/node && npx tsc`
- Test suite: `cd sdk/node && npx vitest run`
- Single test file: `cd sdk/node && npx vitest run test/conformance.test.ts`
- Single test case: `cd sdk/node && npx vitest run test/conformance.test.ts -t "audience_check_mismatch"`

### Python SDK (`sdk/python`)
- Install test deps: `cd sdk/python && pip install -e .[test]`
- Test suite: `cd sdk/python && pytest tests/`
- Single test file: `cd sdk/python && python -m pytest tests/test_conformance.py -q`
- Single test case: `cd sdk/python && python -m pytest tests/test_conformance.py -k "audience_check_mismatch" -q`

### Cross-SDK conformance
- Run harness: `make conformance`
- Canonical per-target single-test commands are tracked in `tests/conformance/cases.json`.

## High-level architecture
- `cmd/oathmesh` is the single entrypoint. Its `serve` command runs issuer mode and can enable gateway mode (`--gateway`) to protect an upstream behind verification.
- `internal/issuer` exposes health/metrics/JWKS/discovery plus mint and OIDC exchange endpoints. Mint/admin endpoints are wrapped by `MintAuth`.
- `internal/sign` mints compact JWS Oath Tokens (`typ=om+jwt`, `alg=EdDSA`) with `jti` generated via UUID and TTL clamped to max 300 seconds.
- `internal/verify` is the canonical 14-step verifier used by CLI verify, gateway, and SDK middleware paths.
- `internal/policy` evaluates ordered policy rules with first-match-wins semantics and enforced default-deny behavior.
- `internal/gateway` verifies incoming tokens, strips raw `Authorization`, injects `X-OathMesh-*` identity headers, then proxies upstream.
- `internal/core` contains shared protocol types and error taxonomy consumed across issuer/verify/gateway/SDKs.
- `tests/conformance` is the cross-language contract harness that keeps Go, Node, and Python SDK behavior aligned.

## Key conventions
- Use canonical receiver auth header format: `Authorization: OathMesh <token>`.
- Keep the core doctrine text consistent in protocol/security-facing work: **"OathMesh authenticates the caller. The receiver authorizes the request."**
- Subject identifiers are URI-like and limited to these schemes: `svc://`, `agent://`, `job://`, `tool://`, `user://` (enforced in both CLI and issuer handlers).
- Preserve verification step ordering/commenting in `internal/verify/verify.go`; step labels are security-significant for auditability and review.
- Prefer `core.OathMeshError` (`error`, `message`, `fix`, optional `request_id`) for outward-facing auth failures; do not leak internal signing/verification internals in API responses.
- Treat `internal/core` as the dependency-minimal protocol boundary used everywhere else.
- Policy validation expects an explicit terminal default deny rule: `{ name: "default", allow: false }`; evaluator still denies when no rule matches.
- Middleware and gateway flows intentionally remove `Authorization` before forwarding to avoid raw token leakage upstream.
- Changes to protocol format, signing algorithms, package dependency graph, or policy schema should be accompanied by an ADR in `docs/decisions/`.
