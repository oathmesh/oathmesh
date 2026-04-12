---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Linting and formatting configuration per language"
---

# Linting & Formatting

## Go

### Formatter
- `gofmt` — mandatory, zero-config
- Run on save, run in CI

### Linter
- **Tool**: `golangci-lint`
- **Config file**: `.golangci.yml` at repo root

```yaml
# .golangci.yml
run:
  timeout: 5m

linters:
  enable:
    - errcheck        # unchecked errors
    - govet           # suspicious constructs
    - staticcheck     # comprehensive static analysis
    - unused          # unused code
    - gosimple        # simplifications
    - ineffassign     # unused assignments
    - gocritic        # opinionated checks
    - revive          # extensible linter
    - gosec           # security issues
    - misspell        # spelling
    - bodyclose       # unclosed HTTP response bodies
    - noctx           # HTTP requests without context
  disable:
    - depguard        # we manage deps via rules/core.md instead

linters-settings:
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

### Enforcement
- CI fails on any lint warning
- No `//nolint` without a comment explaining why

## TypeScript (Node SDK)

### Formatter
- **Tool**: Prettier
- **Config file**: `.prettierrc` in `/sdk-node/`

```json
{
  "semi": true,
  "singleQuote": true,
  "trailingComma": "all",
  "printWidth": 100,
  "tabWidth": 2,
  "arrowParens": "always"
}
```

### Linter
- **Tool**: ESLint with TypeScript parser
- **Config file**: `eslint.config.js` in `/sdk-node/`

Key rules:
- `@typescript-eslint/no-explicit-any`: error
- `@typescript-eslint/strict-boolean-expressions`: error
- `@typescript-eslint/no-unused-vars`: error
- `no-console`: warn (use structured logging instead)
- `eqeqeq`: error

### Type Checking
- **Tool**: `tsc` with `strict: true`
- **Config file**: `tsconfig.json` in `/sdk-node/`

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "exactOptionalPropertyTypes": true
  }
}
```

### Enforcement
- CI fails on any ESLint error or TypeScript error
- Prettier formatting enforced via CI check

## Python (Python SDK)

### Formatter + Linter
- **Tool**: Ruff (replaces flake8, isort, black)
- **Config file**: `ruff.toml` in `/sdk-python/`

```toml
[tool.ruff]
target-version = "py311"
line-length = 100

[tool.ruff.lint]
select = [
  "E",    # pycodestyle errors
  "W",    # pycodestyle warnings
  "F",    # pyflakes
  "I",    # isort
  "N",    # pep8-naming
  "UP",   # pyupgrade
  "S",    # bandit (security)
  "B",    # flake8-bugbear
  "A",    # flake8-builtins
  "C4",   # flake8-comprehensions
  "PT",   # flake8-pytest-style
  "RUF",  # ruff-specific rules
]

[tool.ruff.format]
quote-style = "double"
```

### Type Checking
- **Tool**: mypy with strict mode
- **Config**: `pyproject.toml` in `/sdk-python/`

```toml
[tool.mypy]
strict = true
warn_return_any = true
warn_unused_configs = true
```

### Enforcement
- CI fails on any Ruff error or mypy error
- `ruff format --check` enforced in CI

## YAML (Policy Files, Config)

### Validation
- YAML policy files validated against JSON Schema during tests
- 2-space indentation (no tabs)
- Quoted strings for URIs and glob patterns

## CI Integration

All linters run in CI for every pull request:
1. Go: `golangci-lint run`
2. Node: `npx eslint . && npx prettier --check .`
3. Python: `ruff check . && ruff format --check . && mypy .`
4. YAML: schema validation in test suite
