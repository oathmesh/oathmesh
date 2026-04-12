---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
skill: "api-design"
triggers:
  - "designing issuer API endpoints"
  - "designing SDK public API surface"
  - "implementing CLI commands"
  - "designing configuration schemas"
  - "implementing error response format"
dependencies:
  - "skills/auth.md (for token lifecycle)"
  - "skills/data-modeling.md (for request/response schemas)"
  - "rules/coding-standards.md (for language-specific conventions)"
---

# Skill: API Design

This skill covers the design of all OathMesh public-facing APIs — the issuer's REST API, the SDK middleware API, and the CLI interface.

## Issuer REST API

### Base Path

All issuer API endpoints are versioned under `/v1/`:

### Endpoints

| Method | Path | Purpose | Auth Required |
|---|---|---|---|
| `POST` | `/v1/token` | Mint a new Oath Token | Yes (caller bootstrap identity) |
| `GET` | `/.well-known/oathmesh-issuer` | Issuer metadata discovery | No |
| `GET` | `/.well-known/jwks.json` | Public keys for verification | No |
| `GET` | `/healthz` | Health check (liveness) | No |
| `GET` | `/readyz` | Readiness check | No |

### POST /v1/token — Mint Token

**Request:**
```json
{
  "audience": "https://inventory.internal",
  "action": "inventory.write",
  "scope": ["inventory.read", "inventory.write"],
  "reason": "sync catalog after deploy",
  "ttl_seconds": 120,
  "binding": {
    "request_hash": "sha256:abc123..."
  }
}
```

- `audience` — required. The intended receiver URI.
- `action` — required. The requested operation family.
- `scope` — optional. Full list of permitted operations. If omitted, defaults to `[action]`.
- `reason` — optional. Human-readable reason for the request.
- `ttl_seconds` — optional. Requested TTL. Clamped to issuer's max (default: 120, max: 300).
- `binding.request_hash` — optional. Request hash for binding mode.

The caller's identity (`sub`) and source provenance (`src`) are derived from the caller's bootstrap authentication, not from the request body. Callers must never be able to self-assert their identity.

**Success Response (200):**
```json
{
  "token": "eyJhbGci...",
  "token_type": "om+jwt",
  "expires_in": 120,
  "token_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Error Response (4xx/5xx):**
```json
{
  "error": "invalid_audience",
  "error_description": "audience 'https://unknown.internal' is not in the issuer's allowed audience list",
  "error_uri": "https://docs.oathmesh.dev/errors/invalid_audience"
}
```

### Error Format

All errors follow this structure:

```json
{
  "error": "<error_code>",
  "error_description": "<human-readable explanation with specific details>",
  "error_uri": "<link to documentation for this error>"
}
```

Error descriptions must include:
- What was attempted
- What went wrong
- Enough context to diagnose (e.g., the mismatched values)

Error descriptions must NEVER include:
- Full tokens
- Signing keys
- Internal stack traces (in production mode)

## SDK Middleware API

### Node.js

```typescript
import { createOathmeshMiddleware } from '@oathmesh/node';

const oathmesh = createOathmeshMiddleware({
  audience: 'https://inventory.internal',
  trustedIssuers: [
    {
      issuer: 'https://issuer.oathmesh.dev',
      jwksUri: 'https://issuer.oathmesh.dev/.well-known/jwks.json',
    },
  ],
  policyFile: './policy.yaml',
  audit: {
    output: 'stdout',
    format: 'json',
  },
});

app.use(oathmesh);

// In route handlers:
app.get('/items', (req, res) => {
  const caller = req.oathmesh; // VerifiedCallerContext
  console.log(caller.principal.subject); // "agent://repo/acme/deploy-bot"
  console.log(caller.action);           // "inventory.read"
});
```

### Python

```python
from oathmesh import OathmeshMiddleware

app = FastAPI()
app.add_middleware(
    OathmeshMiddleware,
    audience="https://inventory.internal",
    trusted_issuers=[
        {
            "issuer": "https://issuer.oathmesh.dev",
            "jwks_uri": "https://issuer.oathmesh.dev/.well-known/jwks.json",
        },
    ],
    policy_file="./policy.yaml",
)

@app.get("/items")
async def get_items(request: Request):
    caller = request.state.oathmesh  # VerifiedCallerContext
    print(caller.principal.subject)   # "agent://repo/acme/deploy-bot"
```

### Design Principles

1. **One function/class to get started** — `createOathmeshMiddleware` (Node), `OathmeshMiddleware` (Python)
2. **Configuration is explicit** — no auto-discovery magic, no environment variable guessing
3. **Verified context is always available at a standard location** — `req.oathmesh` (Node), `request.state.oathmesh` (Python)
4. **Errors are structured** — middleware returns JSON error responses, never HTML or plaintext
5. **Audit is built-in** — middleware emits audit events by default, configurable output

## CLI Interface

```
oathmesh mint     — Mint a new Oath Token (requires issuer connection or local mode)
oathmesh verify   — Verify an existing Oath Token against a trusted issuer
oathmesh inspect  — Decode and display token claims without verification
oathmesh serve    — Start the issuer server
oathmesh keys     — Key management subcommands
  oathmesh keys generate  — Generate a new signing key pair
  oathmesh keys rotate    — Rotate the active signing key
  oathmesh keys list      — List all keys in the JWKS
```

### CLI Output

- Default output: human-readable formatted text with colors
- `--json` flag: machine-readable JSON output
- `--quiet` flag: suppress non-essential output
- Exit codes: 0 (success), 1 (error), 2 (invalid input)

### CLI Error Format

```
ERROR: audience_mismatch

  Token was minted for:  https://billing.internal
  Received by:           https://inventory.internal

  The audience in the Oath Token must exactly match the
  receiver's configured audience.

  See: https://docs.oathmesh.dev/errors/audience_mismatch
```
