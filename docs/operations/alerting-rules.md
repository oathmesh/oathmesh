# Alerting Rules for OathMesh Operations

This document defines the Prometheus alerting rules for OathMesh production monitoring.

## Revocation Sync Failures

**Metric:** `oathmesh_revocation_sync_errors`

**Alert Rule:**
```yaml
- alert: RevocationSyncFailure
  expr: increase(oathmesh_revocation_sync_errors[5m]) > 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Revocation list sync is failing"
    description: |
      Revocation list synchronization has failed {{ $value }} times in the last 5 minutes.
      This indicates the issuer's revocation endpoint is unreachable or returning errors.
      
      Stale revocation lists will continue to be used for token verification, which may
      allow access to revoked subjects. This is a critical issue requiring immediate action.
    runbook: "https://github.com/oathmesh/oathmesh/docs/operations/runbooks.md#revocation-sync-failure"
```

**What This Alerts On:**
- Network connectivity issues between the gateway and issuer
- Issuer's revocation endpoint returning errors (5xx, timeouts)
- Malformed revocation list responses
- Any sync attempt that fails for any reason

**Severity:** **CRITICAL**

**Required Actions:**
1. Check issuer health and service status
2. Verify Redis connectivity (if using Redis backend)
3. Check network connectivity between gateway and issuer
4. Review issuer logs for `/v1/revoked-subjects` endpoint errors
5. If Redis is down: restart Redis and verify sync succeeds
6. If issuer is down: wait for recovery and monitor for sync success

**Root Causes:**
- Issuer service is down or unhealthy
- Redis service is down (for Redis-backed implementations)
- Network connectivity broken between gateway and issuer
- DNS resolution issues
- Firewall blocking the revocation endpoint
- Issuer endpoint timeout (> 5 seconds)

**Related Metrics:**
- `oathmesh_verification_errors` - Token verification failures (may increase if revocation list is stale)
- `oathmesh_gateway_requests_total` - Overall gateway traffic

## Configuration for Prometheus

Add this alert group to your `prometheus.yml` or separate rules file:

```yaml
groups:
  - name: oathmesh.rules
    interval: 30s
    rules:
      - alert: RevocationSyncFailure
        expr: increase(oathmesh_revocation_sync_errors[5m]) > 0
        for: 5m
        labels:
          severity: critical
          component: revocation
        annotations:
          summary: "OathMesh revocation sync failed: {{ $labels.instance }}"
          description: "Revocation list sync has failed {{ $value }} times in the last 5 minutes on {{ $labels.instance }}. Check issuer health and network connectivity."
```

## Monitoring Setup

### Metrics Scrape Configuration

Ensure your Prometheus scrape config includes OathMesh metrics endpoint:

```yaml
scrape_configs:
  - job_name: 'oathmesh'
    static_configs:
      - targets: ['localhost:9090']  # Update with actual gateway address
    metrics_path: '/metrics'
```

### Alert Routing

Configure AlertManager to route this alert to your on-call team:

```yaml
route:
  receiver: 'oncall'
  routes:
    - match:
        severity: critical
        component: revocation
      receiver: 'oncall'
      repeat_interval: 5m

receivers:
  - name: 'oncall'
    pagerduty_configs:
      - service_key: 'YOUR_SERVICE_KEY'
        severity: 'critical'
```

## Troubleshooting

### Revocation Sync is Failing

1. **Check issuer connectivity:**
   ```bash
   curl -v http://issuer:8080/v1/revoked-subjects
   ```
   Should return HTTP 200 with JSON: `{"revocations": [...]}`

2. **Check gateway logs:**
   ```bash
   grep "revocation list sync failed" /var/log/oathmesh/gateway.log
   ```

3. **Check Redis (if applicable):**
   ```bash
   redis-cli ping  # Should return PONG
   ```

4. **Review metrics:**
   ```bash
   curl http://localhost:9090/metrics | grep oathmesh_revocation_sync_errors
   ```

### False Positives

The alert will fire if ANY sync attempt fails. Brief transient failures are normal, but if they persist:

- Check network latency to issuer (timeout is 5 seconds)
- Verify issuer is not under heavy load
- Consider increasing sync interval if too frequent

## Metrics Export

OathMesh exports all metrics on the `/metrics` HTTP endpoint. The `oathmesh_revocation_sync_errors` counter increments by 1 for each failed sync attempt (both initial and periodic).

```
# HELP oathmesh_revocation_sync_errors Total number of revocation list sync failures
# TYPE oathmesh_revocation_sync_errors counter
oathmesh_revocation_sync_errors 0
```
