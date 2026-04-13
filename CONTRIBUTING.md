# Contributing to OathMesh

> 💡 Welcome! Here's how to contribute to OathMesh.

## Quick Reference Card

```bash
# Clone and setup
git clone https://github.com/oathmesh/oathmesh.git
cd oathmesh
go mod download
openssl genpkey -algorithm Ed25519 -out private.pem

# Build & test
make build           # Build CLI
make test            # Run all tests
make race            # Tests + race detector
make lint            # Run linter

# Run locally
make docker-up       # Start services
make demo            # Full demo
```

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
# All tests (Go + Node + Python)
make test-all

# Go only (with race detector)
make race

# Node SDK only
make test-node

# Python SDK only
make test-python

# Specific Go package
go test -v ./internal/verify/...
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

## PR Checklist

Before submitting a pull request, verify:

- [ ] All tests pass: `make test`
- [ ] No race conditions: `make race`
- [ ] No lint errors: `make lint`
- [ ] Build succeeds: `make build`
- [ ] Demo runs: `./demo.sh`
- [ ] Docs updated if needed
- [ ] Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)

## Quick Start Checklist

```bash
# 1. Fork & clone
git clone https://github.com/YOUR_USERNAME/oathmesh.git
cd oathmesh

# 2. Create feature branch
git checkout -b feature/your-feature

# 3. Make changes, then test
make race && make lint

# 4. Commit with conventional message
git commit -m "feat: add new verification step"

# 5. Push and open PR
git push -u origin feature/your-feature
```

## Issue Templates

- 🐛 **[Bug Report](.github/ISSUE_TEMPLATE/bug_report.md)** — Use for bugs
- 💡 **[Feature Request](.github/ISSUE_TEMPLATE/feature_request.md)** — Use for new features
