← [Back to Index](../INDEX.md)

# Alerting Rules for OathMesh Operations

This document maps production Prometheus alerts to OathMesh metrics and runbooks.
Rule file: `monitoring/alerts/oathmesh-alerts.yml`

## Rule Inventory

| Alert | Category | Severity | Threshold / Trigger | Runbook |
|---|---|---|---|---|
| `OathMeshIssuerDown` | Availability | critical | `up{job=~"oathmesh\|oathmesh-issuer"} == 0` for 2m | [Issuer Service Down](./incident-response.md#incident-issuer-service-down) |
| `OathMeshVerificationErrorRateWarning` | Availability | warning | verification error ratio > 5% for 10m | [High Verification Error Rate](./on-call-runbook.md#issue-4-high-verification-error-rate) |
| `OathMeshVerificationErrorRateCritical` | Availability | critical | verification error ratio > 15% for 5m | [High Verification Error Rate](./on-call-runbook.md#issue-4-high-verification-error-rate) |
| `OathMeshRestartLoopDetected` | Availability | critical | `changes(process_start_time_seconds[15m]) > 2` for 5m | [Issuer Service Down](./incident-response.md#incident-issuer-service-down) |
| `OathMeshPolicyDenialAnomalyWarning` | Security | warning | policy deny ratio > 20% for 15m | [Verification Failures Triage](./on-call-runbook.md#issue-4-high-verification-error-rate) |
| `OathMeshPolicyDenialAnomalyCritical` | Security | critical | policy deny ratio > 40% for 10m | [Incident Response](./incident-response.md) |
| `OathMeshReplayDetectionSpikeWarning` | Security | warning | replay detections > 5/min for 10m | [Private Key / Abuse Investigation](./incident-response.md#incident-private-key-compromised) |
| `OathMeshReplayDetectionSpikeCritical` | Security | critical | replay detections > 20/min for 5m | [Private Key / Abuse Investigation](./incident-response.md#incident-private-key-compromised) |
| `OathMeshIssuanceLatencyP95Warning` | Performance | warning | `POST /v1/token` p95 > 400ms for 10m | [Monitoring Dashboard Reference](./on-call-runbook.md#monitoring-dashboard-quick-reference) |
| `OathMeshIssuanceLatencyP95Critical` | Performance | critical | `POST /v1/token` p95 > 1s for 5m | [Issuer Service Down](./incident-response.md#incident-issuer-service-down) |
| `OathMeshVerificationLatencyP95Warning` | Performance | warning | gRPC request p95 > 500ms for 10m | [Monitoring Dashboard Reference](./on-call-runbook.md#monitoring-dashboard-quick-reference) |
| `OathMeshVerificationLatencyP95Critical` | Performance | critical | gRPC request p95 > 1.2s for 5m | [Issuer Service Down](./incident-response.md#incident-issuer-service-down) |
| `OathMeshRevocationSyncFailure` | Operations | critical | `increase(oathmesh_revocation_sync_errors[5m]) > 0` for 5m | [Revocation Sync Failed](./on-call-runbook.md#issue-3-revocation-sync-failed) |

## Metrics Used (Implemented in code)

- `oathmesh_verifications_total`
- `oathmesh_verification_errors`
- `oathmesh_policy_denials`
- `oathmesh_replays_detected`
- `oathmesh_request_duration_seconds` (histogram)
- `oathmesh_grpc_request_duration_seconds` (histogram)
- `oathmesh_revocation_sync_errors`

Plus standard Prometheus scrape/runtime metrics:
- `up`
- `process_start_time_seconds`

## Gaps / Current Limitations

The following requested operational/security conditions currently have no dedicated metric in repository code, so no direct alert rule is included yet:

1. **Revoked token attempt spikes** (no standalone counter for `ErrSubjectRevoked`)
2. **Audit write failures** (`AuditSink.Emit` errors are intentionally ignored in `verify.emitAudit`)
3. **Config reload issues** (policy watcher emits logs, but no Prometheus counter/gauge)
4. **Cache miss anomalies** (no JWKS/revocation/replay cache miss metric)

## Prometheus Integration

```yaml
rule_files:
  - monitoring/alerts/oathmesh-alerts.yml
```

Ensure scrape config includes the issuer `/metrics` endpoint:

```yaml
scrape_configs:
  - job_name: "oathmesh-issuer"
    static_configs:
      - targets: ["localhost:4000"]
    metrics_path: "/metrics"
```

## Related Docs

- [On-Call Runbook](./on-call-runbook.md)
- [Incident Response Playbook](./incident-response.md)
- [Production Best Practices](./best-practices.md)
