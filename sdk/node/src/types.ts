/**
 * OathMesh SDK Type Definitions
 * 
 * These types match the Go VerifiedCallerContext struct
 * from internal/core/context.go.
 */

/**
 * The verified identity of the caller after successful token verification.
 */
export interface VerifiedCallerContext {
  principal: Principal;
  action: string;
  tokenId: string;
  environment: string;
  scope?: string[];
  reason?: string;
  source?: Source;
}

/**
 * The authenticated identity of the caller.
 */
export interface Principal {
  /** Canonical issuer URL */
  issuer: string;
  /** Subject URI (svc://, agent://, job://, tool://, user://) */
  subject: string;
}

/**
 * Source provenance — where the call originated.
 */
export interface Source {
  type: string;
  repo?: string;
  workflow?: string;
  runId?: string;
  sha?: string;
}

/**
 * Structured error returned by OathMesh verification.
 */
export interface OathMeshError {
  /** Machine-readable error code from the error taxonomy */
  code: string;
  /** Human-readable error message */
  message: string;
  /** Actionable fix instruction */
  fix?: string;
  /** Request ID for correlation */
  requestId?: string;
}

/**
 * Configuration for the OathMesh verification middleware.
 */
export interface VerifierConfig {
  /** The audience URL this receiver expects */
  audience: string;
  /** List of trusted issuer URLs */
  trustedIssuers: string[];
  /** Optional: path to Pkl policy file */
  policyPath?: string;
}
