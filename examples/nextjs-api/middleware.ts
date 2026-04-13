// examples/nextjs-api — Edge Middleware that protects all /api/* routes.
//
// This runs at the edge BEFORE route handlers. If a request fails
// verification, the route handler never executes.

import { NextRequest, NextResponse } from 'next/server';
import { createEdgeVerifier } from '@oathmesh/sdk/next';

const verify = createEdgeVerifier({
  audience: process.env.OATHMESH_AUDIENCE || 'https://inventory.internal',
  trustedIssuers: (process.env.OATHMESH_TRUSTED_ISSUERS || 'http://localhost:4000').split(','),
});

export async function middleware(request: NextRequest) {
  // Skip health check
  if (request.nextUrl.pathname === '/api/healthz') {
    return NextResponse.next();
  }

  const denied = await verify(request);
  if (denied) return denied;

  return NextResponse.next();
}

export const config = {
  matcher: '/api/:path*',
};
