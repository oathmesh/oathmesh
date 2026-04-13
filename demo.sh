#!/bin/bash
set -eo pipefail

START_TIME=$SECONDS

echo "============================================="
echo "OathMesh Phase 6: End-to-End Demo Execution"
echo "============================================="

# 1. Start Services Cleanly
echo "[1/6] Starting services from cold state..."
docker-compose down -v
docker-compose up -d --build

# Wait for healthchecks to complete
echo "Waiting for issuer and chi-api to become healthy..."

# Poll issuer health (port 4000)
until curl -sf http://localhost:4000/healthz > /dev/null 2>&1; do
    echo "  Waiting for issuer..." && sleep 2
done
echo "Issuer is healthy."

# Poll chi-api health (port 8080 - the upstream demo service)
until curl -sf http://localhost:8080 > /dev/null 2>&1; do
    echo "  Waiting for chi-api..." && sleep 2
done
echo "chi-api is healthy."

# Export paths mapping so we can use local CLI binary which is compiled
go build -o ./bin/oathmesh ./cmd/oathmesh
CLI="./bin/oathmesh"

# 2. Golden Path Minting
echo "[2/6] Minting valid token..."
export OATHMESH_ISSUER="http://localhost:4000"
export OATHMESH_PRIVATE_KEY_FILE="./private.pem"

# Generate local key if missing
if [ ! -f "private.pem" ]; then
    openssl genpkey -algorithm Ed25519 -out private.pem
fi

TOKEN=$($CLI mint \
  --sub "agent://repo/acme/deploy-bot" \
  --aud "https://inventory.internal" \
  --act "deploy" \
  --ttl 300 --quiet)  # Max TTL is 300s. Values above this are clamped silently.

echo "✅ Token Minted successfully."

# 3. Call Inventory API
echo "[3/6] Calling chi-api Inventory API..."
API_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -H "Authorization: OathMesh $TOKEN" "http://localhost:8080/inventory")
HTTP_STATUS=$(echo "$API_RESP" | grep HTTP_STATUS | awk -F: '{print $2}')

if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ Success (200 OK)."
else
    echo "❌ Expected 200 OK, got $HTTP_STATUS"
    exit 1
fi

# 4. Demonstrate Replay Detection
echo "[4/6] Demonstrating Replay Protection..."
# Send the exact SAME token again
REPLAY_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -H "Authorization: OathMesh $TOKEN" "http://localhost:8080/inventory")
REPLAY_STATUS=$(echo "$REPLAY_RESP" | grep HTTP_STATUS | awk -F: '{print $2}')
REPLAY_BODY=$(echo "$REPLAY_RESP" | sed 's/HTTP_STATUS:.*//')

if [ "$REPLAY_STATUS" -eq 401 ] && echo "$REPLAY_BODY" | grep -q "replay_detected"; then
    echo "✅ Replay protection worked. Received 401 replay_detected."
else
    echo "❌ Replay test failed to block correctly. Status: $REPLAY_STATUS"
    exit 1
fi

# 5. Demonstrate Wrong Audience
echo "[5/6] Demonstrating Incorrect Audience..."
BAD_AUD_TOKEN=$($CLI mint \
  --sub "agent://repo/acme/deploy-bot" \
  --aud "https://other.internal" \
  --act "deploy" \
  --ttl 300 --quiet)  # Max TTL is 300s. Values above this are clamped silently.

AUD_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -H "Authorization: OathMesh $BAD_AUD_TOKEN" "http://localhost:8080/inventory")
AUD_STATUS=$(echo "$AUD_RESP" | grep HTTP_STATUS | awk -F: '{print $2}')
AUD_BODY=$(echo "$AUD_RESP" | sed 's/HTTP_STATUS:.*//')

if [ "$AUD_STATUS" -eq 401 ] && echo "$AUD_BODY" | grep -q "audience_mismatch"; then
    echo "✅ Audience mismatch worked. Received 401 audience_mismatch."
else
    echo "❌ Audience test failed. Status: $AUD_STATUS"
    exit 1
fi

# 6. Demonstrate Expiry
echo "[6/6] Demonstrating Token Expiry..."
EXP_TOKEN=$($CLI mint \
  --sub "agent://repo/acme/deploy-bot" \
  --aud "https://inventory.internal" \
  --act "deploy" \
  --ttl 1 --quiet)

echo "Waiting 12 seconds for token to formally expire (bypassing 10s clock skew window)..."
sleep 12

EXP_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -H "Authorization: OathMesh $EXP_TOKEN" "http://localhost:8080/inventory")
EXP_STATUS=$(echo "$EXP_RESP" | grep HTTP_STATUS | awk -F: '{print $2}')
EXP_BODY=$(echo "$EXP_RESP" | sed 's/HTTP_STATUS:.*//')

if [ "$EXP_STATUS" -eq 401 ] && echo "$EXP_BODY" | grep -q "token_expired"; then
    echo "✅ Expiry check worked. Received 401 token_expired."
else
    echo "❌ Expiry test failed. Status: $EXP_STATUS Body: $EXP_BODY"
    exit 1
fi

echo "============================================="
echo "All End-to-End validions passed."
END_TIME=$SECONDS
DURATION=$((END_TIME - START_TIME))
echo "Total runtime: $DURATION seconds."

if [ $DURATION -lt 120 ]; then
   echo "Runtime assertion passed (Under 2 minutes)."
else
   echo "Warning: Demo run took $DURATION seconds (exceeds 120s goal)."
fi

# Clean up
docker-compose down
