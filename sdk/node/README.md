# @oathmesh/sdk

OathMesh token verification for **Express.js** and **Next.js** — TypeScript-first.

```bash
npm install @oathmesh/sdk
```

---

## Express.js

```typescript
import express from 'express';
import { verifyToken } from '@oathmesh/sdk';

const app = express();

app.use(verifyToken({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.dev'],
}));

app.get('/inventory', (req, res) => {
  const caller = req.oathmeshContext!;
  res.json({ subject: caller.principal.subject, action: caller.action });
});
```

---

## Next.js — App Router (Route Handlers)

```typescript
// app/api/inventory/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { withOathMesh } from '@oathmesh/sdk/next';

const oathmesh = withOathMesh({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.dev'],
});

export async function GET(request: NextRequest) {
  const { caller, error } = await oathmesh(request);
  if (error) return error;     // 401 with structured error body

  return NextResponse.json({
    subject: caller.principal.subject,
    action: caller.action,
  });
}
```

## Next.js — Pages Router (API Routes)

```typescript
// pages/api/inventory.ts
import { withOathMeshApi } from '@oathmesh/sdk/next';

export default withOathMeshApi(
  {
    audience: 'https://inventory.internal',
    trustedIssuers: ['https://issuer.oathmesh.dev'],
  },
  (req, res) => {
    const caller = (req as any).oathmeshContext;
    res.json({ subject: caller.principal.subject });
  }
);
```

## Next.js — Edge Middleware

```typescript
// middleware.ts (project root)
import { NextRequest, NextResponse } from 'next/server';
import { createEdgeVerifier } from '@oathmesh/sdk/next';

const verify = createEdgeVerifier({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.dev'],
});

export async function middleware(request: NextRequest) {
  const denied = await verify(request);
  if (denied) return denied;     // 401 — stops the request

  return NextResponse.next();    // Continue to the route handler
}

export const config = {
  matcher: '/api/:path*',
};
```

---

## Core Verifier (Framework-agnostic)

Use `verifyOathToken` directly in any runtime — Hono, Fastify, or raw Node:

```typescript
import { verifyOathToken, extractToken, OathMeshError } from '@oathmesh/sdk';

const token = extractToken(headers.authorization);

try {
  const caller = await verifyOathToken(token!, {
    audience: 'https://inventory.internal',
    trustedIssuers: ['https://issuer.oathmesh.dev'],
  });
  console.log(caller.principal.subject);
} catch (err) {
  if (err instanceof OathMeshError) {
    console.error(err.code, err.fix);
  }
}
```

---

## Lifecycle Hooks

Monitor verification events with `onVerified` and `onDenied`:

```typescript
verifyToken({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.dev'],
  onVerified: (caller) => {
    metrics.increment('oathmesh.allow', { sub: caller.principal.subject });
  },
  onDenied: (err) => {
    logger.warn('oathmesh denied', { code: err.code, message: err.message });
  },
});
```

---

## Error Responses

All verification failures return HTTP 401 with a stable JSON shape:

```json
{
  "error": "audience_mismatch",
  "message": "token audience does not match",
  "fix": "mint with --aud https://inventory.internal"
}
```

| Code | Trigger |
|---|---|
| `claim_missing:token` | Missing/invalid Authorization header |
| `algorithm_not_allowed` | `typ` ≠ `om+jwt` or `alg` = `none`/unsupported |
| `issuer_untrusted` | Issuer not in trusted list |
| `signature_invalid` | JWKS signature verification failed |
| `token_expired` | Past expiry + 10s clock skew |
| `audience_mismatch` | `aud` doesn't match configured audience |
| `claim_missing:{sub,act,jti}` | Missing required claim |
| `verification_failed` | Catch-all for malformed tokens |

---

## Development

```bash
npm install
npm test          # vitest — 12 tests
npm run build     # tsc → dist/
```
