# nextjs-api Troubleshooting

## Pattern Decision Flow

```text
Need route-level auth? --> App Router (`withOathMesh`) or Pages (`withOathMeshApi`)
Need global /api/* auth? --> Edge middleware (`createEdgeVerifier`)
```

## Quick Checks

| Check | Command |
|---|---|
| Install deps | `npm install` |
| Run app | `npm run dev` |
| Test App Router | `curl -H "Authorization: OathMesh $TOKEN" http://localhost:3001/api/inventory` |
| Test Pages Router | `curl -H "Authorization: OathMesh $TOKEN" http://localhost:3001/api/legacy` |

## Common Issues

| Issue | Cause | Fix |
|---|---|---|
| Always `401` on API routes | Missing/invalid `Authorization` header | Send `Authorization: OathMesh <token>` |
| `issuer_untrusted` | `trustedIssuers` mismatch | Match issuer URL exactly |
| Edge runtime error | Node APIs used in edge middleware | Keep edge path to web-standard APIs only |
| Route not protected | Matcher or wiring issue | Verify `matcher: '/api/:path*'` and middleware export |
