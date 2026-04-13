# Quickstart: Protect a Next.js API

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  <b>Add OathMesh token verification to Next.js APIs (App Router, Pages Router, or Edge).</b>
</p>

<p align="center">
  <b>⏱️ Time:</b> ~5 minutes
</p>

---

> 🆕 **New here?** Start with the [Quick Start](../README.md#-quick-start) in the main README.

## Prerequisites

- Node.js 18+, Next.js 13+
- A running OathMesh issuer

## App Router (Route Handlers)

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

export async function POST(request: NextRequest) {
  const { caller, error } = await oathmesh(request);
  if (error) return error;

  const body = await request.json();
  // caller.principal.subject tells you WHO is writing
  return NextResponse.json({ status: 'created', by: caller.principal.subject });
}
```

## Pages Router (API Routes)

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

## Edge Middleware (Protect All API Routes)

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
  if (denied) return denied;     // returns 401 — stops the request
  return NextResponse.next();
}

export const config = {
  matcher: '/api/:path*',        // only protect /api/* routes
};
```

## Which Pattern to Use?

| Pattern | When to Use |
|---|---|
| **App Router (`withOathMesh`)** | New Next.js 13+ projects using `app/` directory |
| **Pages Router (`withOathMeshApi`)** | Existing projects with `pages/api/` routes |
| **Edge Middleware (`createEdgeVerifier`)** | Protect all routes at the edge without per-route wiring |

## Next Steps

- [Protect a FastAPI service](protect-fastapi.md)
- [Protect an Express API](protect-express-api.md)
- [Protect a Go chi API](protect-chi-api.md)
- [GitHub Actions to internal API](github-actions-to-internal-api.md)

---

## Related Documentation

| Document | Description |
|----------|-------------|
| [Node SDK](../sdk/node/README.md) | Full SDK reference for Express/Next.js |
| [Verification Rules](../docs/protocol/verification-rules.md) | 14-step pipeline details |
| [Error Taxonomy](../docs/protocol/error-taxonomy.md) | All error codes |
| [Threat Model](../docs/security/threat-model.md) | Security model |
