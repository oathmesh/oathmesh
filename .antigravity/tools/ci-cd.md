---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "CI/CD pipeline awareness and build targets"
---

# CI/CD Awareness

## Pipeline Structure

OathMesh uses GitHub Actions for CI/CD. The AI should be aware of these pipeline components when writing code, tests, or evaluating changes.

### CI Pipeline (runs on every pull request)

```yaml
# .github/workflows/ci.yml
name: CI

on:
  pull_request:
    branches: [main]

jobs:
  go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go vet ./...
      - run: golangci-lint run
      - run: go test -race -coverprofile=coverage.txt ./...
      - run: go build -o /dev/null ./cmd/oathmesh

  node:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: sdk-node
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
      - run: npm ci
      - run: npx eslint .
      - run: npx prettier --check .
      - run: npx tsc --noEmit
      - run: npx vitest run --coverage

  python:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: sdk-python
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - run: pip install -e ".[dev]"
      - run: ruff check .
      - run: ruff format --check .
      - run: mypy .
      - run: pytest --cov
```

### Build Targets

| Target | Command | Output | When |
|---|---|---|---|
| Go issuer binary | `go build -o oathmesh ./cmd/oathmesh` | Single static binary | CI + Release |
| Go tests | `go test -race ./...` | Test results + coverage | Every PR |
| Node SDK | `npm run build` (tsc) | `dist/` directory | CI + npm publish |
| Node tests | `npx vitest run` | Test results + coverage | Every PR |
| Python SDK | `python -m build` | wheel + sdist | CI + PyPI publish |
| Python tests | `pytest --cov` | Test results + coverage | Every PR |
| Docker image | `docker build -t oathmesh .` | Container image < 30MB | Release |
| E2E tests | `docker compose -f examples/docker-compose.test.yml up --abort-on-container-exit` | Pass/fail | Release |

### Release Pipeline (runs on tag push)

```
Tag v* → Build binaries (Linux/macOS/Windows) → Build Docker image → Publish npm → Publish PyPI → Create GitHub Release
```

## What the AI Should Know

### Before Writing Code
- All code must pass the CI checks above before merging
- The AI should run relevant checks locally before declaring a task complete:
  - Go: `go vet ./... && go test ./...`
  - Node: `npx tsc --noEmit && npx vitest run`
  - Python: `mypy . && pytest`

### Dependency Management

| Package Manager | Lock File | Add Dependency | Update Dependency |
|---|---|---|---|
| Go modules | `go.sum` | `go get <package>` | `go get -u <package>` |
| npm | `package-lock.json` | `npm install <package>` | `npm update <package>` |
| pip | `requirements.txt` or `pyproject.toml` | Add to `pyproject.toml` | `pip install --upgrade <package>` |

Rules (from `rules/core.md` § Dependency Rules):
- Prefer standard library over third-party
- No vendored crypto — use well-audited packages only
- Document justification for dependencies with > 3 transitive dependencies
- Adding dependencies requires CI to pass (no new vulnerabilities)

### Pipeline Secrets

The AI must NEVER:
- Access, modify, or reference CI/CD pipeline secrets
- Suggest hardcoding secrets in workflow files
- Modify `.github/workflows/` files that reference secrets without human review

Pipeline secrets are managed exclusively by the human (per `rules/security-redlines.md`).

### Docker Image Strategy

```dockerfile
# Multi-stage build for minimal image
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o oathmesh ./cmd/oathmesh

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/oathmesh /
ENTRYPOINT ["/oathmesh"]
```

Target: < 30MB final image (per `rules/core.md` § Performance Budgets).
Base image changes require human approval (per `rules/security-redlines.md`).
