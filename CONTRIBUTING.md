# Contributing to OathMesh

## Development Setup

### Prerequisites

- Go 1.22+
- Docker and Docker Compose (for integration tests)
- Node.js 18+ (for Node SDK development)
- Python 3.8+ (for Python SDK development)

### Clone and Build

```bash
git clone https://github.com/oathmesh/oathmesh.git
cd oathmesh

# Install Go dependencies
go mod download

# Build the CLI
make build

# Generate a development key
openssl genpkey -algorithm Ed25519 -out private.pem

# Run tests
make test
```

### Running Locally

```bash
# Start all services
make docker-up

# Run the issuer in dev mode
make dev

# Run the full demo
make demo
```

## Running Tests

```bash
# All tests
make test

# With race detector
make race

# Specific package
go test -v ./internal/verify/...

# Node SDK tests
cd sdk/node && npm install && npm test

# Python SDK tests
cd sdk/python && pip install -e .[test] && pytest tests/
```

## Submitting a Pull Request

### Before You Submit

1. **Run the full test suite:** `make race` (includes race detector)
2. **Run the linter:** `make lint`
3. **Verify the build:** `make build`
4. **Check formatting:** `gofmt -l .` (should return no output)

### PR Requirements

- All tests must pass
- No race conditions (`go test -race`)
- No `go vet` warnings
- Commit messages follow conventional commits: `feat:`, `fix:`, `docs:`, `chore:`
- Security-relevant changes require explicit review

### Security Checklist

If your PR touches `internal/sign/`, `internal/verify/`, or any crypto code:

- [ ] Private key material never logged, never in HTTP responses
- [ ] Full token strings never logged
- [ ] `alg: "none"` rejected at Step 02
- [ ] Audience matching is exact (no globs)
- [ ] TTL enforced server-side, max 300s
- [ ] `jti` generated via `uuid.New()`
- [ ] JWKS fetch uses dedicated client with 5s timeout
- [ ] Replay cache operations are thread-safe

## Project Structure

See [ARCHITECTURE.md](ARCHITECTURE.md) for package descriptions and dependency graph.

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use `slog` for structured logging
- Never expose raw Go errors externally — use the error taxonomy
- Comments on verification steps must cite step number and security rationale
- `internal/core` must have zero external dependencies

## ADRs

Architectural decisions are documented in `docs/decisions/`. Any change to:
- Protocol format or claims
- Signing algorithms
- Package structure or dependency graph
- Policy schema

...requires a new ADR before implementation.
