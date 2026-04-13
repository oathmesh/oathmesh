# Next.js API Example

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  OathMesh-protected Next.js API showing all integration patterns.
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

## Integration Patterns

| Pattern | File | Use When |
|---------|------|----------|
| **App Router** | `app/api/inventory/route.ts` | New Next.js 13+ projects |
| **Edge Middleware** | `middleware.ts` | Protect all `/api/*` at the edge |
| **Pages Router** | `pages/api/legacy.ts` | Existing `pages/` projects |

## Run Standalone

```bash
cd examples/nextjs-api
npm install

OATHMESH_AUDIENCE=https://inventory.internal \
OATHMESH_TRUSTED_ISSUERS=http://localhost:4000 \
npm run dev
```

The API runs at `http://localhost:3001`

## Run with Docker

```bash
# From repo root
docker-compose up nextjs-api
```

The API runs at `http://localhost:3001`

## Test

```bash
# Mint a token (from repo root)
TOKEN=$(./bin/oathmesh mint \
  --sub "svc://frontend/web" \
  --aud "https://inventory.internal" \
  --act "read" --quiet)

# App Router endpoint
curl -H "Authorization: OathMesh $TOKEN" http://localhost:3001/api/inventory

# Pages Router endpoint (legacy)
curl -H "Authorization: OathMesh $TOKEN" http://localhost:3001/api/legacy
```

**Expected response:**
```json
{"subject":"svc://frontend/web","action":"read"}
```

## Endpoints

| Endpoint | Method | Auth Required | Pattern |
|----------|--------|---------------|---------|
| `/api/inventory` | GET | ✅ | App Router |
| `/api/legacy` | GET | ✅ | Pages Router |
| `/*` | * | Edge middleware | Edge |

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
