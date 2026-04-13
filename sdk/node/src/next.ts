/**
 * OathMesh Next.js adapters.
 *
 * Provides verification for:
 *   - App Router Route Handlers (GET, POST, etc.)
 *   - Pages Router API Routes
 *   - Next.js Edge Middleware
 *
 * All adapters use the same core verifier — no framework-specific crypto.
 */

import { OathMeshError, type VerifierConfig, type VerifiedCallerContext, type OathMeshErrorBody } from './types';
import { extractToken, verifyOathToken } from './verify';

// ─── App Router (Route Handlers) ────────────────────────────────────────────

/**
 * Verify an OathMesh token inside a Next.js App Router Route Handler.
 *
 * Returns the verified caller context or throws an OathMeshError.
 * Use with `NextRequest` in route handlers.
 *
 * @example
 * ```typescript
 * // app/api/inventory/route.ts
 * import { NextRequest, NextResponse } from 'next/server';
 * import { withOathMesh } from '@oathmesh/sdk/next';
 *
 * const oathmesh = withOathMesh({
 *   audience: 'https://inventory.internal',
 *   trustedIssuers: ['https://issuer.oathmesh.dev'],
 * });
 *
 * export async function GET(request: NextRequest) {
 *   const { caller, error } = await oathmesh(request);
 *   if (error) return error;
 *
 *   return NextResponse.json({
 *     subject: caller.principal.subject,
 *     action: caller.action,
 *   });
 * }
 * ```
 */
export function withOathMesh(config: VerifierConfig) {
  return async (
    request: Request
  ): Promise<
    | { caller: VerifiedCallerContext; error: null }
    | { caller: null; error: Response }
  > => {
    const token = extractToken(request.headers.get('authorization'));
    const headers = Object.fromEntries(request.headers.entries());

    if (!token) {
      const err = new OathMeshError(
        'claim_missing:token',
        'missing or invalid Authorization header',
        "provide a token in the format 'Authorization: OathMesh <token>'"
      );
      await config.onDenied?.(err, headers);
      return {
        caller: null,
        error: Response.json(err.toJSON(), { status: 401 }),
      };
    }

    try {
      const caller = await verifyOathToken(token, config);
      await config.onVerified?.(caller, headers);
      return { caller, error: null };
    } catch (e) {
      const err = e instanceof OathMeshError
        ? e
        : new OathMeshError('verification_failed', (e as Error).message, 'check token format');
      await config.onDenied?.(err, headers);
      return {
        caller: null,
        error: Response.json(err.toJSON(), { status: 401 }),
      };
    }
  };
}

// ─── Pages Router (API Routes) ─────────────────────────────────────────────

/** Minimal Next.js Pages Router request type (avoids hard next dependency). */
interface NextApiRequest {
  headers: Record<string, string | string[] | undefined>;
  body?: unknown;
  query?: Record<string, string | string[] | undefined>;
  method?: string;
}

/** Minimal Next.js Pages Router response type. */
interface NextApiResponse {
  status(code: number): NextApiResponse;
  json(body: unknown): void;
}

/** Next.js Pages Router API handler type. */
type NextApiHandler = (req: NextApiRequest, res: NextApiResponse) => void | Promise<void>;

/**
 * Wrap a Next.js Pages Router API handler with OathMesh verification.
 *
 * The verified caller context is injected into `req.oathmeshContext`.
 *
 * @example
 * ```typescript
 * // pages/api/inventory.ts
 * import { withOathMeshApi } from '@oathmesh/sdk/next';
 *
 * export default withOathMeshApi(
 *   {
 *     audience: 'https://inventory.internal',
 *     trustedIssuers: ['https://issuer.oathmesh.dev'],
 *   },
 *   (req, res) => {
 *     const caller = (req as any).oathmeshContext;
 *     res.json({ subject: caller.principal.subject });
 *   }
 * );
 * ```
 */
export function withOathMeshApi(
  config: VerifierConfig,
  handler: NextApiHandler
): NextApiHandler {
  return async (req, res) => {
    const authHeader = Array.isArray(req.headers.authorization)
      ? req.headers.authorization[0]
      : req.headers.authorization;
    const token = extractToken(authHeader);
    const headers = req.headers as Record<string, string | undefined>;

    if (!token) {
      const err = new OathMeshError(
        'claim_missing:token',
        'missing or invalid Authorization header',
        "provide a token in the format 'Authorization: OathMesh <token>'"
      );
      await config.onDenied?.(err, headers);
      res.status(401).json(err.toJSON());
      return;
    }

    try {
      const caller = await verifyOathToken(token, config);
      (req as any).oathmeshContext = caller;
      await config.onVerified?.(caller, headers);
      await handler(req, res);
    } catch (e) {
      const err = e instanceof OathMeshError
        ? e
        : new OathMeshError('verification_failed', (e as Error).message, 'check token format');
      await config.onDenied?.(err, headers);
      res.status(401).json(err.toJSON());
    }
  };
}

// ─── Edge Middleware ────────────────────────────────────────────────────────

/**
 * Create a Next.js Edge Middleware verifier.
 *
 * Returns a function you call inside your `middleware.ts`. If verification
 * fails it returns a `Response` you should return immediately. If it passes,
 * it returns `null` and you should call `NextResponse.next()`.
 *
 * @example
 * ```typescript
 * // middleware.ts (project root)
 * import { NextRequest, NextResponse } from 'next/server';
 * import { createEdgeVerifier } from '@oathmesh/sdk/next';
 *
 * const verify = createEdgeVerifier({
 *   audience: 'https://inventory.internal',
 *   trustedIssuers: ['https://issuer.oathmesh.dev'],
 * });
 *
 * export async function middleware(request: NextRequest) {
 *   const denied = await verify(request);
 *   if (denied) return denied;
 *
 *   // Verification passed — forward with injected headers
 *   return NextResponse.next();
 * }
 *
 * export const config = {
 *   matcher: '/api/:path*',
 * };
 * ```
 */
export function createEdgeVerifier(config: VerifierConfig) {
  return async (request: Request): Promise<Response | null> => {
    const token = extractToken(request.headers.get('authorization'));
    const headers = Object.fromEntries(request.headers.entries());

    if (!token) {
      const err = new OathMeshError(
        'claim_missing:token',
        'missing or invalid Authorization header',
        "provide a token in the format 'Authorization: OathMesh <token>'"
      );
      await config.onDenied?.(err, headers);
      return Response.json(err.toJSON(), { status: 401 });
    }

    try {
      const caller = await verifyOathToken(token, config);
      await config.onVerified?.(caller, headers);
      // Verification passed — middleware.ts should call NextResponse.next()
      // with injected X-OathMesh-* headers if desired
      return null;
    } catch (e) {
      const err = e instanceof OathMeshError
        ? e
        : new OathMeshError('verification_failed', (e as Error).message, 'check token format');
      await config.onDenied?.(err, headers);
      return Response.json(err.toJSON(), { status: 401 });
    }
  };
}
