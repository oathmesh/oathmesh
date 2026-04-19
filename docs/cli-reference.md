# OathMesh CLI Reference

<p align="center">
  <img src="../assets/logo.png" width="100" alt="OathMesh Logo">
</p>

<p align="center">
  <b>Command-line interface for minting, serving, and verifying tokens.</b>
</p>

<p align="center">
  <a href="https://github.com/oathmesh/oathmesh/releases">
    <img src="https://img.shields.io/github/v/release/oathmesh/oathmesh" alt="CLI Version">
  </a>
</p>

---

> 🆕 **New here?** Follow the Start Here flow: [README](../README.md) → [QUICKSTART.md](../QUICKSTART.md) → [GETTING_STARTED.md](GETTING_STARTED.md) → [INDEX.md](INDEX.md).

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
| `--tenant` | | | Tenant scope binding securely configuring namespace boundaries |
| `--rqh` | | | Request hash binding (`sha256:...` format) |
| `--inspect` | | | Decode and pretty-print the minted token (with UNVERIFIED warning) |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Signing failure or CLI usage error |
| `2` | Runtime config error (e.g., missing signing key, invalid `--sub`) |

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

# Mint with environment defaults in shell scripts
export OATHMESH_ISSUER="https://issuer.oathmesh.tech"
oathmesh mint \
  --sub "job://github/acme/deploy" \
  --aud "https://inventory.internal" \
  --act "inventory.write" \
  --ttl 90 > token.txt
```

### Common Mint Errors

| Symptom | Likely Cause | Fix |
|---|---|---|
| `config error` with exit `2` | Missing signing key config | Set `OATHMESH_KMS_KEY_ID` or one of `OATHMESH_PRIVATE_KEY`, `OATHMESH_PRIVATE_KEY_B64`, `OATHMESH_PRIVATE_KEY_FILE` |
| `invalid subject scheme` | `--sub` is not a supported URI scheme | Use `svc://`, `agent://`, `job://`, `tool://`, or `user://` |
| Token TTL not what you requested | TTL hint > max allowed | Keep `--ttl` at 300s or below |

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
| `1` | Auth failure (signature, expiry, policy, replay, etc.) or CLI usage error |
| `2` | Runtime config error (e.g., missing token, bad local keyset) |

### Examples

```bash
# Verify with explicit token and issuer
oathmesh verify <token> --audience "https://api.internal" --issuer "https://issuer.oathmesh.tech"

# Pipe from mint (dev mode with local keys)
oathmesh mint --sub "agent://repo/acme/bot" --aud "https://api.internal" --act "read" \
  | oathmesh verify --audience "https://api.internal" --local-keys

# JSON output
oathmesh verify --json <token> --audience "https://api.internal" --issuer "https://issuer.oathmesh.tech"

# Verify from stdin and read only exit code
cat token.txt | oathmesh verify --audience "https://inventory.internal" --issuer "https://issuer.oathmesh.tech" --quiet
echo $?  # 0 valid, 1 auth failure, 2 config error
```

### Common Verify Errors

| Symptom | Likely Cause | Fix |
|---|---|---|
| `issuer_untrusted` | Issuer not passed or not configured | Add `--issuer "https://issuer..."` or fix verifier config |
| `audience_mismatch` | `--audience` differs from token `aud` | Verify using the exact expected audience |
| `token_expired` | Token lifetime elapsed | Mint a new token (tokens are intentionally short-lived) |
| `replay_detected` | Same token used again | Mint per request/work unit; do not reuse tokens |

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

### `POST /v1/token` Response

```json
{
  "token": "<om+jwt>",
  "expires_in": 120,
  "token_type": "OathMesh"
}
```

`expires_in` is the remaining token lifetime in seconds at response time (server-side TTL cap is 300 seconds).

### Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--port` | | `4000` | Listen port |
| `--config` | | | Reserved config file path (currently unused; env vars are source of truth) |
| `--gateway` | | | Enable reverse proxy / gateway mode (requires policy in non-development) |

### First-Run Notes

- `oathmesh serve` reads configuration from environment variables.
- You must provide signing configuration via either `OATHMESH_KMS_KEY_ID` or one of:
  - `OATHMESH_PRIVATE_KEY`
  - `OATHMESH_PRIVATE_KEY_B64`
  - `OATHMESH_PRIVATE_KEY_FILE`

### Gateway Mode

- `oathmesh serve --gateway` also requires:
  - `OATHMESH_GATEWAY_UPSTREAM`
  - `OATHMESH_GATEWAY_AUDIENCE`
  - `OATHMESH_GATEWAY_ISSUERS`
