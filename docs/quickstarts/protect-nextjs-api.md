← [Back to Index](../INDEX.md)

# Tutorial: Protect a Next.js API

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

⏱️ **Time**: ~5 minutes | 📋 **Prerequisites**: Node.js 18+, Next.js 13+, running OathMesh issuer | 🎯 **Outcome**: Next.js API routes (App Router, Pages Router, or Edge) with OathMesh protection

---

> 🆕 **New here?** Start with [Getting Started](../GETTING_STARTED.md) for a guided introduction.

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
  trustedIssuers: ['https://issuer.oathmesh.tech'],
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
    trustedIssuers: ['https://issuer.oathmesh.tech'],
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
  trustedIssuers: ['https://issuer.oathmesh.tech'],
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

## Troubleshooting

| Issue | Likely Cause | Fix |
|---|---|---|
| Route always returns `401` | Middleware protects route but token missing | Send `Authorization: OathMesh <token>` from client/caller |
| `issuer_untrusted` in API routes | `trustedIssuers` mismatch | Match configured issuer to token `iss` exactly |
| Edge runtime errors | Using unsupported Node APIs in edge path | Keep verification logic in supported edge-safe APIs only |

## Next Steps

- [Protect a FastAPI service](protect-fastapi.md)
- [Protect an Express API](protect-express-api.md)
- [Protect a Go chi API](protect-chi-api.md)
- [GitHub Actions to internal API](github-actions-to-internal-api.md)

---

## Related Documentation

| Document | Description |
|----------|-------------|
| [Node SDK](../../sdk/node/README.md) | Full SDK reference for Express/Next.js |
| [Verification Rules](../protocol/verification-rules.md) | 14-step pipeline details |
| [Error Taxonomy](../protocol/error-taxonomy.md) | All error codes |
| [Threat Model](../security/threat-model.md) | Security model |
