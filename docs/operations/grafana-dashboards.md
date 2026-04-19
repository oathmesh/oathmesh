← [Back to Index](../INDEX.md)

# OathMesh Grafana Dashboards

This guide covers importing and using the Phase 2 OathMesh Grafana dashboards:

- `monitoring/dashboards/operations.json`
- `monitoring/dashboards/performance.json`
- `monitoring/dashboards/security.json`

These dashboards only use metrics that exist in `internal/metrics/metrics.go` plus standard Prometheus `up`.

## Prerequisites

1. Prometheus is scraping the OathMesh `/metrics` endpoint.
2. Grafana has a Prometheus data source.
3. Prometheus scrape config includes OathMesh targets (job name matching `oathmesh.*` is recommended for health panels).

## Import Dashboards

For each JSON file:

1. In Grafana, go to **Dashboards → New → Import**.
2. Upload one file from `monitoring/dashboards/`.
3. When prompted, map the `datasource` variable to your Prometheus source.
4. Click **Import**.

## Dashboard Contents

## 1) Operations Dashboard

Focus: operational health and reliability.

Panels include:

- OathMesh scrape health (`up{job=~"oathmesh.*"}`)
- Revocation sync errors in last 5m (`increase(oathmesh_revocation_sync_errors[5m])`)
- Verification outcome rates (success, errors, policy denials, replay detects)
- Verification error ratio %
- Traffic rates (gateway, gRPC, token mint)
- Replay cache size
- gRPC requests by status

## 2) Performance Dashboard

Focus: latency and throughput.

Panels include:

- HTTP p50/p95/p99 latency (`oathmesh_request_duration_seconds_bucket`)
- HTTP throughput by method/route (`oathmesh_request_duration_seconds_count`)
- gRPC p95/p99 latency (`oathmesh_grpc_request_duration_seconds_bucket`)
- gRPC throughput by method (`oathmesh_grpc_requests_total`)
- Core throughput (verifications, token mint, gateway traffic)
- Average HTTP latency by route (`sum(rate(..._sum))/sum(rate(..._count))`)

## 3) Security Dashboard

Focus: verification failures and revocation/replay signals.

Panels include:

- Verification errors, policy denials, replay detections (rates)
- Verification failure ratio %
- Revocation sync errors in last 15m
- Issuer/gateway scrape health (`up`)
- gRPC status distribution
- Verification attempts vs rejects
- Replay cache size (context for replay behavior)

## Metric Gaps (Known)

The following signals are referenced in some docs/runbooks but are not currently exported in `internal/metrics`:

- `oathmesh_issuer_uptime_seconds` (not implemented)
- `oathmesh_issuer_health_check` (not implemented as a metric; dashboard uses `up` instead)
- `oathmesh_http_requests_duration_seconds` (docs name mismatch; implemented metric is `oathmesh_request_duration_seconds`)

Additional useful gaps:

- No revocation sync success counter/last-success timestamp metric.
- Verification errors are not labeled by reason/code, so dashboards cannot break down failures by exact cause using metrics alone.

## Notes

- If your Prometheus job name is not `oathmesh...`, update the `up{job=~"oathmesh.*"}` queries after import.
- Dashboards are intentionally query-only and importable as JSON without requiring custom plugins.
