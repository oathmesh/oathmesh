# Audit Events

## Overview

Every token verification attempt emits a structured NDJSON audit event—whether the outcome is `allow` or `deny`. This is **never conditional**. Audit events are the primary mechanism for operational visibility into OathMesh.

## Event Schema

```json
{
  "event": "oathmesh.verify",
  "outcome": "allow",
  "reason": "policy:deploy-write",
  "jti": "550e8400-e29b-41d4-a716-446655440000",
  "sub": "agent://repo/acme/deploy-bot",
  "aud": "https://inventory.internal",
  "act": "inventory.write",
  "iss": "https://issuer.oathmesh.dev",
  "env": "prod",
  "src": {
    "type": "github_actions",
    "repo": "acme/storefront",
    "workflow": "deploy.yml"
  },
  "timestamp": "2026-04-12T14:30:00.000Z",
  "request_id": "req-uuid-here"
}
```

## Fields

| Field | Type | Description |
|---|---|---|
| `event` | string | Always `"oathmesh.verify"` |
| `outcome` | string | `"allow"` or `"deny"` |
| `reason` | string | On allow: `"policy:<rule_name>"`. On deny: the error code (e.g., `"token_expired"`, `"policy_denied"`) |
| `jti` | string | Token's unique identifier |
| `sub` | string | Caller's subject URI |
| `aud` | string | Intended audience |
| `act` | string | Requested action |
| `iss` | string | Token issuer |
| `env` | string | Environment label (if present) |
| `src` | object | Source provenance (if present) |
| `timestamp` | string | ISO 8601 timestamp of the verification |
| `request_id` | string | Request identifier for correlation |

## Audit Sinks

| Sink | Use |
|---|---|
| `StdoutAuditSink` | Default. Writes NDJSON to stdout. Ideal for containers and cloud-native deployments where logs are collected by the platform. |
| `FileAuditSink` | Writes NDJSON to an append-only file. Path configured via `OATHMESH_AUDIT_FILE`. Protected by a mutex for concurrent writes. |
| `NoopAuditSink` | **Testing only.** Discards all events. Clearly marked in code. Never used in production. |

## Configuration

```bash
OATHMESH_AUDIT_SINK=stdout           # stdout | file | noop
OATHMESH_AUDIT_FILE=/var/log/oathmesh/audit.ndjson  # required when sink=file
```

## Security Rules

- **Never log the full Oath Token string.** Log `jti` + claim summary only.
- **Never log private key material** under any circumstances.
- Audit events are emitted on **every** verification call—allow and deny, no exceptions.
