---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Git conventions — commit format, branch naming, hooks"
---

# Git Conventions

## Commit Message Format

Use Conventional Commits with OathMesh scope:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | When |
|---|---|
| `feat` | New feature or capability |
| `fix` | Bug fix |
| `refactor` | Code restructuring without behavior change |
| `test` | Adding or updating tests |
| `docs` | Documentation changes |
| `chore` | Build system, dependencies, CI/CD |
| `security` | Security-related changes (even if also a fix) |
| `perf` | Performance improvement |

### Scopes

| Scope | Module |
|---|---|
| `issuer` | Issuer service (`/issuer/`) |
| `sdk-node` | Node.js SDK (`/sdk-node/`) |
| `sdk-python` | Python SDK (`/sdk-python/`) |
| `cli` | CLI tool |
| `gateway` | Gateway proxy |
| `spec` | Protocol specification files (`/spec/`) |
| `docs` | Documentation (`/docs/`) |
| `examples` | Example code (`/examples/`) |
| `config` | Configuration changes (`.antigravity/`, CI/CD) |

### Examples

```
feat(issuer): implement POST /v1/token mint endpoint
fix(sdk-node): correct audience validation for multi-audience configs
refactor(issuer): extract signing logic into separate Signer interface
test(sdk-python): add integration tests for JWKS cache refresh
docs(spec): add JSON Schema for token claims
security(issuer): enforce EdDSA algorithm allowlist in verification
chore: update Go dependencies to latest patch versions
```

### Footer Tags

| Tag | Purpose |
|---|---|
| `Closes #N` | Links to issue |
| `ADR: ADR-NNN` | References architecture decision |
| `BREAKING CHANGE:` | Indicates breaking API change (must be in footer) |

## Branch Naming

```
<type>/<brief-description>
```

| Pattern | Example |
|---|---|
| `feat/mint-endpoint` | Feature branch for mint API |
| `fix/audience-validation` | Bug fix branch |
| `refactor/signer-interface` | Refactoring branch |
| `hotfix/replay-cache-bypass` | Hotfix branch |
| `docs/quickstart-guide` | Documentation branch |

### Rules

- `main` is the default branch — always deployable
- Feature branches are created from `main` and merged back via pull request
- Hotfix branches may bypass normal review for speed but require post-mortem (`workflows/hotfix.md`)
- Delete branches after merge
- Never force-push to `main`

## Tags

Release tags follow semantic versioning:

```
v0.1.0   — first MVP release
v0.1.1   — patch release
v0.2.0   — minor feature release
v1.0.0   — first stable release (protocol and API considered stable)
```

## Hooks (Recommended)

### Pre-commit
- Run linter for changed files
- Check for secrets (basic pattern matching for private keys, API keys)
- Validate commit message format

### Pre-push
- Run full test suite for affected modules
- Verify no `.env` files are staged

## Ignored Files

`.gitignore` must include at minimum:
```
.env
.env.*
*.key
*.pem
*.p12
node_modules/
dist/
__pycache__/
.mypy_cache/
coverage/
*.log
```

Changes to `.gitignore` require human approval per `rules/security-redlines.md`.
