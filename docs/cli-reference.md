# OathMesh CLI Reference

<p align="center">
  <b>Command-line interface for minting, serving, and verifying tokens.</b>
</p>

<p align="center">
  <a href="https://github.com/oathmesh/oathmesh/releases">
    <img src="https://img.shields.io/github/v/release/oathmesh/oathmesh" alt="CLI Version">
  </a>
</p>

---

> 🆕 **New here?** Start with the [Quick Start](../README.md#-quick-start) in the main README.

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | | Machine-readable JSON output for all commands |
| `--quiet` | `-q` | Suppress informational output (errors to stderr only) |
| `--verbose` | `-v` | Enable debug logging via slog |

---

## `oathmesh mint`

Mint a signed Oath Token for machine-to-machine authentication.

The token is output on stdout (pipeable).

### Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--sub` | ✓ | | Subject URI (must use `svc://`, `agent://`, `job://`, `tool://`, or `user://` scheme) |
| `--aud` | ✓ | | Audience URL |
| `--act` | ✓ | | Action string |
| `--ttl` | | `120` | TTL hint in seconds (clamped server-side to max 300) |
| `--scope` | | | Scope values (repeatable) |
| `--reason` | | | Reason claim |
| `--env` | | | Environment label |
| `--rqh` | | | Request hash binding (`sha256:...` format) |
| `--inspect` | | | Decode and pretty-print the minted token (with UNVERIFIED warning) |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Signing failure |
| `2` | Config error (missing key, invalid flags) |

### Examples

```bash
# Basic mint
oathmesh mint --sub "agent://repo/acme/deploy-bot" --aud "https://api.internal" --act "inventory.write"

# With TTL, scope, and reason
oathmesh mint \
  --sub "job://github/acme/deploy" \
  --aud "https://api.internal" \
  --act "deploy" \
  --ttl 60 \
  --scope "deploy.write" \
  --reason "scheduled deployment"

# Mint and inspect
oathmesh mint --sub "agent://repo/acme/bot" --aud "https://api.internal" --act "read" --inspect

# Pipe to verify
oathmesh mint --sub "agent://repo/acme/bot" --aud "https://api.internal" --act "read" \
  | oathmesh verify --audience "https://api.internal" --local-keys

# JSON output
oathmesh mint --json --sub "agent://repo/acme/bot" --aud "https://api.internal" --act "read"
```

---

## `oathmesh verify`

Verify an Oath Token using the full 14-step verification pipeline.

The token can be provided as a positional argument, via `--token` flag, or read from stdin.

### Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--audience` | ✓ | | Receiver audience URL |
| `--issuer` | | | Trusted issuer URLs (repeatable) |
| `--token` | | | Token string (or provide as positional arg, or pipe to stdin) |
| `--local-keys` | | | Use local keyset instead of fetching JWKS from issuer URL (dev only) |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Valid token |
| `1` | Auth failure (signature, expiry, policy, replay, etc.) |
| `2` | Config error (missing flags, bad keyset) |

### Examples

```bash
# Verify with explicit token and issuer
oathmesh verify <token> --audience "https://api.internal" --issuer "https://issuer.oathmesh.dev"

# Pipe from mint (dev mode with local keys)
oathmesh mint --sub "agent://repo/acme/bot" --aud "https://api.internal" --act "read" \
  | oathmesh verify --audience "https://api.internal" --local-keys

# JSON output
oathmesh verify --json <token> --audience "https://api.internal" --issuer "https://issuer.oathmesh.dev"
```

---

## `oathmesh inspect`

Decode and display the header and claims of an Oath Token **without verification**.

> ⚠ The output is UNVERIFIED — do not trust for authorization decisions.

Shows: header fields, all claims, and an expiry countdown.

### Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--token` | | | Token string (or provide as positional arg, or pipe to stdin) |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Successfully decoded |
| `1` | Parse error |

### Examples

```bash
# Inspect a token
oathmesh inspect <token>

# Pipe from mint
oathmesh mint --sub "agent://repo/acme/bot" --aud "https://api.internal" --act "read" | oathmesh inspect

# JSON output piped to jq
oathmesh inspect --json <token> | jq '.claims.sub'
```

---

## `oathmesh serve`

Start the OathMesh issuer HTTP server.

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/token` | Mint endpoint |
| `POST` | `/v1/exchange/github` | GitHub OIDC exchange |
| `GET` | `/.well-known/jwks.json` | Public keys |
| `GET` | `/.well-known/oathmesh-issuer` | Discovery |
| `GET` | `/healthz` | Liveness check (no auth) |

### Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--port` | | `4000` | Listen port |
| `--config` | | | Pkl config file path |
| `--gateway` | | | Enable reverse proxy / gateway mode |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Clean shutdown |
| `1` | Startup error |

### Examples

```bash
oathmesh serve
oathmesh serve --port 8080
oathmesh serve --config internal/config/issuer.pkl
```

---

## `oathmesh keys rotate`

Generate a new Ed25519 key pair and publish alongside the current key in JWKS.

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Rotation successful |
| `1` | Rotation error |
| `2` | Keyset load error |

### Examples

```bash
oathmesh keys rotate
oathmesh keys rotate --json
```

---

## `oathmesh policy validate`

Validate a `.pkl` or `.json` policy file against the OathMesh policy schema.

### Checks Performed

- `version == 1`
- At least one issuer
- At least one audience
- At least one rule
- Last rule is `{ name: "default", allow: false }`

### Arguments

| Arg | Required | Description |
|-----|----------|-------------|
| `<file>` | ✓ | Path to `.pkl` or `.json` policy file |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Valid policy |
| `1` | Invalid policy (schema errors reported) |

### Examples

```bash
# Validate JSON policy
oathmesh policy validate policy/production.json

# Validate Pkl policy (requires pkl binary in PATH)
oathmesh policy validate policy/production.pkl

# JSON output
oathmesh policy validate --json policy/production.json
```

---

## Pipeline Example

The canonical mint → verify pipeline:

```bash
export OATHMESH_PRIVATE_KEY="$(cat private.pem)"
export OATHMESH_ISSUER="https://issuer.oathmesh.dev"

oathmesh mint \
  --sub "agent://repo/acme/deploy-bot" \
  --aud "https://inventory.internal" \
  --act "inventory.write" \
| oathmesh verify \
  --audience "https://inventory.internal" \
  --issuer "https://issuer.oathmesh.dev" \
  --local-keys
```
