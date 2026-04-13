/**
 * @oathmesh/sdk — OathMesh verification middleware for Express.js
 *
 * @example
 * ```typescript
 * import { verifyToken } from '@oathmesh/sdk';
 *
 * app.use(verifyToken({
 *   audience: 'https://inventory.internal',
 *   trustedIssuers: ['https://issuer.oathmesh.dev'],
 * }));
 *
 * app.get('/inventory', (req, res) => {
 *   const caller = req.oathmeshContext!;
 *   res.json({ subject: caller.principal.subject });
 * });
 * ```
 */

export { verifyToken } from './middleware';
export {
  OathMeshError,
  type VerifiedCallerContext,
  type VerifierConfig,
  type Principal,
  type Source,
  type ErrorCode,
} from './types';
