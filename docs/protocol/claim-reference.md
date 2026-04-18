# Claim Reference

## Required Claims

All required claims **must** be present in every Oath Token. A missing required claim triggers immediate rejection with `claim_missing:{claim}`.

| Claim | Type | Description |
|---|---|---|
| `iss` | string | Canonical issuer URL. Must exactly match the receiver's trusted issuer configuration. |
| `sub` | string (URI) | Subject URI identifying the caller. Must use a standardized scheme: `svc://`, `agent://`, `job://`, `tool://`, or `user://`. |
| `aud` | string (URL) | Audience URL identifying the intended receiver. Must exactly match the receiver's configured audience. No globbing, no prefix matching. |
| `act` | string | Action the caller intends to perform (e.g., `"inventory.write"`, `"deploy"`, `"read"`). |
| `iat` | integer (Unix timestamp) | Issued-at time. Set server-side by the issuer. Callers cannot influence this value. |
| `exp` | integer (Unix timestamp) | Expiration time. Set server-side as `iat + clamped_ttl`. Maximum value: `iat + 300`. Callers cannot influence this value. |
| `jti` | string (UUID) | Unique token identifier. Generated via `uuid.New()` — cryptographically random, never sequential, never predictable. Used for replay detection. |

## Optional Claims

Optional claims enrich the audit trail and policy evaluation context. They are not required for verification to succeed.

| Claim | Type | Description |
|---|---|---|
| `scope` | string[] | List of scope values (e.g., `["inventory.read", "inventory.write"]`). Policy rules can require all listed scopes to be present. |
| `reason` | string | Human-readable reason for the call. Appears in audit events. |
| `src` | object | Source provenance — where the call originated. |
| `src.type` | string | Source system type (e.g., `"github_actions"`). |
| `src.repo` | string | Repository name (e.g., `"acme/storefront"`). |
| `src.workflow` | string | Workflow file (e.g., `"deploy.yml"`). |
| `src.run_id` | string | CI run identifier. |
| `src.sha` | string | Git commit SHA. |
| `delegated_by` | string (URI) | If the call is on behalf of a human, their `user://` URI. |
| `env` | string | Environment label (e.g., `"prod"`, `"staging"`, `"dev"`). |
| `tenant` | string | Tenant identifier explicitly isolating logical boundaries for multi-tenant evaluation natively. |
| `rqh` | string | Request hash binding. Format: `"sha256:<hex>"`. When present, the receiver verifies `sha256(canonical_request) == rqh`. See [Replay Defense](../security/replay-defense.md). |

## Subject URI Schemes

Subjects must use standardized URI schemes. Freeform strings are not permitted.

| Scheme | Use | Example |
|---|---|---|
| `svc://` | Services and microservices | `svc://payments/processor` |
| `agent://` | AI agents and bots | `agent://repo/acme/deploy-bot` |
| `job://` | CI/CD jobs | `job://github/acme/storefront/deploy.yml` |
| `tool://` | MCP-adjacent tool clients | `tool://cursor/terminal` |
| `user://` | Human delegation context only | `user://mustafa` |

## TTL Policy

| Parameter | Value |
|---|---|
| Default TTL | 120 seconds |
| Maximum TTL | 300 seconds (hard ceiling, enforced server-side) |
| Write/mutate recommended | 60 seconds |
| `ttl_hint` from caller | Accepted, clamped to `max(1, min(hint, 300))` |
| `exp` field | Always server-side — callers cannot influence it |
