# chi-api Example

<p align="center">
  OathMesh-protected Go chi API using the SDK middleware.
</p>

<p align="center">
  <a href="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml">
    <img src="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml/badge.svg" alt="CI Status">
  </a>
</p>

---

## Prerequisites

- Go 1.22+
- Docker & Docker Compose (for Docker run option)
- Running OathMesh issuer (see main README)

## Run Standalone

```bash
cd examples/chi-api
go mod download

OATHMESH_AUDIENCE=https://inventory.internal \
OATHMESH_TRUSTED_ISSUERS=http://localhost:4000 \
go run main.go
```

The API runs at `http://localhost:8080`

## Run with Docker

```bash
# From repo root
docker-compose up chi-api
```

The API runs at `http://localhost:8080`

## Test

```bash
# Mint a token (from repo root)
TOKEN=$(./bin/oathmesh mint \
  --sub "agent://repo/acme/bot" \
  --aud "https://inventory.internal" \
  --act "deploy" --quiet)

# Call the protected endpoint
curl -H "Authorization: OathMesh $TOKEN" http://localhost:8080/inventory
```

**Expected response:**
```json
{"subject":"agent://repo/acme/bot","action":"deploy"}
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
  <sub>See <a href="../../sdk/go/middleware/README.md">Go SDK</a> for middleware docs</sub>
</p>
