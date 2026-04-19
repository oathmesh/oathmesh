# Enterprise Audit Logging

Audit logging in OathMesh should produce evidence that is useful for engineers, security teams, and auditors.

## Goals

- Reconstruct who did what, when, where, and why
- Detect suspicious or unauthorized behavior quickly
- Provide durable, tamper-evident records for investigations and audits

## Event model baseline

Start with OathMesh protocol guidance:

- [Audit Events Reference](../protocol/audit-events.md)
- [Verification Rules](../protocol/verification-rules.md)
- [Logging Guidance](../security/logging-guidance.md)

At minimum, log these fields for every decision:

- `timestamp` (UTC, ISO-8601)
- `event_type` (verification allow/deny, issuer mint/reject, policy decision)
- `request_id` / trace correlation ID
- `iss`, `sub`, `aud`, `act`, `jti`
- decision result and structured reason code
- environment metadata (service name, region, cluster, version)

## Logging architecture

Recommended flow:

1. Emit structured NDJSON from issuer and receiver components
2. Forward to centralized collection (SIEM/data lake)
3. Enforce immutable retention for compliance period
4. Restrict read access by role
5. Alert on critical deny/replay/tamper signals

See operational counterparts:

- [Alerting Rules](../operations/alerting-rules.md)
- [Grafana Dashboards](../operations/grafana-dashboards.md)
- [Incident Response](../operations/incident-response.md)

## Data minimization and privacy

Do not log secrets, full tokens, or unnecessary personal data.

- Prefer token metadata and hashed identifiers where possible
- Mask high-risk fields before storage
- Use retention tiers aligned to policy and regulation

For retention design:

- [GDPR Data Retention](../compliance/gdpr-data-retention.md)
- [Privacy Operational Controls](../compliance/privacy-operational-controls.md)

## Integrity and tamper resistance

- Use append-only pipelines where possible
- Enable storage immutability/WORM for required windows
- Sign or checksum log batches for chain-of-custody
- Separate write and read permissions
- Audit all access to audit data itself

## Review and evidence cadence

Run a consistent review process:

- Daily: failed verification spikes, replay anomalies, unknown audiences
- Weekly: policy-deny trend review by service owner
- Monthly: control attestation sampling and evidence export
- Quarterly: tabletop incident simulation and log forensics drill

## Operational checklist

- [ ] Structured logs enabled on all issuer/receiver paths
- [ ] Correlation IDs propagated across services
- [ ] Replay-deny events alerting configured
- [ ] Retention and immutability policies documented
- [ ] Access controls and approvals documented
- [ ] Review cadence assigned to named owners

Next: map this implementation to control frameworks in [Compliance Guide](compliance.md).
