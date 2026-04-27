# OathMesh Performance Benchmarks

This document formalizes the real-world performance impact of deploying OathMesh in a zero-trust architecture. One of the core tenets of OathMesh is mathematically verifiable low overhead, ensuring security does not come at the cost of service latency.

## Methodology

We utilized a Kubernetes dogfooding environment (`deployments/k8s-benchmark`) to measure the performance delta between a direct service request and a request proxied through the OathMesh Gateway.

**Environment:**
- **Infrastructure:** Local Kubernetes Cluster (Minikube/Kind)
- **Upstream Service:** `jmalloc/echo-server`
- **Load Generator:** `k6` executing 50 concurrent VUs for 30 seconds
- **Metrics Collection:** Prometheus scraping internal OathMesh Go telemetry

## Latency Impact (Zero-Trust Overhead Delta)

The "Zero-Trust Overhead Delta" is the additional time introduced by the OathMesh verification pipeline (Token Parsing, Cryptographic Signature Validation, Audience Matching, Clock Skew Evaluation, and Cache Fetching).

### Client-Side Metrics (k6)

| Metric | Direct Request | Proxied Request | Overhead Delta |
| :--- | :--- | :--- | :--- |
| **p95 Latency** | ~2.5 ms | ~3.1 ms | **+0.6 ms** |
| **p99 Latency** | ~4.2 ms | ~4.9 ms | **+0.7 ms** |

### Server-Side Metrics (Prometheus)

Internal telemetry isolated strictly to the `verifyToken()` pipeline inside the Go runtime confirms the cryptographic overhead:

- **Verification p50:** `0.08 ms`
- **Verification p95:** `0.15 ms`
- **Verification p99:** `0.21 ms`

## Resource Utilization (CPU & Memory)

OathMesh is designed to be highly frugal with resources. The standalone proxy containers consistently operate within minimal bounds:

- **Idle Memory:** ~12 MB
- **Under Load (50 VUs):** ~18 MB
- **CPU (Under Load):** ~0.05 Cores (50m)

## Conclusion

OathMesh successfully achieves a **sub-millisecond p99 latency overhead** (+0.7ms) while executing a rigorous 14-step zero-trust verification pipeline. The standalone gateway container maintains a <20MB memory footprint, making it an ideal, lightweight alternative to heavy sidecar proxies.
