/**
 * OathMesh SDK Types
 *
 * These types mirror the Go VerifiedCallerContext from internal/core/context.go.
 */

/** Authenticated identity of the caller. */
export interface Principal {
  /** Canonical issuer URL (e.g., "https://issuer.oathmesh.dev") */
  issuer: string;
  /** Subject URI — always a scheme: svc://, agent://, job://, tool://, user:// */
  subject: string;
}

/** Source provenance — where the call originated. */
export interface Source {
  type: string;
  repo?: string;
  workflow?: string;
  runId?: string;
  sha?: string;
}

/** The verified identity after successful token verification. */
export interface VerifiedCallerContext {
  principal: Principal;
  action: string;
  tokenId: string;
  environment: string;
  scope?: string[];
  reason?: string;
  source?: Source;
}

/** Machine-readable error code from the OathMesh error taxonomy. */
export type ErrorCode =
  | 'claim_missing:token'
  | 'claim_missing:iss'
  | 'claim_missing:sub'
  | 'claim_missing:aud'
  | 'claim_missing:act'
  | 'claim_missing:jti'
  | 'signature_invalid'
  | 'issuer_untrusted'
  | 'token_expired'
  | 'audience_mismatch'
  | 'algorithm_not_allowed'
  | 'replay_detected'
  | 'policy_denied'
  | 'binding_mismatch'
  | 'binding_required'
  | 'verification_failed';

/** Structured error returned on verification failure. */
export class OathMeshError extends Error {
  public readonly code: ErrorCode;
  public readonly fix?: string;

  constructor(code: ErrorCode, message: string, fix?: string) {
    super(message);
    this.name = 'OathMeshError';
    this.code = code;
    this.fix = fix;
  }

  /** Serialize to the standard OathMesh error JSON shape. */
  toJSON() {
    return {
      error: this.code,
      message: this.message,
      fix: this.fix,
    };
  }
}

/** Configuration for the OathMesh verifier. */
export interface VerifierConfig {
  /** The audience URL this receiver expects (exact match, no globs). */
  audience: string;
  /** Trusted issuer URLs (explicit allowlist — no wildcards, no auto-discovery). */
  trustedIssuers: string[];
  /**
   * Enforces that tokens MUST include an rqh claim.
   * When true, tokens without rqh are rejected with error "binding_required".
   * Recommended for all write/mutate endpoints to prevent tampering.
   * Default: false (for backward compatibility).
   */
  requireRequestBinding?: boolean;
  /**
   * Replay cache for preventing token reuse.
   * If provided, tokens with duplicate jti within TTL are rejected.
   * Use InMemoryReplayCache for development, or implement Redis-based cache for production.
   * Default: undefined (no replay checking).
   */
  replayCache?: ReplayCache;
  /**
   * Policy evaluator for authorization decisions.
   * If provided, token verification includes policy evaluation.
   * Use JsonPolicyEvaluator with a JSON policy document.
   * Default: undefined (no policy enforcement).
   */
  policyEvaluator?: PolicyEvaluator;
  /**
   * Called on every denied request. Use for logging, metrics, or alerting.
   * Runs after the error response is determined but before it is sent.
   */
  onDenied?: (err: OathMeshError, headers: Record<string, string | undefined>) => void | Promise<void>;
  /**
   * Called on every successful verification. Use for logging or metrics.
   */
  onVerified?: (ctx: VerifiedCallerContext, headers: Record<string, string | undefined>) => void | Promise<void>;
}

/** The JSON body shape returned on verification failure. */
export interface OathMeshErrorBody {
  error: ErrorCode;
  message: string;
  fix?: string;
  request_id?: string;
}

/**
 * Replay cache interface for preventing token reuse attacks.
 * Implementations can be in-memory (for single-instance) or Redis (for multi-instance).
 */
export interface ReplayCache {
  /**
   * Check if a token JTI has been seen before.
   * Returns true if the token has been replayed (already used).
   * Returns false if this is the first time seeing this JTI.
   */
  check(jti: string): boolean | Promise<boolean>;
  /**
   * Record a token JTI as seen.
   * Should be called after successful verification.
   */
  add(jti: string): void | Promise<void>;
}

/**
 * In-memory replay cache implementation for development/single-instance.
 * Uses a Map with TTL to automatically expire old entries.
 */
export class InMemoryReplayCache implements ReplayCache {
  private cache: Map<string, number> = new Map();
  private defaultTTL: number;

  constructor(defaultTTL: number = 300) {
    this.defaultTTL = defaultTTL;
  }

  check(jti: string): boolean {
    const expiresAt = this.cache.get(jti);
    if (expiresAt === undefined) {
      return false;
    }
    if (Date.now() > expiresAt) {
      this.cache.delete(jti);
      return false;
    }
    return true;
  }

  add(jti: string): void {
    this.cache.set(jti, Date.now() + this.defaultTTL * 1000);
  }
}

/**
 * Policy input for evaluation.
 */
export interface PolicyInput {
  iss: string;
  sub: string;
  aud: string;
  act: string;
  scope?: string[];
  env?: string;
}

/**
 * Policy decision result.
 */
export interface PolicyDecision {
  outcome: 'allow' | 'deny';
  ruleName?: string;
  denyReason?: string;
}

/**
 * Policy evaluator interface.
 * Implementations evaluate token claims against policy rules.
 */
export interface PolicyEvaluator {
  evaluate(input: PolicyInput): PolicyDecision | Promise<PolicyDecision>;
}

/**
 * JSON policy rule format.
 */
export interface JsonPolicyRule {
  match?: {
    sub?: string;
    aud?: string;
    act?: string;
    scope?: string[];
    env?: string;
  };
  allow: boolean;
  ruleName?: string;
  denyReason?: string;
}

/**
 * JSON policy document format.
 */
export interface JsonPolicyDocument {
  rules: JsonPolicyRule[];
}

/**
 * JSON policy evaluator that loads and evaluates simple JSON policies.
 */
export class JsonPolicyEvaluator implements PolicyEvaluator {
  private policy: JsonPolicyDocument;

  constructor(policy: JsonPolicyDocument) {
    this.policy = policy;
  }

  evaluate(input: PolicyInput): PolicyDecision {
    for (const rule of this.policy.rules) {
      if (this.matchesRule(input, rule)) {
        return {
          outcome: rule.allow ? 'allow' : 'deny',
          ruleName: rule.ruleName,
          denyReason: rule.denyReason,
        };
      }
    }
    // Default deny if no rules match
    return {
      outcome: 'deny',
      denyReason: 'no matching policy rule',
    };
  }

  private matchesRule(input: PolicyInput, rule: JsonPolicyRule): boolean {
    if (!rule.match) return false;
    const match = rule.match;

    if (match.sub && !this.matchPattern(input.sub, match.sub)) return false;
    if (match.aud && !this.matchPattern(input.aud, match.aud)) return false;
    if (match.act && !this.matchPattern(input.act, match.act)) return false;
    if (match.env && input.env !== match.env) return false;

    return true;
  }

  private matchPattern(value: string, pattern: string): boolean {
    if (pattern.includes('*')) {
      const regex = new RegExp('^' + pattern.replace(/\*/g, '.*') + '$');
      return regex.test(value);
    }
    return value === pattern;
  }
}
