# FastAPI Example

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  OathMesh-protected FastAPI service using the Python SDK.
</p>

<p align="center">
  <a href="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml">
    <img src="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml/badge.svg" alt="CI Status">
  </a>
</p>

---

## Prerequisites

- Python 3.9+
- Docker & Docker Compose (for Docker run option)
- Running OathMesh issuer (see main README)

## Run Standalone

```bash
cd examples/fastapi-api
pip install -r requirements.txt

OATHMESH_AUDIENCE=https://inventory.internal \
OATHMESH_TRUSTED_ISSUERS=http://localhost:4000 \
uvicorn main:app --host 0.0.0.0 --port 8000
```

The API runs at `http://localhost:8000`

## Run with Docker

```bash
# From repo root
docker-compose up fastapi-api
```

The API runs at `http://localhost:8000`

## Test

```bash
# Mint a token (from repo root)
TOKEN=$(./bin/oathmesh mint \
  --sub "job://ci/nightly" \
  --aud "https://inventory.internal" \
  --act "read" --quiet)

# Call the protected endpoint
curl -H "Authorization: OathMesh $TOKEN" http://localhost:8000/inventory
```

**Expected response:**
```json
{"subject":"job://ci/nightly","action":"read"}
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
  <sub>See <a href="../../sdk/python/README.md">Python SDK</a> for middleware docs</sub>
</p>
