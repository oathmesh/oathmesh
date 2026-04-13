# OathMesh

**Every machine call gets a short-lived, signed identity.**

OathMesh replaces shared secrets — API keys, static tokens, long-lived credentials — with scoped, verifiable, auditable call assertions for services, agents, CI/CD jobs, and tools.

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
| **Go** | `sdk/go/middleware` | chi, stdlib `net/http` |
| **Node.js / TypeScript** | `@oathmesh/sdk` | Express, **Next.js** (App Router, Pages Router, Edge Middleware) |
| **Python** | `oathmesh` | FastAPI, Flask, Django |

### Go

```go
r.Use(middleware.OathMeshMiddleware(cfg))
caller := middleware.CallerFrom(r.Context())
```

### Express

```typescript
import { verifyToken } from '@oathmesh/sdk';
app.use(verifyToken({ audience, trustedIssuers }));
// req.oathmeshContext is fully typed
```

### Next.js (App Router)

```typescript
import { withOathMesh } from '@oathmesh/sdk/next';
const oathmesh = withOathMesh({ audience, trustedIssuers });

export async function GET(request: NextRequest) {
  const { caller, error } = await oathmesh(request);
  if (error) return error;
  return NextResponse.json({ subject: caller.principal.subject });
}
```

### FastAPI

```python
from oathmesh import verify_token, VerifierConfig, OathMeshError
caller = verify_token(request.headers["authorization"], config)
# caller.principal.subject, caller.action, caller.token_id
```

---

## Examples

| Example | Language | Path |
|---|---|---|
| chi API | Go | [`examples/chi-api/`](examples/chi-api/) |
| Express API | TypeScript | [`examples/express-api/`](examples/express-api/) |
| Next.js API | TypeScript | [`examples/nextjs-api/`](examples/nextjs-api/) |
| FastAPI | Python | [`examples/fastapi-api/`](examples/fastapi-api/) |
| GitHub Actions | YAML | [`examples/github-actions/`](examples/github-actions/) |
| curl demo | bash | [`examples/curl/`](examples/curl/) |

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

**Gateway Mode** (`oathmesh serve --gateway`): reverse proxy that verifies tokens, strips the `Authorization` header, and injects `X-OathMesh-*` context headers before forwarding to upstream services.

---

## Documentation

### Quickstarts
- [Protect a Go chi API](docs/quickstarts/protect-chi-api.md)
- [Protect an Express API](docs/quickstarts/protect-express-api.md)
- [Protect a Next.js API](docs/quickstarts/protect-nextjs-api.md)
- [Protect a FastAPI service](docs/quickstarts/protect-fastapi.md)
- [GitHub Actions to internal API](docs/quickstarts/github-actions-to-internal-api.md)
- [Local demo with Docker Compose](docs/quickstarts/local-demo-docker-compose.md)

### Protocol
- [Overview](docs/overview.md) · [Concepts](docs/concepts.md)
- [Token Format](docs/protocol/token-format.md) · [Claim Reference](docs/protocol/claim-reference.md)
- [Verification Rules (14 steps)](docs/protocol/verification-rules.md)
- [Error Taxonomy](docs/protocol/error-taxonomy.md) · [Audit Events](docs/protocol/audit-events.md)

### Configuration
- [Issuer Config](docs/config/issuer-config.md) · [Pkl Policy Guide](docs/config/pkl-policy-guide.md)
- [CLI Reference](docs/cli-reference.md)

### Security
- [Threat Model](docs/security/threat-model.md) · [Key Management](docs/security/key-management.md)
- [Replay Defense](docs/security/replay-defense.md) · [Logging Guidance](docs/security/logging-guidance.md)

### Migration
- [Replace an API Key in one afternoon](docs/migration/replace-api-key.md)

### Architecture
- [ARCHITECTURE.md](ARCHITECTURE.md) · [CONTRIBUTING.md](CONTRIBUTING.md)

---

## Technology Stack

| Concern | Choice |
|---|---|
| Language | Go 1.22+ |
| HTTP framework | chi/v5 |
| Signing | crypto/ed25519 (stdlib) |
| Config DSL | Apple Pkl |
| Audit | NDJSON (stdout / file) |
| Replay cache | In-memory / Redis |

---

## License

[MIT](LICENSE)
