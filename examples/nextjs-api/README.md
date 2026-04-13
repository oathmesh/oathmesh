# Next.js API Example

OathMesh-protected Next.js API using `@oathmesh/sdk/next`.

Shows all three integration patterns:

| Pattern | File | Use When |
|---|---|---|
| **App Router** | `app/api/inventory/route.ts` | New Next.js 13+ projects |
| **Edge Middleware** | `middleware.ts` | Protect all `/api/*` at the edge |
| **Pages Router** | `pages/api/legacy.ts` | Existing `pages/` projects |

## Run

```bash
cd examples/nextjs-api
npm install
OATHMESH_AUDIENCE=https://inventory.internal \
OATHMESH_TRUSTED_ISSUERS=http://localhost:4000 \
npm run dev
```

## Test

```bash
TOKEN=$(oathmesh mint --sub "svc://frontend/web" \
  --aud "https://inventory.internal" --act "read" --quiet)

# App Router endpoint
curl -H "Authorization: OathMesh $TOKEN" http://localhost:3001/api/inventory

# Pages Router endpoint (legacy)
curl -H "Authorization: OathMesh $TOKEN" http://localhost:3001/api/legacy
```
