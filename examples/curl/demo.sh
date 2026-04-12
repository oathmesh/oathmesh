#!/bin/bash
set -eo pipefail

ISSUER_URL=${ISSUER_URL:-"http://localhost:4000"}
API_URL=${API_URL:-"http://localhost:8080"}

echo "----------------------------------------"
echo "OathMesh Direct Curl CLI Example"
echo "----------------------------------------"

echo "[1/3] Minting Token directly via Curl..."
RESPONSE=$(curl -s -X POST "$ISSUER_URL/v1/token" \
  -H "Content-Type: application/json" \
  -d '{
    "sub": "script://examples/curl/demo",
    "aud": "https://inventory.internal",
    "act": "read",
    "ttl": 3600
  }')

TOKEN=$(echo $RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ Failed to mint token. Is the issuer running?"
    exit 1
fi

echo "✅ Token Minted."

echo "[2/3] Calling API..."
API_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -H "Authorization: OathMesh $TOKEN" "$API_URL/inventory")

HTTP_STATUS=$(echo "$API_RESP" | grep HTTP_STATUS | awk -F: '{print $2}')
BODY=$(echo "$API_RESP" | sed 's/HTTP_STATUS:.*//')

if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ API Request Succeeded (200 OK)"
    echo "$BODY" | jq .
else
    echo "❌ API Request Failed ($HTTP_STATUS)"
    echo "$BODY"
    exit 1
fi

echo "[3/3] Success. Check issuer logs to see the Audit event emitted."
