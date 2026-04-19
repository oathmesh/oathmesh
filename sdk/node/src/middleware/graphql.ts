/**
 * Apollo Server v4 GraphQL Middleware
 *
 * Pre-execution authentication and rate limiting for GraphQL operations.
 * Integrates with OathMesh SDK for token verification and field authorization.
 */

import { verifyOathToken, OathMeshError } from '../index';
import type { VerifierConfig } from '../types';
import type { OathMeshGraphQLContext, OathMeshGraphQLConfig, RateLimitBucket } from './types';

/**
 * In-memory rate limiting store.
 * Maps subject (token.sub) to rate limit buckets.
 */
const rateLimitStore = new Map<string, RateLimitBucket>();

/**
 * Default rate limits (per minute, per subject).
 */
const DEFAULT_LIMITS = {
  queriesPerMinute: 100,
  mutationsPerMinute: 10,
};

/**
 * Check if the current request exceeds rate limits.
 * Returns error code or null if allowed.
 */
function checkRateLimit(
  subject: string,
  operationType: 'query' | 'mutation',
  limits: { queriesPerMinute?: number; mutationsPerMinute?: number }
): string | null {
  const now = Date.now();
  const limit = operationType === 'query'
    ? limits.queriesPerMinute ?? DEFAULT_LIMITS.queriesPerMinute
    : limits.mutationsPerMinute ?? DEFAULT_LIMITS.mutationsPerMinute;

  let bucket = rateLimitStore.get(subject);

  if (!bucket || now - bucket.windowStart >= 60000) {
    // Create new window
    bucket = {
      requestCounts: new Map(),
      windowStart: now,
    };
    rateLimitStore.set(subject, bucket);
  }

  const currentCount = bucket.requestCounts.get(operationType) || 0;

  if (currentCount >= limit) {
    return 'rate_limit_exceeded';
  }

  bucket.requestCounts.set(operationType, currentCount + 1);
  return null;
}

/**
 * Determine the operation type (query or mutation) from a GraphQL operation.
 */
function getOperationType(operation: any): 'query' | 'mutation' {
  if (!operation || !operation.operation) {
    return 'query'; // default to query
  }
  return operation.operation === 'mutation' ? 'mutation' : 'query';
}

/**
 * Create Apollo Server v4 middleware for OathMesh authentication.
 *
 * This middleware runs before GraphQL execution and:
 * 1. Extracts JWT from Authorization header
 * 2. Verifies token using OathMesh SDK
 * 3. Applies rate limiting per operation type
 * 4. Injects verified claims into GraphQL context
 *
 * @param config OathMesh GraphQL middleware configuration
 * @returns Apollo Server plugin for use in apollo.plugins
 *
 * @example
 * ```typescript
 * import { ApolloServer } from '@apollo/server';
 * import { createOathMeshMiddleware } from '@oathmesh/sdk/middleware/graphql';
 *
 * const server = new ApolloServer({
 *   typeDefs,
 *   resolvers,
 *   plugins: [
 *     createOathMeshMiddleware({
 *       verifier: { audience: 'https://api.example.com', trustedIssuers: [...] },
 *       rateLimits: { queriesPerMinute: 100, mutationsPerMinute: 10 },
 *     }),
 *   ],
 * });
 * ```
 */
export function createOathMeshMiddleware(config: OathMeshGraphQLConfig) {
  const verifierConfig: VerifierConfig = {
    audience: config.verifier.audience,
    trustedIssuers: config.verifier.trustedIssuers,
    requireRequestBinding: config.verifier.requireRequestBinding,
    replayCache: config.verifier.replayCache,
    policyEvaluator: config.verifier.policyEvaluator,
  };

  return {
    async didResolveOperation(requestContext: any) {
      const { request, contextValue } = requestContext;

      // Extract token from Authorization header
      const authHeader = request.http?.headers?.get?.('authorization') || '';
      const token = extractOathToken(authHeader);

      let verified = false;
      let claims: any = null;
      let subject: string | null = null;

      if (!token) {
        // Missing token — return 401
        throw new OathMeshError(
          'claim_missing:token',
          'missing or invalid Authorization header',
          'provide a token in the format "Authorization: OathMesh <token>"'
        );
      }

      try {
        // Verify token
        claims = await verifyOathToken(token, verifierConfig);
        verified = true;
        subject = claims.principal.subject;
      } catch (err) {
        // Verification failed
        if (err instanceof OathMeshError) {
          throw err;
        }
        throw new OathMeshError(
          'verification_failed',
          err instanceof Error ? err.message : String(err),
          'check token and issuer configuration'
        );
      }

      // Check rate limits
      if (subject) {
        const operationType = getOperationType(requestContext.operation);
        const rateLimitError = checkRateLimit(subject, operationType, config.rateLimits || {});

        if (rateLimitError) {
          if (config.onRateLimitExceeded) {
            config.onRateLimitExceeded(subject, operationType);
          }
          throw new OathMeshError(
            'rate_limit_exceeded',
            `rate limit exceeded for ${operationType} operations`,
            'wait before retrying'
          );
        }
      }

      // Inject context
      const context: OathMeshGraphQLContext = {
        claims,
        verified,
        rateLimit: {
          remaining: 0, // TODO: calculate from bucket
          resetAt: new Date(),
        },
      };

      contextValue.oathmesh = context;
    },
  };
}

/**
 * Extract OathMesh token from Authorization header.
 * Accepts "OathMesh <token>" format.
 */
function extractOathToken(authHeader: string): string | null {
  if (!authHeader) return null;
  if (authHeader.startsWith('OathMesh ')) {
    return authHeader.slice(9);
  }
  return null;
}

/**
 * Get the OathMesh context from a GraphQL resolver.
 * Use this in field resolvers to access verified claims.
 *
 * @example
 * ```typescript
 * const resolvers = {
 *   Query: {
 *     currentUser: async (_: any, __: any, context: any) => {
 *       const oathmesh = getOathMeshContext(context);
 *       console.log(`Verified subject: ${oathmesh.claims.principal.subject}`);
 *       return { id: '123' };
 *     }
 *   }
 * };
 * ```
 */
export function getOathMeshContext(context: any): OathMeshGraphQLContext | null {
  return context?.oathmesh || null;
}
