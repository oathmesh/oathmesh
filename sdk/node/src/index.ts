/**
 * @oathmesh/sdk — OathMesh verification SDK for Node.js / TypeScript
 *
 * Entry points:
 *   - `@oathmesh/sdk`       → Express middleware + core verifier
 *   - `@oathmesh/sdk/next`  → Next.js App Router, Pages Router, Edge Middleware
 *
 * @example Express
 * ```typescript
 * import { verifyToken } from '@oathmesh/sdk';
 *
 * app.use(verifyToken({
 *   audience: 'https://inventory.internal',
 *   trustedIssuers: ['https://issuer.oathmesh.dev'],
 * }));
 * ```
 *
 * @example Next.js App Router
 * ```typescript
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
 *   return NextResponse.json({ subject: caller.principal.subject });
 * }
 * ```
 */

// Express middleware
export { verifyToken } from './middleware';

// Core verifier (framework-agnostic)
export { verifyOathToken, extractToken } from './verify';

// Types
export {
  OathMeshError,
  type VerifiedCallerContext,
  type VerifierConfig,
  type Principal,
  type Source,
  type ErrorCode,
  type OathMeshErrorBody,
} from './types';
