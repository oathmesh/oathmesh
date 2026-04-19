← [Back to Index](../INDEX.md)

# OathMesh Production Go-Live Checklist

All items below are **must-pass** before production launch.

## Issuer Security
- [ ] `OATHMESH_ISSUER` is `https://...` in production.
- [ ] Signing key and mint secret are sourced from a secret manager/runtime secret, not committed files.
- [ ] Key rotation process is tested using [Key Rotation Guide](../deployment/key-rotation.md).
- [ ] Issuer runs as non-root with restricted runtime permissions.
- [ ] `OATHMESH_ENV=production` is set.

## Receiver Security
- [ ] Trusted issuers are explicitly configured (`OATHMESH_TRUSTED_ISSUERS` or gateway issuers) and match issuer URL.
- [ ] Audience enforcement is explicitly configured (`OATHMESH_AUDIENCE` or gateway audience) per receiver service.
- [ ] Replay cache is enabled and backed by Redis for distributed deployments.
- [ ] Request binding (`rqh`) is required for write/mutate endpoints.

## Revocation Reliability
- [ ] Redis is highly available for your environment and has persistence/backups configured.
- [ ] Alert on `oathmesh_revocation_sync_errors` is active and routed to on-call.
- [ ] Revocation outage drill completed (simulate Redis failure, validate recovery and sync health).

## Observability and Response
- [ ] Metrics endpoint is scraped and dashboards include revocation sync, verification errors, latency, uptime, and Redis health.
- [ ] Log policy prohibits full token/private key logging (only `jti` + claim summary).
- [ ] On-call has tested procedures from [On-Call Runbook](./on-call-runbook.md).
- [ ] Incident response procedures are current in [Incident Response Playbook](./incident-response.md).

## Deployment and TLS
- [ ] Deployment baseline follows one of:
  - [Kubernetes Deployment Guide](../deployment/kubernetes.md)
  - [Docker Compose Deployment](../deployment/docker-compose.md)
- [ ] TLS baseline validated per [TLS Configuration Guide](../deployment/tls.md).

## Final Sign-Off
- [ ] Security owner sign-off
- [ ] SRE/on-call owner sign-off
- [ ] Service owner sign-off

For detailed rationale and control guidance, see [Production Best Practices](./best-practices.md).
