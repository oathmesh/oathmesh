#!/bin/bash
# run_python.sh - Conformance test runner for Python FastAPI
# Tests against the FastAPI example service on port 8000

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURES="$SCRIPT_DIR/fixtures.json"
RESULTS="$SCRIPT_DIR/results_python.json"

ISSUER_URL="${OATHMESH_ISSUER:-http://localhost:4000}"
API_URL="${API_URL:-http://localhost:8000}"

echo "=== OathMesh Conformance Tests (Python SDK) ==="
echo "API URL: $API_URL"
echo ""

# Function to test a single token
test_token() {
    local id="$1"
    local token="$2"
    local expected_status="$3"
    local expected_error="$4"
    
    local response
    local http_status
    
    response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
        -H "Authorization: OathMesh $token" \
        "$API_URL/inventory" 2>/dev/null) || true
    
    http_status=$(echo "$response" | grep "HTTP_STATUS" | awk -F: '{print $2}')
    body=$(echo "$response" | sed 's/HTTP_STATUS:.*//')
    
    local actual_error=""
    if [ -n "$body" ]; then
        actual_error=$(echo "$body" | grep -o '"error":"[^"]*"' | cut -d'"' -f4 || echo "")
    fi
    
    echo "{\"id\":\"$id\",\"expected_status\":$expected_status,\"actual_status\":$http_status,\"expected_error\":\"$expected_error\",\"actual_error\":\"$actual_error\"}"
}

# Test static fixtures from JSON
echo "Testing static fixtures..."
results="["
first=true

while IFS= read -r line; do
    id=$(echo "$line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    token=$(echo "$line" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    expect_status=$(echo "$line" | grep -o '"expect_status":[0-9]*' | cut -d':' -f2)
    expect_error=$(echo "$line" | grep -o '"expect_error_code":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$id" ]; then
        result=$(test_token "$id" "$token" "$expect_status" "$expect_error")
        if [ "$first" = true ]; then
            results="$results$result"
            first=false
        else
            results="$results,$result"
        fi
    fi
done < <(cat "$FIXTURES" | grep -o '{[^}]*}')

# Dynamic test: valid_token (mint fresh)
echo "Testing dynamic fixture: valid_token"
export OATHMESH_ISSUER="$ISSUER_URL"
export OATHMESH_PRIVATE_KEY_FILE="$SCRIPT_DIR/test-key.pem"
VALID_TOKEN=$("$SCRIPT_DIR/../../bin/oathmesh" mint \
    --sub "agent://test/svc" \
    --aud "https://inventory.internal" \
    --act "read" \
    --ttl 120 \
    --quiet 2>/dev/null) || VALID_TOKEN=""

if [ -n "$VALID_TOKEN" ]; then
    response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
        -H "Authorization: OathMesh $VALID_TOKEN" \
        "$API_URL/inventory" 2>/dev/null) || true
    http_status=$(echo "$response" | grep "HTTP_STATUS" | awk -F: '{print $2}')
    body=$(echo "$response" | sed 's/HTTP_STATUS:.*//')
    actual_error=$(echo "$body" | grep -o '"error":"[^"]*"' | cut -d'"' -f4 || echo "")
    
    result="{\"id\":\"valid_token\",\"expected_status\":200,\"actual_status\":$http_status,\"expected_error\":null,\"actual_error\":\"$actual_error\"}"
    results="$results,$result"
fi

# Dynamic test: expired_token
echo "Testing dynamic fixture: expired_token"
EXPIRED_TOKEN=$("$SCRIPT_DIR/../../bin/oathmesh" mint \
    --sub "agent://test/svc" \
    --aud "https://inventory.internal" \
    --act "read" \
    --ttl 1 \
    --quiet 2>/dev/null) || EXPIRED_TOKEN=""

if [ -n "$EXPIRED_TOKEN" ]; then
    sleep 12
    response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
        -H "Authorization: OathMesh $EXPIRED_TOKEN" \
        "$API_URL/inventory" 2>/dev/null) || true
    http_status=$(echo "$response" | grep "HTTP_STATUS" | awk -F: '{print $2}')
    body=$(echo "$response" | sed 's/HTTP_STATUS:.*//')
    actual_error=$(echo "$body" | grep -o '"error":"[^"]*"' | cut -d'"' -f4 || echo "")
    
    result="{\"id\":\"expired_token\",\"expected_status\":401,\"actual_status\":$http_status,\"expected_error\":\"token_expired\",\"actual_error\":\"$actual_error\"}"
    results="$results,$result"
fi

# Dynamic test: not_yet_valid_iat - now static fixture in fixtures.json
echo "Testing static fixture: not_yet_valid_iat"

# Dynamic test: replayed_token
echo "Testing dynamic fixture: replayed_token"
REPLAY_TOKEN=$("$SCRIPT_DIR/../../bin/oathmesh" mint \
    --sub "agent://test/svc" \
    --aud "https://inventory.internal" \
    --act "read" \
    --ttl 120 \
    --quiet 2>/dev/null) || REPLAY_TOKEN=""

if [ -n "$REPLAY_TOKEN" ]; then
    # First use - should succeed
    curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: OathMesh $REPLAY_TOKEN" \
        "$API_URL/inventory" > /dev/null 2>&1 || true
    
    # Second use - should fail with replay_detected
    response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
        -H "Authorization: OathMesh $REPLAY_TOKEN" \
        "$API_URL/inventory" 2>/dev/null) || true
    http_status=$(echo "$response" | grep "HTTP_STATUS" | awk -F: '{print $2}')
    body=$(echo "$response" | sed 's/HTTP_STATUS:.*//')
    actual_error=$(echo "$body" | grep -o '"error":"[^"]*"' | cut -d'"' -f4 || echo "")
    
    result="{\"id\":\"replayed_token\",\"expected_status\":401,\"actual_status\":$http_status,\"expected_error\":\"replay_detected\",\"actual_error\":\"$actual_error\"}"
    results="$results,$result"
fi

results="$results]"

echo "$results" | python -m json.tool > "$RESULTS" 2>/dev/null || echo "$results" > "$RESULTS"

echo ""
echo "Results written to: $RESULTS"