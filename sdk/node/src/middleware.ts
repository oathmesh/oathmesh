/**
 * OathMesh Express middleware adapter.
 *
 * Wraps the core verifier for Express.js / Connect-compatible frameworks.
 */

import type { Request, Response, NextFunction } from 'express';
import { OathMeshError, type VerifierConfig, type VerifiedCallerContext } from './types';
import { extractToken, verifyOathToken } from './verify';

// Extend Express Request with oathmeshContext
declare global {
  namespace Express {
    interface Request {
      oathmeshContext?: VerifiedCallerContext;
    }
  }
}

/**
 * Creates an Express middleware that verifies OathMesh tokens.
 *
 * On success, populates `req.oathmeshContext` with the verified caller identity.
 * On failure, responds with 401 and a structured error body.
 *
 * @example
 * ```typescript
 * import express from 'express';
 * import { verifyToken } from '@oathmesh/sdk';
 *
 * const app = express();
 * app.use(verifyToken({
 *   audience: 'https://inventory.internal',
 *   trustedIssuers: ['https://issuer.oathmesh.tech'],
 * }));
 *
 * app.get('/inventory', (req, res) => {
 *   const caller = req.oathmeshContext!;
 *   res.json({ subject: caller.principal.subject });
 * });
 * ```
 */
export function verifyToken(config: VerifierConfig) {
  return async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    const token = extractToken(req.headers.authorization);

    if (!token) {
      const err = new OathMeshError(
        'claim_missing:token',
        'missing or invalid Authorization header',
        "provide a token in the format 'Authorization: OathMesh <token>'"
      );
      await config.onDenied?.(err, req.headers as Record<string, string | undefined>);
      res.status(401).json(err.toJSON());
      return;
    }

    try {
      const context = await verifyOathToken(token, config);
      req.oathmeshContext = context;
      await config.onVerified?.(context, req.headers as Record<string, string | undefined>);
      next();
    } catch (err) {
      const oathErr = err instanceof OathMeshError
        ? err
        : new OathMeshError('verification_failed', (err as Error).message, 'check token format');
      await config.onDenied?.(oathErr, req.headers as Record<string, string | undefined>);
      res.status(401).json(oathErr.toJSON());
    }
  };
}
