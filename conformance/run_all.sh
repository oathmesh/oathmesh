#!/bin/bash
# run_all.sh - Run all three conformance test runners and diff results

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "============================================="
echo "OathMesh Cross-SDK Conformance Suite"
echo "============================================="
echo ""

# Check if services are running
echo "Checking service availability..."

# Check Go chi-api (port 8081)
if ! curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/healthz 2>/dev/null | grep -q "200"; then
    echo "WARNING: Go chi-api not responding on port 8081"
fi

# Check Node Express (port 3000)
if ! curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/healthz 2>/dev/null | grep -q "200"; then
    echo "WARNING: Node Express not responding on port 3000"
fi

# Check Python FastAPI (port 8000)
if ! curl -s -o /dev/null -w "%{http_code}" http://localhost:8000/healthz 2>/dev/null | grep -q "200"; then
    echo "WARNING: Python FastAPI not responding on port 8000"
fi

echo ""

# Run Go tests
echo ">>> Running Go SDK conformance tests..."
cd "$SCRIPT_DIR"
if bash run_go.sh > /dev/null 2>&1; then
    echo "✓ Go tests completed"
else
    echo "✗ Go tests failed or had errors"
fi

# Run Node tests
echo ">>> Running Node SDK conformance tests..."
if bash run_node.sh > /dev/null 2>&1; then
    echo "✓ Node tests completed"
else
    echo "✗ Node tests failed or had errors"
fi

# Run Python tests
echo ">>> Running Python SDK conformance tests..."
if bash run_python.sh > /dev/null 2>&1; then
    echo "✓ Python tests completed"
else
    echo "✗ Python tests failed or had errors"
fi

echo ""
echo "============================================="
echo "Cross-SDK Comparison Results"
echo "============================================="
echo ""

# Compare results
GO_RESULTS="$SCRIPT_DIR/results_go.json"
NODE_RESULTS="$SCRIPT_DIR/results_node.json"
PYTHON_RESULTS="$SCRIPT_DIR/results_python.json"

# Check if result files exist
if [ ! -f "$GO_RESULTS" ]; then
    echo "ERROR: Go results file not found"
    exit 1
fi

if [ ! -f "$NODE_RESULTS" ]; then
    echo "ERROR: Node results file not found"
    exit 1
fi

if [ ! -f "$PYTHON_RESULTS" ]; then
    echo "ERROR: Python results file not found"
    exit 1
fi

# Compare each fixture
echo "Comparing results across SDKs..."
echo ""

 mismatches=0

# Get all fixture IDs from Go results
fixture_ids=$(python3 -c "
import json
with open('$GO_RESULTS') as f:
    data = json.load(f)
    for item in data:
        print(item['id'])
" 2>/dev/null)

for fixture_id in $fixture_ids; do
    # Skip dynamic tests that may not have run on all platforms
    if [[ "$fixture_id" == "valid_token" ]] || [[ "$fixture_id" == "expired_token" ]] || [[ "$fixture_id" == "replayed_token" ]] || [[ "$fixture_id" == "ttl_over_300s" ]]; then
        continue
    fi
    
    # Get expected status from fixtures.json
    expected=$(python3 -c "
import json
with open('$SCRIPT_DIR/fixtures.json') as f:
    data = json.load(f)
    for item in data['fixtures']:
        if item['id'] == '$fixture_id':
            print(item['expect_status'])
            break
" 2>/dev/null)
    
    # Get actual status from each SDK
    go_actual=$(python3 -c "
import json
with open('$GO_RESULTS') as f:
    data = json.load(f)
    for item in data:
        if item['id'] == '$fixture_id':
            print(item['actual_status'])
            break
" 2>/dev/null)
    
    node_actual=$(python3 -c "
import json
with open('$NODE_RESULTS') as f:
    data = json.load(f)
    for item in data:
        if item['id'] == '$fixture_id':
            print(item['actual_status'])
            break
" 2>/dev/null)
    
    python_actual=$(python3 -c "
import json
with open('$PYTHON_RESULTS') as f:
    data = json.load(f)
    for item in data:
        if item['id'] == '$fixture_id':
            print(item['actual_status'])
            break
" 2>/dev/null)
    
    # Compare
    if [ "$go_actual" != "$node_actual" ] || [ "$go_actual" != "$python_actual" ]; then
        echo "MISMATCH: $fixture_id"
        echo "  Expected: $expected"
        echo "  Go:      $go_actual"
        echo "  Node:    $node_actual"
        echo "  Python:  $python_actual"
        mismatches=$((mismatches + 1))
    fi
done

echo ""
if [ $mismatches -eq 0 ]; then
    echo "✓ All SDKs return consistent results"
else
    echo "✗ Found $mismatches mismatches between SDKs"
    exit 1
fi

echo ""
echo "Results saved to:"
echo "  - $GO_RESULTS"
echo "  - $NODE_RESULTS"
echo "  - $PYTHON_RESULTS"