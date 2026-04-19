# OathMesh Troubleshooting

Use this guide for fast triage across issuer, SDK, and deployment issues.

## Fast Triage Flow

```text
START
  |
  +--> Is issuer reachable? (GET /.well-known/jwks.json)
  |      | yes
  |      v
  |   Is token valid? (oathmesh verify)
  |      | yes
  |      v
  |   Does receiver config match token (iss/aud)?
  |      | yes
  |      v
  |   Is request denied by policy/replay/revocation?
  |      | yes
  |      v
  |   Inspect audit logs + error taxonomy step
  |
  +--> no at any step:
         fix that layer first, then retry end-to-end
```

## First 5 Checks

| Check | Command | Expected |
|---|---|---|
| Issuer health | `curl -fsS $OATHMESH_ISSUER/.well-known/jwks.json` | JSON with `keys` |
| Mint token | `oathmesh mint --sub ... --aud ... --act ... --quiet` | token string |
| Verify token offline | `oathmesh verify --token "$TOKEN" --aud ... --iss ...` | `valid: true` |
| Receiver audience | inspect app config/env | exact match to token `aud` |
| Receiver issuers | inspect app config/env | exact match to token `iss` |

## Common Errors

| Symptom | Likely Cause | Fix |
|---|---|---|
| `issuer_untrusted` | Receiver trusted issuer list missing token `iss` | Add exact issuer URL; no wildcard |
| `audience_mismatch` | Token `aud` and receiver `audience` differ | Mint with correct `--aud` or update receiver config |
| `token_expired` | Token TTL elapsed or clock skew too large | Mint fresh token; sync clocks (NTP) |
| `replay_detected` | Same token reused | Mint one token per request or per operation |
| `signature_invalid` | Wrong key/JWKS stale/issuer mismatch | Validate issuer JWKS and key rotation process |
| `claim_missing:token` | Missing `Authorization` header | Send `Authorization: OathMesh <token>` |
| `policy_denied` | Policy default deny or no matching allow rule | Add explicit allow rule and redeploy policy |

## Deployment-Level Triage

```text
Request denied in production
  |
  +--> App logs: error code + step
  +--> Audit log: allow/deny event + reason
  +--> Issuer logs: mint traffic and failures
  +--> Replay store health (Redis/memory)
  +--> Policy file and hot-reload status
```

## Example-Specific Guides

- [Go chi example troubleshooting](../examples/chi-api/TROUBLESHOOTING.md)
- [Express example troubleshooting](../examples/express-api/TROUBLESHOOTING.md)
- [FastAPI example troubleshooting](../examples/fastapi-api/TROUBLESHOOTING.md)
- [Next.js example troubleshooting](../examples/nextjs-api/TROUBLESHOOTING.md)
- [gRPC server troubleshooting](../examples/grpc-server/TROUBLESHOOTING.md)

## Still Blocked?

1. Capture exact error code and step from response body.
2. Capture one failing request path + headers (without secrets/token body).
3. Include issuer URL, expected audience, and deployment mode.
4. Open an issue or use support channels in `SUPPORT.md`.
