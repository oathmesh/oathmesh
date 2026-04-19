# Enterprise Compliance Guide

This guide helps teams map OathMesh controls to common enterprise requirements and maintain auditable evidence.

## Scope and assumptions

OathMesh provides application-layer call identity and verification controls. It should be combined with your cloud IAM, network controls, and endpoint hardening.

Start here:

- [SOC2 Compliance Matrix](../security/soc2-compliance.md)
- [Threat Model](../security/threat-model.md)
- [Key Management](../security/key-management.md)
- [Privacy Operational Controls](../compliance/privacy-operational-controls.md)

## Control domains

### 1) Identity and access control

Relevant OathMesh controls:

- Short-lived signed tokens (max TTL)
- Audience and action scoping
- Replay defense using `jti`
- Default-deny policy behavior

Evidence examples:

- Policy files and approval records
- Verification-deny logs with reason codes
- Key rotation run records

References:

- [Token Format](../protocol/token-format.md)
- [Claim Reference](../protocol/claim-reference.md)
- [Replay Defense](../security/replay-defense.md)
- [Policy Overview](../policies/overview.md)

### 2) Change management and secure operations

Relevant practices:

- Controlled policy deployment process
- Incident response runbooks
- Monitoring and alerting for token verification failures

Evidence examples:

- PR approvals and release logs
- On-call incident timelines
- Alert definitions and dashboard snapshots

References:

- [On-Call Runbook](../operations/on-call-runbook.md)
- [Production Checklist](../operations/production-checklist.md)
- [Best Practices](../operations/best-practices.md)

### 3) Cryptography and key lifecycle

Relevant controls:

- Asymmetric signing (Ed25519)
- Key separation by environment/tenant
- Regular rotation and retirement workflows

Evidence examples:

- Key inventory and ownership
- Rotation schedules and execution logs
- Emergency rotation playbooks

References:

- [Key Management](../security/key-management.md)
- [Deployment Key Rotation](../deployment/key-rotation.md)
- [Operations Key Rotation](../operations/key-rotation.md)

### 4) Logging, monitoring, and accountability

Relevant controls:

- Structured audit events for allow/deny outcomes
- Alerting on suspicious patterns
- Controlled access to logs and retention boundaries

Evidence examples:

- SIEM queries and alert runbooks
- Log retention policies
- Quarterly review sign-offs

References:

- [Enterprise Audit Logging](audit-logging.md)
- [Audit Events Reference](../protocol/audit-events.md)
- [Logging Guidance](../security/logging-guidance.md)

### 5) Privacy and data governance

Relevant controls:

- Data minimization in tokens/logs
- Retention limits and deletion workflows
- Documented lawful basis and data handling boundaries

Evidence examples:

- Retention policy statements
- Data flow diagrams
- Data subject request handling procedures

References:

- [GDPR Data Retention](../compliance/gdpr-data-retention.md)
- [Privacy Operational Controls](../compliance/privacy-operational-controls.md)

## Practical audit packet checklist

- [ ] Architecture overview and trust boundaries
- [ ] Current policy set and approval workflow
- [ ] Key management and rotation evidence
- [ ] Representative audit logs (allow + deny)
- [ ] Monitoring/alerting configuration snapshots
- [ ] Incident response exercise records
- [ ] Privacy and retention policy artifacts

## Shared-responsibility reminder

OathMesh strengthens machine-call identity and verification. Final compliance posture depends on how your organization configures, operates, and monitors the full stack.
