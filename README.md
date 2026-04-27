<p align="center">
  <img src="assets/logo.png" width="200" alt="OathMesh Logo">
</p>

<h1 align="center">OathMesh</h1>

<p align="center">
  <b>🔐 Every machine call gets a short-lived, signed identity.</b>
</p>

<p align="center">
  <img src="assets/social-preview.png" alt="OathMesh in action" width="600">
</p>

<p align="center">
  Stop leaking API keys. Replace static secrets with cryptographically verified tokens that expire in 5 minutes or less.
</p>

> ⚠️ **Pre-production:** OathMesh has not yet received an independent security audit, but it is currently structurally ready for Early Adopter/MVP deployments.

<p align="center">
  <a href="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml">
    <img src="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml/badge.svg" alt="CI Status">
  </a>
  <a href="https://www.npmjs.com/package/@oathmesh/sdk">
    <img src="https://img.shields.io/npm/v/@oathmesh/sdk.svg" alt="npm version">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/releases">
    <img src="https://img.shields.io/pypi/v/oathmesh.svg" alt="pypi version">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/releases">
    <img src="https://img.shields.io/github/v/release/oathmesh/oathmesh" alt="GitHub Release">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/oathmesh/oathmesh.svg" alt="License">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/stargazers">
    <img src="https://img.shields.io/github/stars/oathmesh/oathmesh" alt="Stars">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/graphs/contributors">
    <img src="https://img.shields.io/github/contributors/oathmesh/oathmesh" alt="Contributors">
  </a>
</p>

---

## ✨ Features

- 🔑 **Zero API Keys** — No more long-lived secrets in environment variables
- ⏱️ **Short-Lived Tokens** — Maximum 300 seconds TTL, auto-expiring credentials
- 🛡️ **Zero-Trust Security** — Every request must prove its identity
- 🔒 **Ed25519 Signatures** — Modern elliptic curve cryptography (KMS-backed via `SignFunc` abstraction)
- 📋 **14-Step Verification** — Func-slice pipeline with step-annotated errors for instant diagnosis
- 🌐 **Polyglot SDKs** — Go, Node.js (TypeScript), and Python supported
- 📊 **Full Audit Trail** — Every allow and deny logged via composable `FanOutAuditSink`
- 🔄 **Policy-Driven** — Apple Pkl-based rules, hot-reload, default deny
- 🌐 **Gateway Integrations** — Native Envoy `ext_authz` and Kong Go PDK plugins
- ⚡ **Performance Proven** — Sub-millisecond p99 latency overhead mathematically verified
- 🛠️ **CLI Native** — Terminal-driven management for robust GitOps integration
- 🛑 **Stateful Revocation** — Redis-backed O(1) revocation lists directly synced
- 🤖 **CI Native Auto-Sign** — Built-in OIDC exchange mappings for GitHub Actions and GitLab CI
- ⚡ **Circuit-Breaker Replay Defense** — Redis failover to in-process cache (never fails open)
- 🚀 **JWKS Pre-Warming** — Eliminate cold-start latency with `JWKSCache.PreWarm(ctx)`

---

## 📚 SDKs

