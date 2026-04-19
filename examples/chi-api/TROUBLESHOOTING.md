# chi-api Troubleshooting

## Request Path (Go chi)

```text
Caller --> Authorization: OathMesh <token> --> chi middleware --> handler
                                             | pass     | fail
                                             v          v
                                         caller ctx   401 JSON error
```

## Quick Checks

| Check | Command |
|---|---|
| Start service | `go run main.go` |
| Mint token | `oathmesh mint --sub agent://repo/acme/bot --aud https://inventory.internal --act deploy --quiet` |
| Call endpoint | `curl -H "Authorization: OathMesh $TOKEN" http://localhost:8080/inventory` |

## Common Issues

| Issue | Cause | Fix |
|---|---|---|
| `caller context missing` | Route not protected with middleware | Ensure route is inside `r.Group(... r.Use(OathMeshMiddleware(cfg)))` |
| `issuer_untrusted` | `TrustedIssuers` mismatch | Add exact issuer URL in config |
| `replay_detected` on second request | Same token reused | Mint a fresh token per request |
| `connection refused :8080` | App not running or wrong port | Start app and verify port |
