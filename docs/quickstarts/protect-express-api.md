← [Back to Index](../INDEX.md)

# Tutorial: Protect an Express API

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

⏱️ **Time**: ~5 minutes | 📋 **Prerequisites**: Node.js 18+, running OathMesh issuer | 🎯 **Outcome**: Express.js API with OathMesh token verification middleware

---

> 🆕 **New here?** Start with [Getting Started](../GETTING_STARTED.md) for a guided introduction.

## Prerequisites

- Node.js 18+
- A running OathMesh issuer

## Step 1: Install

```bash
npm install @oathmesh/sdk
```

## Step 2: Mount the Middleware

```typescript
import express from 'express';
import { verifyToken } from '@oathmesh/sdk';

const app = express();

app.use(verifyToken({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.tech'],
  onDenied: (err) => console.warn('denied:', err.code),
}));

app.get('/inventory', (req, res) => {
  const caller = req.oathmeshContext!;
  res.json({
    subject: caller.principal.subject,
    action: caller.action,
  });
});

app.listen(3000, () => console.log('Listening on :3000'));
```

## Step 3: Test It

```bash
TOKEN=$(oathmesh mint \
  --sub "svc://frontend/web" \
  --aud "https://inventory.internal" \
  --act "inventory.read" \
  --quiet)

curl -H "Authorization: OathMesh $TOKEN" http://localhost:3000/inventory
```

## Troubleshooting

| Issue | Likely Cause | Fix |
|---|---|---|
| `401 issuer_untrusted` | `trustedIssuers` mismatch | Use the exact issuer URL used to mint tokens |
| `401 audience_mismatch` | API audience differs from token | Align middleware `audience` with `--aud` used at mint |
| TypeScript cannot find `req.oathmeshContext` | Missing SDK typing import/setup | Follow the Node SDK README type augmentation guidance |

## Next Steps

- [Protect a Next.js API](protect-nextjs-api.md)
- [Protect a FastAPI service](protect-fastapi.md)
- [Protect a Go chi API](protect-chi-api.md)
- [GitHub Actions to internal API](github-actions-to-internal-api.md)

---

## Related Documentation

| Document | Description |
|----------|-------------|
| [Node SDK](../../sdk/node/README.md) | Full SDK reference |
| [Verification Rules](../protocol/verification-rules.md) | 14-step pipeline details |
| [Error Taxonomy](../protocol/error-taxonomy.md) | All error codes |
| [Threat Model](../security/threat-model.md) | Security model |
