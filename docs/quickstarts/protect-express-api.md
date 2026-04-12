# Quickstart: Protect an Express API

**Time:** ~5 minutes

This guide adds OathMesh token verification to an Express.js API using the `@oathmesh/sdk` package.

## Prerequisites

- Node.js 18+
- A running OathMesh issuer

## Step 1: Install the SDK

```bash
npm install @oathmesh/sdk
```

## Step 2: Mount the Middleware

```javascript
const express = require('express');
const { verifyToken } = require('@oathmesh/sdk');

const app = express();

// Mount the OathMesh verification middleware
app.use(verifyToken({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.dev'],
}));

app.get('/inventory', (req, res) => {
  // req.oathmeshContext is populated by the middleware
  const caller = req.oathmeshContext;
  res.json({
    caller: caller.principal.subject,
    action: caller.action,
  });
});

app.listen(3000, () => console.log('Listening on :3000'));
```

## Step 3: TypeScript Support

The SDK ships with TypeScript declarations. `req.oathmeshContext` is fully typed:

```typescript
import { verifyToken, VerifiedCallerContext } from '@oathmesh/sdk';

app.get('/inventory', (req, res) => {
  const caller: VerifiedCallerContext = req.oathmeshContext!;
  res.json({ subject: caller.principal.subject });
});
```

## Step 4: Test It

```bash
TOKEN=$(oathmesh mint \
  --sub "svc://frontend/web" \
  --aud "https://inventory.internal" \
  --act "inventory.read" \
  --quiet)

curl -H "Authorization: OathMesh $TOKEN" http://localhost:3000/inventory
```

## Error Responses

All errors follow the OathMesh taxonomy:

```json
{
  "code": "audience_mismatch",
  "message": "wrong audience"
}
```

## Next Steps

- [Protect a FastAPI service](protect-fastapi.md)
- [GitHub Actions to internal API](github-actions-to-internal-api.md)
