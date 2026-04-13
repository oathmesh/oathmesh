# Express API Example

OathMesh-protected Express.js API using `@oathmesh/sdk`.

## Run

```bash
cd examples/express-api
npm install
OATHMESH_AUDIENCE=https://inventory.internal \
OATHMESH_TRUSTED_ISSUERS=http://localhost:4000 \
npx ts-node index.ts
```

## Test

```bash
TOKEN=$(oathmesh mint --sub "svc://frontend/web" \
  --aud "https://inventory.internal" --act "read" --quiet)

curl -H "Authorization: OathMesh $TOKEN" http://localhost:3000/inventory
```