- If `OATHMESH_ENV` is not `development`, it also requires:
  - `OATHMESH_GATEWAY_POLICY`

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Clean shutdown |
| `1` | Startup error |

### Examples

```bash
OATHMESH_PRIVATE_KEY_FILE=./private.pem oathmesh serve
OATHMESH_PRIVATE_KEY_FILE=./private.pem oathmesh serve --port 8080
OATHMESH_PRIVATE_KEY_FILE=./private.pem OATHMESH_GATEWAY_UPSTREAM=http://localhost:3000 OATHMESH_GATEWAY_AUDIENCE=https://api.internal OATHMESH_GATEWAY_ISSUERS=http://localhost:4000 oathmesh serve --gateway
OATHMESH_ENV=production OATHMESH_PRIVATE_KEY_FILE=./private.pem OATHMESH_GATEWAY_UPSTREAM=https://upstream.internal OATHMESH_GATEWAY_AUDIENCE=https://api.internal OATHMESH_GATEWAY_ISSUERS=https://issuer.oathmesh.tech OATHMESH_GATEWAY_POLICY=policy/production.json oathmesh serve --gateway
```

### Common Serve Errors

| Symptom | Likely Cause | Fix |
|---|---|---|
| Startup fails immediately | Missing required environment config | Set issuer/signing env vars, and gateway env vars when `--gateway` is enabled |
| `address already in use` | Port conflict | Change `--port` or stop the other process |
| `/healthz` unhealthy | Dependency or config issue | Run with `--verbose` and check startup logs |

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

## `oathmesh revoke`

Revoke all tokens for a given subject by publishing to the active Redis revocation backend.

### Arguments & Configuration

Requires `OATHMESH_MINT_SECRET`. Issuer URL can be provided with `--issuer` or `OATHMESH_ISSUER`.

| Arg | Required | Description |
|-----|----------|-------------|
| `<subject>` | ✓ | Subject URI string to revoke entirely (e.g. `svc://bad-actor`) |
| `--issuer` | | Issuer URL (can use `OATHMESH_ISSUER` instead) |

### Examples

```bash
export OATHMESH_MINT_SECRET="development_secret"
export OATHMESH_ISSUER="http://localhost:4000"

oathmesh revoke 'svc://compromised-agent'
```

---

## `oathmesh unrevoke`

Revert a subject's revocation status from the Redis caching backend.

### Arguments & Configuration

Requires `OATHMESH_MINT_SECRET`. Issuer URL can be provided with `--issuer` or `OATHMESH_ISSUER`.

| Arg | Required | Description |
|-----|----------|-------------|
| `<subject>` | ✓ | Subject URI string to unrevoke (e.g. `svc://bad-actor`) |
| `--issuer` | | Issuer URL (can use `OATHMESH_ISSUER` instead) |

### Examples

```bash
export OATHMESH_MINT_SECRET="development_secret"
export OATHMESH_ISSUER="http://localhost:4000"
oathmesh unrevoke 'svc://compromised-agent'
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
oathmesh policy validate ./policy/production.json
oathmesh policy validate ./policy/example.pkl
oathmesh policy validate --json ./policy/production.json
```

### Common Policy Validation Errors

| Symptom | Likely Cause | Fix |
|---|---|---|
| `version` error | Policy version not supported | Set `version = 1` |
| `default deny` error | Last rule is not explicit deny | End with `{ name: "default", allow: false }` |
| Pkl parse failure | Invalid Pkl syntax | Validate syntax and ensure `pkl` is installed/in PATH |

---

## Pipeline Example

The canonical mint → verify pipeline:

```bash
export OATHMESH_PRIVATE_KEY="$(cat private.pem)"
export OATHMESH_ISSUER="https://issuer.oathmesh.tech"

oathmesh mint \
  --sub "agent://repo/acme/deploy-bot" \
  --aud "https://inventory.internal" \
  --act "inventory.write" \
| oathmesh verify \
  --audience "https://inventory.internal" \
  --issuer "https://issuer.oathmesh.tech" \
  --local-keys
```

## Troubleshooting Workflow

When a command fails:

1. Re-run with `--verbose` to capture detailed logs.
2. Re-run with `--json` for machine-readable error fields.
3. Confirm expected exit code (`0`, `1`, or `2`) to separate auth failures from config failures.
4. For verification failures, map `error.code` using [Error Taxonomy](protocol/error-taxonomy.md).
