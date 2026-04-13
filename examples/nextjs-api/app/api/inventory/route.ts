// examples/nextjs-api — Next.js App Router example with OathMesh verification.
//
// This shows all three Next.js integration patterns:
//   1. App Router Route Handler (this file)
//   2. Edge Middleware (middleware.ts)
//   3. Pages Router API Route (pages/api/legacy.ts)

import { NextRequest, NextResponse } from 'next/server';
import { withOathMesh } from '@oathmesh/sdk/next';

const oathmesh = withOathMesh({
  audience: process.env.OATHMESH_AUDIENCE || 'https://inventory.internal',
  trustedIssuers: (process.env.OATHMESH_TRUSTED_ISSUERS || 'http://localhost:4000').split(','),
  onDenied: (err) => {
    console.warn(`[oathmesh] denied: ${err.code} — ${err.message}`);
  },
  onVerified: (ctx) => {
    console.log(`[oathmesh] allowed: ${ctx.principal.subject} → ${ctx.action}`);
  },
});

// GET /api/inventory
export async function GET(request: NextRequest) {
  const { caller, error } = await oathmesh(request);
  if (error) return error;

  return NextResponse.json({
    status: 'success',
    data: ['widget-a', 'widget-b', 'widget-c'],
    caller: {
      subject: caller.principal.subject,
      action: caller.action,
      tokenId: caller.tokenId,
    },
  });
}

// POST /api/inventory
export async function POST(request: NextRequest) {
  const { caller, error } = await oathmesh(request);
  if (error) return error;

  const body = await request.json();
  return NextResponse.json({
    status: 'created',
    item: body,
    createdBy: caller.principal.subject,
  });
}
