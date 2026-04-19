---
title: GDPR Data Retention Operations Guide
description: Operator guidance for data minimization, retention windows, deletion workflows, and DSR handling in OathMesh.
tags: [compliance, gdpr, data-retention, privacy, operations]
---

← [Back to Index](../INDEX.md)

# GDPR Data Retention Operations Guide

> **Scope:** Operator-facing technical guidance for OathMesh deployments.  
> **Boundary:** This is not legal advice; use it with your counsel and internal privacy program.

## 1) Data Categories in OathMesh

OathMesh is designed for machine identity, but some fields can still be personal data depending on your environment (for example, a `user://` subject, repo names, or delegated actor IDs).

| Category | Where it appears | Typical fields | Operational note |
|---|---|---|---|
| Token claims (processed at verify time) | In-memory during issuance/verification, claim summary in logs | `iss`, `sub`, `aud`, `act`, `iat`, `exp`, `jti`, optional `scope`, `reason`, `src.*`, `delegated_by`, `env`, `tenant` | Do not persist raw token strings; persist only minimum claim summary needed for security/audit operations. |
| Audit logs | NDJSON audit stream (`stdout` or file sink) | `event`, `outcome`, `reason`, `jti`, `sub`, `aud`, `act`, `iss`, `env`, `src.*`, `timestamp`, `request_id` | Audit events are emitted on allow + deny; retention is operator-managed in log platform. |
| Telemetry/metrics | `/metrics`, dashboards, alerting | verification/error counters, latency histograms, revocation sync errors | Keep labels low-cardinality and avoid adding direct identifiers to metric labels. |

References: [Claim Reference](../protocol/claim-reference.md), [Audit Events](../protocol/audit-events.md), [Logging Guidance](../security/logging-guidance.md), [Alerting Rules](../operations/alerting-rules.md)

## 2) Lawful-Basis and Data Minimization Guidance (Operator Framing)

Treat OathMesh data as **security and service-operation data**. Configure processing so only fields required for:

1. token verification and replay defense,
2. incident detection/response,
3. controlled forensic review,
4. compliance evidence generation.

Operator best practices:

- Use stable machine identifiers in `sub` where possible (`svc://`, `agent://`, `job://`) instead of person-identifying values.
- Use `reason` as a short operational reason; avoid free-form text containing personal data.
- Only include optional claims (`src`, `tenant`, `delegated_by`, `scope`) when needed for authorization or auditability.
- Follow the logging rule: log `jti` + claim summary; never log raw tokens or private keys.
- Apply role-based access to audit pipelines so only responders and auditors can query raw event detail.

## 3) Retention Schedule and Deletion Strategy

OathMesh does not enforce storage retention; implement these controls in your log/backup systems.

### Recommended Retention Schedule

| Data set | Hot retention (searchable) | Cold/archive retention | Deletion trigger |
|---|---|---|---|
| Audit NDJSON events | 30-90 days | Up to 12 months (or per policy) | Age-based TTL + legal hold exceptions |
| Operational app logs | 14-30 days | 90-180 days | Age-based TTL |
| Metrics time series | 14-30 days | Optional downsampled aggregate only | TSDB retention window |
| Incident evidence exports | Case duration only | Case record policy | Incident closure + approved retention period |

### Deletion Strategy

1. **Primary logs:** Use index/table partition TTL in SIEM/log storage for automatic aging.
2. **Archives/backups:** Set backup expiration not to exceed approved archive period.
3. **Targeted erasure workflow:** For subject-linked records, run scoped delete/anonymization in downstream stores (subject to security/legal hold constraints).
4. **Verification:** Maintain deletion job logs showing dataset, range, actor, and completion status.

## 4) Data Subject Request (DSR) Handling Playbook

Use this operational flow when handling access/erasure requests involving OathMesh data:

1. **Intake and identity verification**  
   Validate requester identity via your existing privacy process (outside OathMesh).
2. **Scope discovery**  
   Identify relevant keys for search: `sub`, `delegated_by`, `src.repo`, `request_id`, time range.
3. **Data retrieval**  
   Query audit stores for matching events; export only required fields.
4. **Review and redaction**  
   Remove third-party identifiers not required for disclosure.
5. **Action**  
   - Access request: provide filtered export.  
   - Erasure/restriction request: apply retention exception workflow and targeted anonymization/deletion where permissible.
6. **Recordkeeping**  
   Capture ticket ID, datasets searched, response date, approver, and any legal/security hold rationale.

## 5) Breach Notification Workflow References

Use existing incident runbooks as the execution path:

- [Incident Response Playbook](../operations/incident-response.md)
- [On-Call Runbook](../operations/on-call-runbook.md)
- [Production Best Practices](../operations/best-practices.md)

For privacy incidents, add these required operator steps:

1. Flag incident as potential personal-data exposure.
2. Preserve relevant audit evidence and access logs.
3. Coordinate Security + Privacy/Legal response owners.
4. Execute jurisdiction-specific notification obligations through your legal process.
5. Document timeline, impacted data categories, and remediation actions.
