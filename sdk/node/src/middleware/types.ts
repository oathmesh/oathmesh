/**
 * GraphQL Middleware Types
 *
 * Type definitions for OathMesh GraphQL middleware integration.
 */

import type { VerifiedCallerContext } from '../types';

/**
 * OathMesh GraphQL context extension.
 * Attached to the GraphQL context during middleware execution.
 */
export interface OathMeshGraphQLContext {
  /** Verified caller claims from token verification. */
  claims: VerifiedCallerContext;
  /** Whether the token was successfully verified. */
  verified: boolean;
  /** Rate limit status. */
  rateLimit?: {
    /** Remaining requests in current window. */
    remaining: number;
    /** When the rate limit window resets. */
    resetAt: Date;
  };
}

/**
 * Configuration for OathMesh GraphQL middleware.
 */
export interface OathMeshGraphQLConfig {
  /** OathMesh verifier configuration. */
  verifier: {
    audience: string;
    trustedIssuers: string[];
    requireRequestBinding?: boolean;
    replayCache?: any;
    policyEvaluator?: any;
  };
  /** Rate limits per minute. */
  rateLimits?: {
    /** Queries per minute per subject. */
    queriesPerMinute?: number;
    /** Mutations per minute per subject. */
    mutationsPerMinute?: number;
  };
  /** Callback when rate limit is exceeded. */
  onRateLimitExceeded?: (subject: string, operationType: 'query' | 'mutation') => void;
}

/**
 * Rate limit bucket for a subject.
 * @internal
 */
export interface RateLimitBucket {
  requestCounts: Map<string, number>; // operation type -> count
  windowStart: number;
}
