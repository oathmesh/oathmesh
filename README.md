<p align="center">
  <img src="assets/logo.png" width="120" alt="OathMesh Logo">
</p>

# OathMesh

<p align="center">
  <b>Every machine call gets a short-lived, signed identity.</b>
</p>

<p align="center">
  <a href="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml">
    <img src="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml/badge.svg" alt="CI Status">
  </a>
  <a href="https://www.npmjs.com/package/@oathmesh/oathmesh">
    <img src="https://img.shields.io/npm/v/@oathmesh/oathmesh.svg" alt="npm version">
  </a>
  <a href="https://pypi.org/project/oathmesh/">
    <img src="https://img.shields.io/pypi/v/oathmesh.svg" alt="pypi version">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/oathmesh/oathmesh.svg" alt="License">
  </a>
</p>

---

### **Why OathMesh?**
API keys are leaked every day. Static secrets are a security nightmare. **OathMesh** replaces vulnerable long-lived credentials with short-lived (≤300s), scoped, and verifiable **Oath Tokens**.

- **Zero-Trust Identity:** No more "trusted networks". Every service must prove who it is.
- **Fail-Closed Security:** A 14-step verification pipeline rejects malformed, expired, or replayed tokens.
- **Polyglot Ready:** First-class SDKs for **Go**, **Node.js**, and **Python**.
- **Audit Everything:** Every call (allowed or denied) generates a structured audit event.

---

## Quick Start

```bash
git clone https://github.com/oathmesh/oathmesh.git
cd oathmesh

# Generate a development key
openssl genpkey -algorithm Ed25519 -out private.pem

# Build the CLI
make build

# Start services
docker-compose up -d

# Mint a token
TOKEN=$(./bin/oathmesh mint \
  --sub "agent://repo/acme/deploy-bot" \
  --aud "https://inventory.internal" \
  --act "deploy" --quiet)

# Call a protected API
curl -H "Authorization: OathMesh $TOKEN" http://localhost:8081/inventory
```

Run `./demo.sh` for the full automated golden-path demo.

---

## SDKs

| Language | Package | Frameworks |
|---|---|---|
| **Go** | [`github.com/oathmesh/oathmesh`](https://github.com/oathmesh/oathmesh) | chi, stdlib `net/http` |
| **Node.js** | [`@oathmesh/oathmesh`](https://www.npmjs.com/package/@oathmesh/oathmesh) | Express, **Next.js** (App, Pages, Edge) |
| **Python** | [`oathmesh`](https://pypi.org/project/oathmesh/) | FastAPI, Flask, Django |

### Go

```go
r.Use(middleware.OathMeshMiddleware(cfg))
caller := middleware.CallerFrom(r.Context())
```

### Express (TypeScript)

```typescript
import { verifyToken } from '@oathmesh/oathmesh';
app.use(verifyToken({ audience, trustedIssuers }));
// req.oathmeshContext is fully typed
```

### Next.js (App Router)

```typescript
import { withOathMesh } from '@oathmesh/oathmesh/next';
const oathmesh = withOathMesh({ audience, trustedIssuers });

export async function GET(request: NextRequest) {
  const { caller, error } = await oathmesh(request);
  if (error) return error;
  return NextResponse.json({ subject: caller.principal.subject });
}
```

### FastAPI (Python)

```python
from oathmesh import verify_token, VerifierConfig
caller = verify_token(request.headers["authorization"], config)
# caller.principal.subject, caller.action, caller.token_id
```

---

## Architecture

```
Caller ──▶ Issuer ──▶ signs Oath Token (Ed25519, ≤300s TTL)
  │
  └──▶ Receiver (or Gateway)
         ├── 14-step verification pipeline
         ├── Pkl policy evaluation (default deny)
         ├── NDJSON audit event (always — allow AND deny)
         └── VerifiedCallerContext → your handler
```

**Gateway Mode** (`oathmesh serve --gateway`): A reverse proxy that verifies tokens and injects security context headers into your existing upstream services.

---

## Documentation

### Quickstarts
- [Protect a Go chi API](docs/quickstarts/protect-chi-api.md)
- [Protect an Express API](docs/quickstarts/protect-express-api.md)
- [Protect a Next.js API](docs/quickstarts/protect-nextjs-api.md)
- [Protect a FastAPI service](docs/quickstarts/protect-fastapi.md)
- [GitHub Actions to internal API](docs/quickstarts/github-actions-to-internal-api.md)

### Protocol & Security
- [Token Format](docs/protocol/token-format.md) · [Claim Reference](docs/protocol/claim-reference.md)
- [Verification Rules](docs/protocol/verification-rules.md) · [Threat Model](docs/security/threat-model.md)
- [Replay Defense](docs/security/replay-defense.md) · [Key Management](docs/security/key-management.md)

---

## License
[MIT](LICENSE)
