# Quickstart: Protect an Express API

**Time:** ~5 minutes

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
  trustedIssuers: ['https://issuer.oathmesh.dev'],
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

## Next Steps

- [Protect a Next.js API](protect-nextjs-api.md)
- [Protect a FastAPI service](protect-fastapi.md)
