← [Back to Index](../INDEX.md)

# OathMesh Production Best Practices

This guide consolidates secure, reliable operating practices for production OathMesh deployments.

## 1) Issuer Hardening

### Secrets and key material
- Keep `OATHMESH_PRIVATE_KEY` and `OATHMESH_MINT_SECRET` in a secret manager or runtime secret store (never in Git or image layers).
- Use `OATHMESH_PRIVATE_KEY` or `OATHMESH_PRIVATE_KEY_B64` in production; avoid file-based key loading except controlled bootstrap flows.
- Restrict key read access to the issuer runtime identity only (least privilege IAM/RBAC).
- Rotate signing keys on a schedule and keep overlap windows controlled. Follow [Key Rotation Runbook](./key-rotation.md).

### Transport and endpoint security
- Set `OATHMESH_ISSUER` to an `https://` URL in production.
- Enforce TLS 1.2+ at ingress/LB and verify certificate rotation is active. See [TLS Configuration Guide](../deployment/tls.md).
- Restrict issuer ingress to trusted callers (private ingress, allowlists, WAF, or mTLS at edge).

### Runtime hardening
- Run issuer containers as non-root with restricted security context (drop unnecessary capabilities, read-only FS when possible).
- Apply CPU/memory requests and limits; set liveness/readiness probes.
- Pin image digests for production rollouts and use explicit change control.
- Keep `OATHMESH_ENV=production`.

References: [Kubernetes Deployment Guide](../deployment/kubernetes.md), [Docker Compose Deployment](../deployment/docker-compose.md), [Issuer Configuration Reference](../config/issuer-config.md)

## 2) Receiver Hardening

### Trust and claim validation
- Configure explicit trusted issuers (`OATHMESH_TRUSTED_ISSUERS` or `OATHMESH_GATEWAY_ISSUERS`) and keep them aligned with issuer URL.
- Enforce exact audience validation (`OATHMESH_AUDIENCE` or `OATHMESH_GATEWAY_AUDIENCE`) per service.
- Reject tokens that fail required claim checks (`iss`, `sub`, `aud`, `act`, `iat`, `exp`, `jti`).

### Replay and request binding
- Back replay detection with Redis for multi-instance receivers; avoid memory-only replay cache for distributed production paths.
- Monitor replay rejections and verification failures for anomaly spikes.
- Require request binding (`rqh`) for write/mutate endpoints:
  - Node SDK: `requireRequestBinding: true`
  - Python SDK: `require_request_binding=True`
  - Go/policy paths: enforce request-hash checks in verification policy

References: [Replay Defense](../security/replay-defense.md), [Verification Rules](../protocol/verification-rules.md), [Claim Reference](../protocol/claim-reference.md)

## 3) Revocation Reliability

- Treat Redis availability as a security dependency, not just a performance dependency.
- Configure Redis persistence/backups and set restart/eviction policies deliberately.
- Alert on any revocation sync failure (`increase(oathmesh_revocation_sync_errors[5m]) > 0`) and page on-call.
- Track sustained sync error rates, not only single spikes, to catch partial outages.
- Add a recovery check to ensure sync returns to healthy after Redis/issuer remediation.

References: [Alerting Rules](./alerting-rules.md), [On-Call Runbook](./on-call-runbook.md), [Incident Response Playbook](./incident-response.md)

## 4) Observability Checklist

Must-have telemetry and operational checks:

- Metrics
  - `oathmesh_revocation_sync_errors` (critical)
  - `oathmesh_verification_errors`
  - `oathmesh_http_requests_duration_seconds` (latency SLO)
  - `oathmesh_issuer_uptime_seconds`
  - Redis health/memory/client metrics
- Logging
  - Structured logs for verify/mint outcomes and revocation sync events
  - Never log full tokens or private key material; log `jti` + claim summary only
- Alerting
  - Revocation sync failure routed to on-call
  - Verification error surge alerts
  - Issuer/Redis health and resource pressure alerts
- Dashboards
  - One at-a-glance panel for issuer health, revocation sync, verification errors, and p95/p99 latency

References: [Alerting Rules](./alerting-rules.md), [On-Call Runbook](./on-call-runbook.md), [Logging Guidance](../security/logging-guidance.md)

## 5) Incident Readiness Checklist

- Keep runbooks current and linked in alerts:
  - [Incident Response Playbook](./incident-response.md)
  - [On-Call Runbook](./on-call-runbook.md)
  - [Alerting Rules](./alerting-rules.md)
- Pre-stage emergency key rotation access and verify responders can execute it end-to-end.
- Verify responders can restart issuer/Redis and confirm revocation sync recovery.
- Run periodic incident drills for:
  - private key compromise
  - issuer outage
  - Redis revocation outage
- Require post-incident updates to runbooks/checklists after every high/critical event.

## 6) Privacy and Retention Operations

- Define explicit retention windows for audit logs, operational logs, and metrics in your log/backup platforms.
- Ensure deletion workflows and backup expiration are aligned with the same retention classes.
- Restrict access to audit stores with least privilege, MFA, and query/export auditing.
- Keep privacy runbooks linked for responder use:
  - [GDPR Data Retention Operations Guide](../compliance/gdpr-data-retention.md)
  - [Privacy Operational Controls](../compliance/privacy-operational-controls.md)

For release gating, use [Production Go-Live Checklist](./production-checklist.md).
