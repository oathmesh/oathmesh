# OathMesh

**OathMesh gives every machine call a short-lived signed identity.**

OathMesh is a micro-protocol and developer platform that replaces shared machine secrets (API keys, static tokens) with short-lived, signed JWT-based call identity for agents, CI/CD jobs, internal tools, automation bots, and service-to-service calls.

---

## What Is OathMesh

OathMesh is:
- **A protocol** — a narrow, implementable standard for signed short-lived machine-call assertions
- **A product** — an issuer service, verifier middleware, gateway, SDK set, CLI, and audit pipeline built on that protocol

OathMesh is NOT:
- A user authentication system
- A browser login or OAuth platform for humans
- A service mesh or data plane
- A replacement for cloud IAM
- A replacement for SPIFFE (can run beside it)

---

## Core Doctrine

> **"OathMesh authenticates the caller. The receiver authorizes the request."**

---

## Getting Started

### Quick Start (Local Development)

```bash
# Clone the repository
git clone https://github.com/oathmesh/oathmesh.git
cd oathmesh

# Generate a private key for local development
openssl genpkey -algorithm Ed25519 -out private.pem

# Export the private key
export OATHMESH_PRIVATE_KEY="$(cat private.pem)"

# Start services with Docker Compose
docker-compose up

# In another terminal, mint a token
./bin/oathmesh mint --sub "agent://repo/acme/deploy-bot" \
  --aud "https://inventory.internal" \
  --act "inventory.write"

# Verify the token
echo "<token>" | ./bin/oathmesh verify --audience "https://inventory.internal" \
  --issuer "https://issuer.oathmesh.dev"
```

### Run the Demo

```bash
./demo.sh
```

## Architecture & Gateway Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                     OathMesh System                         │
│                                                             │
│  ┌──────────┐    ┌──────────┐    ┌──────────────────────┐ │
│  │  Caller  │───▶│  Issuer  │    │     Receiver         │ │
│  │ (agent,  │    │ Service  │    │  ┌────────────────┐  │ │
│  │  CI job, │    │          │    │  │   Verifier     │  │ │
│  │  bot)    │◀───│ Mint API │    │  │   Middleware   │  │ │
│  └──────────┘    │ JWKS     │    │  └───────┬────────┘  │ │
│       │          └──────────┘    │  ┌───────▼────────┐  │ │
│       │               │          │  │  Policy Engine │  │ │
│       │          ┌────▼─────┐    │  └───────┬────────┘  │ │
│       └─────────▶│ Gateway  │───▶│  ┌───────▼────────┐  │ │
│                  │ Proxy    │    │  │  Audit Logger  │  │ │
│                  └──────────┘    │  └────────────────┘  │ │
│                                  └──────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

OathMesh natively supports **Gateway Mode** (`oathmesh serve --gateway`), transforming the root issuer server into an intelligent HTTP reverse proxy. This seamlessly intercepts upstream requests, dynamically pulling and verifying distributed tokens from incoming `Authorization` headers, blocking anomalies, injecting HTTP Context Headers (`X-OathMesh-Subject`, `X-OathMesh-Action`), and formally logging audit trails natively.

---

## Polyglot SDK Integrations

OathMesh provides native middleware enforcing verifications for all popular endpoints organically without requiring the Gateway Proxy:

- **Go (chi)**: `sdk/go/middleware`
- **Node.js (TypeScript/Express)**: `@oathmesh/sdk` inside `sdk/node`
- **Python (FastAPI)**: `oathmesh` SDK inside `sdk/python`

All SDKs natively support fetching distributed JWKS, matching audience specifications, verifying explicit timeouts against configurable clock bounds (max 10s variance default), and injecting contextual definitions intuitively back into route logic.

---

## Documentation
- [CLI Reference](docs/cli-reference.md) — mint, verify, inspect, serve, keys rotate
- **Phase 8 (Upcoming Follow-up):**
  - Protocol Reference — Token format, claims, verification rules
  - Security — Threat model, key management, replay defense
  - Integration Quickstarts

---

## Technology Stack

| Concern | Choice |
|---------|--------|
| Language | Go 1.22+ |
| Config DSL | Apple Pkl |
| HTTP framework | chi/v5 |
| Signing | crypto/ed25519 (stdlib) |

---

## Community

- GitHub: https://github.com/oathmesh/oathmesh
- Issues: https://github.com/oathmesh/oathmesh/issues

---

## License

MIT
