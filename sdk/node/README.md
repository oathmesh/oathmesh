# @oathmesh/sdk

OathMesh token verification middleware for Express.js — **TypeScript-first**.

## Installation

```bash
npm install @oathmesh/sdk
```

## Quick Start

```typescript
import express from 'express';
import { verifyToken } from '@oathmesh/sdk';

const app = express();

// Protect all routes
app.use(verifyToken({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.dev'],
}));

app.get('/inventory', (req, res) => {
  // req.oathmeshContext is fully typed
  const caller = req.oathmeshContext!;
  res.json({
    subject: caller.principal.subject,
    action: caller.action,
  });
});

app.listen(3000);
```

## JavaScript Usage

Works identically in plain JavaScript:

```javascript
const { verifyToken } = require('@oathmesh/sdk');

app.use(verifyToken({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.dev'],
}));
```

## API

### `verifyToken(config: VerifierConfig)`

Returns an Express middleware. On success, populates `req.oathmeshContext`.

| Config Field | Type | Description |
|---|---|---|
| `audience` | `string` | Expected audience URL (exact match) |
| `trustedIssuers` | `string[]` | Trusted issuer URLs |

### `req.oathmeshContext: VerifiedCallerContext`

```typescript
interface VerifiedCallerContext {
  principal: { issuer: string; subject: string };
  action: string;
  tokenId: string;
  environment: string;
  scope?: string[];
  reason?: string;
  source?: { type: string; repo?: string; workflow?: string; runId?: string; sha?: string };
}
```

### `OathMeshError`

Thrown/returned on all verification failures:

```typescript
class OathMeshError extends Error {
  code: ErrorCode;   // e.g., 'audience_mismatch', 'token_expired'
  message: string;   // human-readable cause
  fix?: string;      // actionable fix instruction
}
```

## Error Responses

All errors return HTTP 401 with a JSON body:

```json
{
  "error": "audience_mismatch",
  "message": "wrong audience",
  "fix": "mint with aud set to https://inventory.internal"
}
```

| Code | Trigger |
|---|---|
| `claim_missing:token` | Missing or non-`OathMesh` Authorization header |
| `algorithm_not_allowed` | `typ` is not `om+jwt` or `alg` is `none` |
| `issuer_untrusted` | Issuer not in `trustedIssuers` |
| `signature_invalid` | JWKS signature verification failed |
| `token_expired` | Token past expiry + 10s clock skew |
| `audience_mismatch` | `aud` doesn't match configured audience |
| `claim_missing:sub` | Missing `sub` claim |
| `claim_missing:act` | Missing `act` claim |
| `claim_missing:jti` | Missing `jti` claim |

## Development

```bash
npm install
npm test          # vitest
npm run build     # tsc → dist/
```
