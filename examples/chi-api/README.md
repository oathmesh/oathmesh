# chi-api Example

OathMesh-protected Go chi API using the `sdk/go/middleware` package.

## Run (Standalone)

```bash
cd examples/chi-api
OATHMESH_AUDIENCE=https://inventory.internal \
OATHMESH_TRUSTED_ISSUERS=http://localhost:4000 \
go run main.go
```

## Run (Docker)

```bash
# From repo root
docker-compose up chi-api
```

## Test

```bash
TOKEN=$(oathmesh mint --sub "agent://repo/acme/bot" \
  --aud "https://inventory.internal" --act "deploy" --quiet)

curl -H "Authorization: OathMesh $TOKEN" http://localhost:8080/inventory
```
