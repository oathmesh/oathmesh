# Logging Guidance

## What to Log

### Always Log (Audit Events)

Every verification attempt emits an audit event containing:

- `jti` — token unique identifier
- `sub` — caller subject URI
- `aud` — intended audience
- `act` — requested action
- `iss` — token issuer
- `outcome` — `allow` or `deny`
- `reason` — policy rule name or error code
- `timestamp` — ISO 8601
- `request_id` — for correlation

### Application Logs

- Issuer startup: port, configured issuer URL, key `kid`, audit sink type
- JWKS fetch events: URL, success/failure, cache hit/miss
- Policy load events: file path, rule count, hot-reload events
- Rate limiting events: source IP, requests blocked
- Error details (internal only): full error chain for debugging

## What NEVER to Log

### Private Key Material

**Never log the private key.** Not the full PEM, not a hash of it, not a truncated version. This is non-negotiable.

### Full Oath Token Strings

**Never log the full token.** Tokens contain a valid signature — logging them creates a replay risk if logs are compromised. Log `jti` + claim summary only.

### Correct ✅

```json
{"event":"oathmesh.verify","jti":"550e8400-...","sub":"agent://repo/acme/deploy-bot","outcome":"allow"}
```

### Incorrect ❌

```json
{"event":"oathmesh.verify","token":"eyJ0eXAiOiJvbStqd3QiLCJhbGciOi..."}
```

## Structured Logging

OathMesh uses Go's `slog` package for structured logging. All log output is JSON-formatted for machine parsing.

### Log Levels

| Level | Use |
|---|---|
| `debug` | JWKS cache details, policy evaluation trace, request parsing |
| `info` | Startup, shutdown, key rotation, policy reload (default) |
| `warn` | Rate limiting triggered, JWKS fetch retry, cache unavailable |
| `error` | Unrecoverable errors, panic recovery, Redis connection failure |

### Configuration

```bash
OATHMESH_LOG_LEVEL=info  # debug | info | warn | error
```

## Log Retention

Audit logs (`NDJSON` events) should be retained according to your organization's compliance requirements. OathMesh does not enforce retention — it only emits events. Use your log aggregation platform (Datadog, Elasticsearch, CloudWatch, etc.) for retention and alerting.
