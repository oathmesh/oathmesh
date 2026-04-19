# Quick Start

> 📖 **For detailed, step-by-step guidance:** See [Getting Started](docs/GETTING_STARTED.md) in the docs folder. That guide will walk you through different paths (local demo, protecting your own API, or learning concepts first).

Get from zero to a protected API call in 5 commands.

## Prerequisites

- Go 1.23+
- Docker & Docker Compose
- openssl

## 5 Commands to Protected API

```bash
# 1. Clone and enter the repo
git clone https://github.com/oathmesh/oathmesh.git
cd oathmesh

# 2. Generate a development key
openssl genpkey -algorithm Ed25519 -out private.pem

# 3. Start services (issuer + demo API)
docker-compose up -d --build

# 4. Mint a token (max TTL 300s)
TOKEN=$(./bin/oathmesh mint \
  --sub "agent://repo/acme/deploy-bot" \
  --aud "https://inventory.internal" \
  --act "deploy" \
  --quiet)

# 5. Call the protected API
curl -H "Authorization: OathMesh $TOKEN" http://localhost:8080/inventory
```

Expected response:
```json
{"subject":"agent://repo/acme/deploy-bot","action":"deploy"}
```

## What Just Happened?

1. **Generated an Ed25519 key** — Your machine identity signing key
2. **Started the issuer** — Serves tokens at `http://localhost:4000`
3. **Started a demo API** — Protected by OathMesh middleware at `http://localhost:8080`
4. **Minted a token** — 300-second max TTL, cryptographically signed
5. **Called the API** — Token verified, identity extracted, request allowed

## Next Steps

- Read the full [Architecture documentation](ARCHITECTURE.md)
- Explore [SDK documentation](sdk/)
- Learn about [verification rules](docs/protocol/verification-rules.md)
- Set up [production deployment](docs/deployment/)

## Troubleshooting

**Port already in use?** Stop existing containers: `docker-compose down`

**Token expired?** Mint a new one — tokens max out at 300 seconds

**Want to run without Docker?** See [From Source](README.md#option-2-from-source)