| Language | Package | Frameworks |
|----------|---------|------------|
| **Go** | [`github.com/oathmesh/oathmesh`](https://github.com/oathmesh/oathmesh) | chi, stdlib `net/http` |
| **Node.js** | [`@oathmesh/sdk`](https://www.npmjs.com/package/@oathmesh/sdk) | Express, **Next.js** (App, Pages, Edge) |
| **Python** | [`oathmesh`](https://github.com/oathmesh/oathmesh/releases) | FastAPI, Flask, Django |

### SDK Feature Comparison

| Feature | Go SDK | Node.js SDK | Python SDK |
|---------|--------|-------------|------------|
| **Token verification** | ✅ Full 14-step | ✅ Full 14-step (Go-aligned semantics) | ✅ Full 14-step (Go-aligned semantics) |
| **alg:none rejection** | ✅ | ✅ | ✅ |
| **Exact audience match** | ✅ | ✅ | ✅ |
| **Subject format validation** | ✅ | ✅ | ✅ |
| **rqh binding** | ✅ | ✅ | ✅ |
| **Binding-required mode (`rqh`)** | ✅ | ✅ | ✅ |
| **Future `iat` rejection** | ✅ | ✅ | ✅ |
| **Replay cache** | ✅ Built-in | ✅ Built-in (InMemoryReplayCache) | ✅ Built-in (InMemoryReplayCache) |
| **Revocation list (step 13.5)** | ✅ Conformance-covered | ✅ Conformance-covered (InMemory/Redis) | ✅ Conformance-covered (InMemory/Redis) |
| **Fail-closed Caching** | ✅ | ✅ | ✅ |
| **Policy evaluation** | ✅ Built-in (Pkl) | ✅ Built-in (JSON) | ✅ Built-in (JSON) |

> **Conformance note:** Node.js and Python verifiers were tightened toward the canonical Go step semantics (for example: `alg=none` rejection, subject format validation, required request binding semantics, future-`iat` rejection, and fail-closed cache behaviors). This establishes exact behavioral parity.

---

## 🚦 Start Here

Use this canonical developer entry flow:

1. **Step 1 (commands):** [QUICKSTART.md](QUICKSTART.md)
2. **Step 2 (guided onboarding):** [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)
3. **Step 3 (full docs index):** [docs/INDEX.md](docs/INDEX.md)

Step 1 is the canonical runnable path for local verification (issuer `http://localhost:4000`, protected `chi-api` at `http://localhost:8081`).

### ✅ Local quality checks (before a PR)

Run the minimal local quality workflow:

```bash
make quality-local
```

If `make` is not available (for example on some Windows setups), run the same flow manually:

```bash
go test ./...
golangci-lint run ./...  # if installed
govulncheck ./...        # if installed
```

---

## 📦 Installation

### Go

```bash
go install github.com/oathmesh/oathmesh/cmd/oathmesh@latest
```

### Node.js / TypeScript

```bash
npm install @oathmesh/sdk
# or
yarn add @oathmesh/sdk
# or
pnpm add @oathmesh/sdk
```

### Python

```bash
pip install oathmesh
# or
poetry add oathmesh
```

### Docker

```bash
docker pull oathmesh/oathmesh:latest
```

---

## 💻 Usage Examples

### Go Middleware

```go
r.Use(middleware.OathMeshMiddleware(cfg))
caller := middleware.CallerFrom(r.Context())
// caller.Principal.Subject, caller.Action, caller.TokenID
```

### Express (TypeScript)

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

### FastAPI (Python)

```python
from oathmesh import verify_token, VerifierConfig

caller = verify_token(request.headers["authorization"], config)
# caller.principal.subject, caller.action, caller.token_id
```

---

## ⚖️ Comparison

| Feature | API Keys | Traditional JWT | OathMesh |
|---------|----------|------------------|----------|
| **Lifetime** | Infinite (leaked = compromised) | Hours to days | ≤ 300 seconds |
| **Cryptography** | None (just strings) | HS256, RS256 common | Ed25519 only |
| **Replay Protection** | ❌ | ❌ | ✅ Unique `jti` per token |
| **Policy Engine** | ❌ | ❌ | ✅ Pkl-based rules |
| **Audit Logging** | ❌ | Optional | ✅ Every allow/deny |
| **Scoped Actions** | ❌ | Optional | ✅ `act` claim required |

---

## 🏗️ Architecture

```
┌──────────┐    ┌─────────┐    ┌─────────────────────┐
│ Caller   │───▶│ Issuer  │───▶│ Signs Oath Token    │
│ (bot, CI,│    │         │    │ (Ed25519, ≤300s TTL)│
│  service)│    └─────────┘    └─────────────────────┘
└──────────┘                         │
                     ┌────────────────┴────────────┐
                     ▼                             ▼
              ┌──────────────┐              ┌──────────────┐
              │   Receiver   │              │   Gateway    │
              │  (your API)  │              │ (proxy mode) │
              └──────────────┘              └──────────────┘
                     │                             │
              ┌──────┴──────┐               ┌───────┴───────┐
              │ 14-step     │               │ Injects       │
              │ verification│               │ X-OathMesh-*  │
              │ pipeline    │               │ headers       │
              └─────────────┘               └───────────────┘
```

**Gateway Mode** (`oathmesh serve --gateway`): A reverse proxy that verifies tokens and injects security context headers into your existing upstream services.

---

## 🗺️ Roadmap

- 🔜 **Rust SDK** — Coming soon
- 🔜 **Java SDK** — Coming soon  
- 🗓️ **Policy UI** — Visual policy editor
- 🗓️ **Audit Dashboard** — Web-based log viewer

---

## 🤝 Contributing

Contributions are welcome! Here's how to get started:

1. **Fork** the repository
2. **Clone** your fork: `git clone https://github.com/YOUR_USERNAME/oathmesh.git`
3. **Create a branch**: `git checkout -b feature/your-feature-name`
4. **Make your changes** and add tests
5. **Run tests**: `make test` (Go) / `npm test` (Node) / `pytest` (Python)
6. **Submit a PR** — We'll review and merge!

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## ⭐ Show Your Support

If OathMesh helps you build safer systems, please:

- **Star** this repository ⭐
- **Share** it with your team
- **Open an issue** if you find a bug or have a feature request
- **Contribute** — We need SDKs for more languages!

[![Star this repo](https://img.shields.io/github/stars/oathmesh/oathmesh?style=social)](https://github.com/oathmesh/oathmesh)

---

## 📖 Documentation

### Quickstarts
- [Protect a Go chi API](docs/quickstarts/protect-chi-api.md)
- [Protect an Express API](docs/quickstarts/protect-express-api.md)
- [Protect a Next.js API](docs/quickstarts/protect-nextjs-api.md)
- [Protect a FastAPI service](docs/quickstarts/protect-fastapi.md)
- [GitHub Actions to internal API](docs/quickstarts/github-actions-to-internal-api.md)

### Tutorials
- [Getting started: issuer + receiver + verify](docs/tutorials/getting-started.md)
- [gRPC integration](docs/tutorials/grpc-integration.md)
- [GraphQL integration (Node + Python)](docs/tutorials/graphql-integration.md)
- [CI/CD machine identity](docs/tutorials/ci-cd-machine-identity.md)

### Deployment
- [Linux VM Deployment (systemd)](docs/deployment/vm.md)
- [Docker Compose Deployment](docs/deployment/docker-compose.md)
- [Kubernetes Deployment Guide](docs/deployment/kubernetes.md)
- [TLS Configuration Guide](docs/deployment/tls.md)

### Protocol & Security
- [Token Format](docs/protocol/token-format.md) · [Claim Reference](docs/protocol/claim-reference.md)
- [Verification Rules](docs/protocol/verification-rules.md) · [Threat Model](docs/security/threat-model.md)
- [Replay Defense](docs/security/replay-defense.md) · [Key Management](docs/security/key-management.md)
- [SOC2 Compliance Matrix](docs/security/soc2-compliance.md)

### Policy
- [Policy Overview](docs/policies/overview.md)
- [Policy Examples](docs/policies/examples.md)
- [Policy Migration Guide](docs/policies/migration.md)

### Integrations
- [Envoy `ext_authz` Integration](docs/integrations/envoy.md)
- [Kong Go PDK Plugin](docs/integrations/kong.md)

### Performance
- [Kubernetes Zero-Trust Benchmarks](docs/PERFORMANCE.md)

### Documentation Hub
- [Full Documentation Index](docs/INDEX.md)

---

## 🔒 Security

For security vulnerabilities, please see [SECURITY.md](SECURITY.md). **Do NOT open a public issue** for security vulnerabilities.

---

## 📄 License

[MIT](LICENSE)

---

<p align="center">
  Built with ❤️ by the <a href="https://github.com/oathmesh/oathmesh/graphs/contributors">OathMesh team</a>
</p>
