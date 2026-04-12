# @oathmesh/sdk

OathMesh verification middleware for Express.js and Node.js.

## Installation

```bash
npm install @oathmesh/sdk
```

## Usage

```javascript
const express = require('express');
const { verifyToken } = require('@oathmesh/sdk');

const app = express();

app.use(verifyToken({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.dev'],
}));

app.get('/inventory', (req, res) => {
  const caller = req.oathmeshContext;
  res.json({ caller: caller.principal.subject });
});
```

## TypeScript

Full TypeScript declarations are included. `req.oathmeshContext` is typed as `VerifiedCallerContext`.

```typescript
import { verifyToken, VerifiedCallerContext } from '@oathmesh/sdk';
```

## API

### `verifyToken(config)`

Returns an Express middleware function.

**Config:**
- `audience` (string, required) — The audience URL this receiver expects.
- `trustedIssuers` (string[], required) — List of trusted issuer URLs.

### `req.oathmeshContext`

After successful verification, the request object is populated with:

```typescript
{
  principal: { subject: string; issuer: string };
  action: string;
  tokenId: string;
  environment: string;
}
```

## Error Responses

On verification failure, the middleware returns 401 with a JSON body:

```json
{ "code": "audience_mismatch", "message": "wrong audience" }
```

See the [Error Taxonomy](../../docs/protocol/error-taxonomy.md) for all error codes.
