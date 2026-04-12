# Quickstart: Local Demo with Docker Compose

**Time:** ~10 minutes

This guide gets you from zero to a fully working OathMesh demo using Docker Compose.

## Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for building the CLI)

## Step 1: Clone and Build

```bash
git clone https://github.com/oathmesh/oathmesh.git
cd oathmesh

# Generate a development key
openssl genpkey -algorithm Ed25519 -out private.pem

# Build the CLI
go build -o bin/oathmesh ./cmd/oathmesh
```

## Step 2: Start Services

```bash
docker-compose up -d
```

This starts:
- **Issuer** on port 4000 — mints and signs Oath Tokens
- **chi-api** on port 8080 — example inventory API protected by OathMesh middleware
- **Redis** — replay cache backend
- **upstream-demo** — mock upstream for gateway mode

Wait for health checks:
```bash
curl http://localhost:4000/healthz   # Should return "OK"
curl http://localhost:8080/healthz   # Should return "OK"
```

## Step 3: Mint a Token

```bash
export OATHMESH_ISSUER=http://localhost:4000
export OATHMESH_PRIVATE_KEY_FILE=./private.pem

TOKEN=$(./bin/oathmesh mint \
  --sub "agent://repo/acme/deploy-bot" \
  --aud "https://inventory.internal" \
  --act "deploy" \
  --quiet)

echo "Token: $TOKEN"
```

## Step 4: Call the Protected API

```bash
curl -s -H "Authorization: OathMesh $TOKEN" \
  http://localhost:8080/inventory | jq .
```

Expected output:
```json
{
  "status": "success",
  "data": ["item1", "item2"],
  "caller": {
    "principal": { "issuer": "...", "subject": "agent://repo/acme/deploy-bot" },
    "action": "deploy",
    "tokenId": "..."
  }
}
```

## Step 5: See Failures in Action

**Replay detection** — send the same token again:
```bash
curl -s -H "Authorization: OathMesh $TOKEN" http://localhost:8080/inventory
# → 401 {"error":"replay_detected"}
```

**Wrong audience:**
```bash
BAD=$(./bin/oathmesh mint --sub "agent://repo/acme/bot" \
  --aud "https://other.internal" --act "read" --quiet)
curl -s -H "Authorization: OathMesh $BAD" http://localhost:8080/inventory
# → 401 {"error":"audience_mismatch"}
```

**Expired token:**
```bash
SHORT=$(./bin/oathmesh mint --sub "agent://repo/acme/bot" \
  --aud "https://inventory.internal" --act "read" --ttl 1 --quiet)
sleep 12  # Wait past the 10s clock skew tolerance
curl -s -H "Authorization: OathMesh $SHORT" http://localhost:8080/inventory
# → 401 {"error":"token_expired"}
```

## Step 6: Run the Automated Demo

```bash
./demo.sh
```

This runs all of the above steps automatically and validates each outcome.

## Cleanup

```bash
docker-compose down -v
```
