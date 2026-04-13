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
}

/** Configuration for the OathMesh verification middleware. */
export interface VerifierConfig {
  /** The audience URL this receiver expects (exact match, no globs). */
  audience: string;
  /** Trusted issuer URLs (explicit allowlist — no wildcards, no auto-discovery). */
  trustedIssuers: string[];
}
