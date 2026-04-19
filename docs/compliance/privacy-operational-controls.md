# Privacy Operational Controls for OathMesh

> **Scope:** Technical privacy controls for operators running OathMesh in production.  
> **Boundary:** This is implementation guidance, not legal advice.

## 1) Pseudonymization and Redaction

Apply these controls to reduce exposure while preserving security utility:

### Token and claim handling

- Never log raw Oath token strings.
- Keep `sub` machine-oriented where possible (`svc://`, `agent://`, `job://`).
- If human-linked values are unavoidable (`user://`, `delegated_by`), consider deterministic hashing in downstream analytics views while preserving raw data in restricted audit storage.
- Treat `reason` and `src.*` as potentially sensitive metadata; avoid storing excess context.

### Log pipeline controls

- Implement ingest-time processors that:
  - reject events containing token-like bearer strings,
  - truncate unexpectedly long free-text fields,
  - optionally hash selected identifiers in non-forensic datasets.
- Separate forensic audit storage (restricted) from observability analytics (pseudonymized).

### Metrics controls

- Do not place `sub`, `jti`, `request_id`, repo names, or other high-cardinality identifiers in metric labels.
- Keep metric dimensions operational (`service`, `route`, `outcome`, `environment`) to avoid indirect identification.

## 2) Access Control and Auditability Controls

Use layered controls around audit and telemetry stores:

| Control area | Recommended control |
|---|---|
| Access model | RBAC with least privilege; separate responder, auditor, and admin roles. |
| Authentication | SSO + MFA for log/SIEM and backup consoles. |
| Authorization | Time-bound break-glass roles for incident response with approval trail. |
| Query auditing | Enable immutable audit logs for who queried/exported what and when. |
| Data egress | Restrict bulk export; require ticket/reference ID for downloads. |
| Key management | Encrypt log/backup stores with managed keys and rotation policy. |

Operational checks:

- Review privileged access at least quarterly.
- Alert on unusual log export volume or off-hours access.
- Include privacy controls in incident drills and post-incident reviews.

## 3) Backup/Restore Retention Alignment

Backup policy must not silently bypass retention policy.

### Alignment rules

1. **Retention parity:** Backup expiration should be less than or equal to approved archive retention for each data class.
2. **Deletion propagation:** Ensure deletion/anonymization workflows cover snapshots/backups where feasible.
3. **Restore controls:** Restored datasets must inherit current access controls, masking, and retention timers.
4. **Test evidence:** Run periodic restore tests and record whether old data exceeded retention policy.

### Practical implementation pattern

- Tag backup sets by data class (`audit`, `ops-logs`, `metrics`) and retention class.
- Automate expiration in backup tooling rather than manual cleanup.
- Maintain a restore runbook requiring privacy review before non-emergency restores of historical logs.

References: [Logging Guidance](../security/logging-guidance.md), [Production Best Practices](../operations/best-practices.md), [Incident Response Playbook](../operations/incident-response.md)
