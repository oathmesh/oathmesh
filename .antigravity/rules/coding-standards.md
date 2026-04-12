---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Language-specific coding conventions for every OathMesh module"
---

# OathMesh Coding Standards

These standards apply to all code written for OathMesh. They are organized by language and module. When `rules/core.md` performance budgets conflict with style preferences, performance wins.

## General Conventions (All Languages)

### Naming

- **Files**: `kebab-case` for source files, `UPPER_CASE` for constants files
- **Directories**: `kebab-case`
- **Environment variables**: `OATHMESH_` prefix, `UPPER_SNAKE_CASE` (e.g., `OATHMESH_ISSUER_URL`, `OATHMESH_MAX_TTL`)
- **Config keys**: `snake_case` in YAML/JSON (e.g., `max_ttl_seconds`, `trusted_issuers`)
- **HTTP headers**: standard capitalisation (e.g., `Authorization`, `Content-Type`)

### File Structure

Every source file:
1. License header (Apache 2.0)
2. Package/module declaration
3. Imports (stdlib first, then third-party, then internal — separated by blank lines)
4. Constants and types
5. Public API
6. Private helpers
7. No file exceeds 500 lines — split if approaching

### Comments

- Public APIs: doc comments explaining what, why, and any constraints
- Complex logic: explain the why, not the what
- No commented-out code in main branches — use version control
- No `TODO` without a linked issue number or ADR reference

### Error Handling

- Errors must include context about what operation was attempted
- Use structured error types with codes, not string matching
- Never swallow errors silently
- Log errors at the point of handling, not the point of creation
- Error codes follow the taxonomy in `spec/errors.json` (once defined)

---

## Go (Issuer, Gateway, CLI)

### Style

- Follow `gofmt` and `go vet` with zero warnings
- Use `golangci-lint` with the config in `tools/linting.md`
- Effective Go and Go Code Review Comments are the baseline

### Naming

- Package names: short, lowercase, no underscores (e.g., `issuer`, `verifier`, `policy`)
- Exported types: `PascalCase` (e.g., `OathToken`, `VerifiedCallerContext`, `PolicyRule`)
- Unexported types: `camelCase`
- Interfaces: describe behavior, not implementation (e.g., `Signer`, `Verifier`, `PolicyEvaluator`)
- Test files: `*_test.go` in the same package

### Patterns

- Use `context.Context` as the first parameter of all public functions that do I/O
- Use `errors.Is/As` for error inspection — never compare error strings
- Use structured logging with `slog` (Go 1.21+) — never `log.Printf`
- HTTP handlers return structured JSON errors, never plain text
- Configuration via environment variables with `OATHMESH_` prefix, parsed at startup
- No `init()` functions — explicit initialization only
- No global mutable state — pass dependencies through constructors

### Project Layout

```
/issuer/
  cmd/
    oathmesh/         — main entrypoint for CLI + server
      main.go
  internal/
    issuer/           — token minting, key management
    verifier/         — token verification logic (shared with gateway)
    policy/           — YAML policy loading and evaluation
    audit/            — structured audit event emission
    transport/        — HTTP handlers, middleware
    config/           — configuration loading
  pkg/
    oathtoken/        — public token types and claim definitions
    jwks/             — JWKS fetching and caching
```

### Testing

- Table-driven tests for all claim validation logic
- Golden file tests for token encoding/decoding
- Integration tests use `httptest.Server` — no real network calls in unit tests
- Benchmark tests for verification and policy evaluation paths

---

## TypeScript/Node.js (SDK)

### Style

- Strict TypeScript — `strict: true` in `tsconfig.json`, no `any` types
- Use ESM modules (`"type": "module"` in `package.json`)
- Prettier for formatting, ESLint for linting (configs in `tools/linting.md`)

### Naming

- Files: `kebab-case.ts` (e.g., `oath-token.ts`, `verified-caller-context.ts`)
- Types/interfaces: `PascalCase` (e.g., `OathToken`, `VerifiedCallerContext`, `PolicyRule`)
- Functions: `camelCase` (e.g., `verifyOathToken`, `evaluatePolicy`)
- Constants: `UPPER_SNAKE_CASE` (e.g., `DEFAULT_TTL_SECONDS`, `MAX_TTL_SECONDS`)
- No default exports — use named exports exclusively

### Patterns

- Middleware pattern: `createOathmeshMiddleware(options)` returns Express/Koa/Hono-compatible middleware
- The verified caller context is attached to the request object at `req.oathmesh`
- Async/await everywhere — no raw Promise chains or callbacks
- Use `zod` for runtime validation of config and claims
- JWT operations use `jose` library (standard, well-maintained)
- No `node-jsonwebtoken` — it lacks EdDSA support and is less actively maintained

### Project Layout

```
/sdk-node/
  src/
    index.ts              — public API re-exports
    middleware.ts          — Express/Koa/Hono middleware factory
    verifier.ts           — token verification logic
    policy.ts             — YAML policy loading and evaluation
    types.ts              — OathToken, VerifiedCallerContext, PolicyRule types
    errors.ts             — structured error types
    jwks.ts               — JWKS fetching and caching
    audit.ts              — audit event emission
  test/
    verifier.test.ts
    policy.test.ts
    middleware.test.ts
  package.json
  tsconfig.json
  vitest.config.ts
```

---

## Python (SDK)

### Style

- Python 3.11+ minimum
- Type hints on all public functions and methods
- `ruff` for linting and formatting (config in `tools/linting.md`)
- `mypy` with `strict = true`

### Naming

- Files: `snake_case.py` (e.g., `oath_token.py`, `verified_caller_context.py`)
- Classes: `PascalCase` (e.g., `OathToken`, `VerifiedCallerContext`)
- Functions: `snake_case` (e.g., `verify_oath_token`, `evaluate_policy`)
- Constants: `UPPER_SNAKE_CASE`
- Private: single underscore prefix (e.g., `_parse_claims`)

### Patterns

- Middleware pattern: ASGI middleware class for FastAPI/Starlette, WSGI middleware for Flask
- The verified caller context is attached to request state at `request.state.oathmesh`
- JWT operations use `PyJWT` with `cryptography` backend or `python-jose` with `cryptography`
- Use `pydantic` for config validation and claim schemas
- Use `httpx` for JWKS fetching (async-native)

### Project Layout

```
/sdk-python/
  src/
    oathmesh/
      __init__.py
      middleware.py       — ASGI/WSGI middleware
      verifier.py         — token verification logic
      policy.py           — YAML policy loading and evaluation
      types.py            — OathToken, VerifiedCallerContext dataclasses
      errors.py           — structured error types
      jwks.py             — JWKS fetching and caching
      audit.py            — audit event emission
  tests/
    test_verifier.py
    test_policy.py
    test_middleware.py
  pyproject.toml
  ruff.toml
```

---

## YAML (Policy Files)

### Conventions

- 2-space indentation
- No tabs
- Quoted strings for URIs and glob patterns
- Policy files must start with `version: 1`
- Default deny is implicit — if no rule matches, the request is denied
- Rule names are `kebab-case` and unique within a file

### Example

```yaml
version: 1

issuers:
  - "https://issuer.oathmesh.dev"

audiences:
  - "https://inventory.internal"

rules:
  - name: storefront-read
    match:
      sub: "agent://repo/acme/*"
      act: "inventory.read"
    allow: true

  - name: deploy-write
    match:
      sub: "agent://repo/acme/deploy-bot"
      act: "inventory.write"
      src:
        type: "github_actions"
        repo: "acme/storefront"
    allow: true

  - name: default-deny
    allow: false
```
