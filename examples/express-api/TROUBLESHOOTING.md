# express-api Troubleshooting

## Request Path (Express)

```text
Caller --> Authorization header --> verifyToken middleware --> route handler
                                 | pass                    | fail
                                 v                         v
                           req.oathmeshContext         401 JSON error
```

## Quick Checks

| Check | Command |
|---|---|
| Install deps | `npm install` |
| Run app | `npx ts-node index.ts` |
| Test call | `curl -H "Authorization: OathMesh $TOKEN" http://localhost:3000/inventory` |

## Common Issues

| Issue | Cause | Fix |
|---|---|---|
| `401 issuer_untrusted` | Wrong `trustedIssuers` value | Match issuer URL exactly |
| `401 audience_mismatch` | `audience` does not match token | Align app config with mint `--aud` |
| `req.oathmeshContext` is undefined | Middleware not mounted before routes | Move `app.use(verifyToken(...))` above protected routes |
| TypeScript type errors on request context | Missing module augmentation/types setup | Follow `sdk/node/README.md` typing guidance |
