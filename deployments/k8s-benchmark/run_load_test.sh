#!/bin/bash
set -e

echo "🚀 Starting OathMesh Kubernetes Benchmark..."

# Check requirements
if ! command -v k6 &> /dev/null; then
    echo "❌ k6 is not installed. Please install k6 (https://k6.io/docs/getting-started/installation/)."
    exit 1
fi
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed."
    exit 1
fi

echo "📦 Deploying benchmarking environment to Kubernetes..."
kubectl apply -f deploy.yaml
kubectl wait --for=condition=ready pod -l app=oathmesh-gateway -n oathmesh-benchmark --timeout=120s
kubectl wait --for=condition=ready pod -l app=upstream-service -n oathmesh-benchmark --timeout=120s
kubectl wait --for=condition=ready pod -l app=prometheus -n oathmesh-benchmark --timeout=120s

# Setup port-forwarding in the background
echo "🔌 Setting up port-forwarding..."
kubectl port-forward svc/oathmesh-gateway 8080:8080 -n oathmesh-benchmark > /dev/null 2>&1 &
GW_PID=$!
kubectl port-forward svc/upstream-service 8081:8080 -n oathmesh-benchmark > /dev/null 2>&1 &
UP_PID=$!
kubectl port-forward svc/oathmesh-issuer 4000:4000 -n oathmesh-benchmark > /dev/null 2>&1 &
ISS_PID=$!
kubectl port-forward svc/prometheus 9090:9090 -n oathmesh-benchmark > /dev/null 2>&1 &
PROM_PID=$!

sleep 5 # Wait for port-forwards to initialize

# Generate a test token
echo "🔑 Generating OathMesh Token for Load Testing..."
TOKEN=$(curl -s -X POST http://localhost:4000/api/v1/tokens \
  -H "Content-Type: application/json" \
  -d '{"subject":"svc://k6-load-tester", "audience":"http://upstream-service.local", "action":"benchmark"}' | jq -r .token)

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Failed to generate token. Is the issuer running?"
    kill $GW_PID $UP_PID $ISS_PID $PROM_PID
    exit 1
fi

# Create a k6 script dynamically
cat <<EOF > k6_script.js
import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    direct: {
      executor: 'constant-vus',
      vus: 50,
      duration: '30s',
      exec: 'directRequest',
    },
    proxied: {
      executor: 'constant-vus',
      vus: 50,
      duration: '30s',
      startTime: '35s',
      exec: 'proxiedRequest',
    },
  },
};

const TOKEN = "${TOKEN}";

export function directRequest() {
  const res = http.get('http://localhost:8081/');
  check(res, { 'status was 200': (r) => r.status == 200 });
}

export function proxiedRequest() {
  const res = http.get('http://localhost:8080/', {
    headers: { 'Authorization': 'OathMesh ' + TOKEN },
  });
  check(res, { 'status was 200': (r) => r.status == 200 });
}
EOF

echo "🔥 Running Load Tests with k6..."
k6 run k6_script.js > k6_results.txt

echo "📊 Fetching Prometheus Server-Side Metrics..."
# We query Prometheus for the 99th percentile of the OathMesh verification duration
PROM_QUERY='histogram_quantile(0.99, sum(rate(oathmesh_verification_duration_seconds_bucket[1m])) by (le))'
P99_VERIFY=$(curl -s -g "http://localhost:9090/api/v1/query?query=${PROM_QUERY}" | jq -r '.data.result[0].value[1]')

echo "=========================================="
echo "🎯 BENCHMARK RESULTS"
echo "=========================================="
cat k6_results.txt | grep "http_req_duration"
echo "------------------------------------------"
echo "Server-Side OathMesh p99 Verification Latency: $(echo $P99_VERIFY | awk '{printf "%.4f", $1 * 1000}') ms"
echo "=========================================="

echo "🧹 Cleaning up..."
kill $GW_PID $UP_PID $ISS_PID $PROM_PID
rm k6_script.js
kubectl delete namespace oathmesh-benchmark
echo "✅ Done."
