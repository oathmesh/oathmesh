# Express API Example

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  OathMesh-protected Express.js API using @oathmesh/oathmesh.
</p>

<p align="center">
  <a href="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml">
    <img src="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml/badge.svg" alt="CI Status">
  </a>
</p>

---

## Prerequisites

- Node.js 18+
- Docker & Docker Compose (for Docker run option)
- Running OathMesh issuer (see main README)

## Run Standalone

```bash
cd examples/express-api
npm install

OATHMESH_AUDIENCE=https://inventory.internal \
OATHMESH_TRUSTED_ISSUERS=http://localhost:4000 \
npx ts-node index.ts
```

The API runs at `http://localhost:3000`

## Run with Docker

```bash
# From repo root
docker-compose up express-api
```

The API runs at `http://localhost:3000`

## Test

```bash
# Mint a token (from repo root)
TOKEN=$(./bin/oathmesh mint \
  --sub "svc://frontend/web" \
  --aud "https://inventory.internal" \
  --act "read" --quiet)

# Call the protected endpoint
curl -H "Authorization: OathMesh $TOKEN" http://localhost:3000/inventory
```

**Expected response:**
```json
{"subject":"svc://frontend/web","action":"read"}
```

## Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/healthz` | GET | ❌ | Health check |
| `/inventory` | GET | ✅ | Protected inventory |

## Cleanup

```bash
# Standalone: Ctrl+C

# Docker
docker-compose down
```

---

<p align="center">
  <sub>See <a href="../../sdk/node/README.md">Node SDK</a> for middleware docs</sub>
</p>
