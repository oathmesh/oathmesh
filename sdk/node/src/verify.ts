/**
 * Core OathMesh token verification — framework-agnostic.
 *
 * This module contains the pure verification logic shared by all framework
 * adapters (Express, Next.js App Router, Next.js Pages Router, Edge Middleware).
 * It has no dependency on any HTTP framework.
 */

import { createHash } from 'crypto';
import { createRemoteJWKSet, jwtVerify, decodeProtectedHeader, decodeJwt } from 'jose';
import { OathMeshError, type VerifierConfig, type VerifiedCallerContext, type PolicyInput } from './types';

/**
 * Module-level JWKS cache. Keys are issuer URLs, values are jose JWKS functions.
 * Shared across all verifier instances within the same process to avoid
 * redundant JWKS fetches in serverless (Next.js) environments where
 * module scope persists across requests.
 */
const globalJWKSCache = new Map<string, ReturnType<typeof createRemoteJWKSet>>();
const SUBJECT_REGEX = /^(agent|svc|job|tool|user):\/\/[a-zA-Z0-9/_.-]{1,256}$/;
const CLOCK_SKEW_SECONDS = 10;
const MAX_EXP_UNIX = 4102444800; // 2100-01-01
type RequiredStringClaimName = 'iss' | 'sub' | 'aud' | 'act' | 'jti';
type RequiredNumberClaimName = 'iat' | 'exp';

function getJWKS(issuer: string): ReturnType<typeof createRemoteJWKSet> {
  let jwks = globalJWKSCache.get(issuer);
  if (!jwks) {
    const url = new URL('/.well-known/jwks.json', issuer);
    jwks = createRemoteJWKSet(url, { cacheMaxAge: 300_000 });
    globalJWKSCache.set(issuer, jwks);
  }
  return jwks;
}

/**
 * Extract the raw token string from an Authorization header value.
 *
 * Accepts:
 *   - `OathMesh <token>` (canonical)
 *
 * @returns The raw token string, or null if the header is missing/invalid.
 */
export function extractToken(authHeader: string | null | undefined): string | null {
  if (!authHeader) return null;
  if (authHeader.startsWith('OathMesh ')) return authHeader.slice(9);
  return null;
}

/**
 * Verify an OathMesh token string and return the verified caller context.
 *
 * This is the core verification function. Framework adapters call this
 * and handle HTTP responses themselves.
 *
 * @param token - Raw token string (without "OathMesh " prefix)
 * @param config - Verifier configuration
 * @returns Verified caller context on success
 * @throws OathMeshError on any verification failure
 */
export async function verifyOathToken(
  token: string,
  config: VerifierConfig
): Promise<VerifiedCallerContext> {
  const { audience, trustedIssuers } = config;
  const parts = token.split('.');
  if (parts.length !== 3) {
    throw new OathMeshError(
      'claim_missing',
      `invalid token format: expected 3 segments, got ${parts.length}`,
      'provide a valid OathMesh token in compact JWS format (header.payload.signature)'
    );
  }

  // Step 02: Decode and validate header
  let header: ReturnType<typeof decodeProtectedHeader>;
  try {
    header = decodeProtectedHeader(token);
  } catch {
    throw new OathMeshError('claim_missing', 'failed to decode token header', 'provide a valid base64url-encoded token header');
  }

  if (header.typ !== 'om+jwt') {
    throw new OathMeshError(
      'claim_missing',
      `token type "${header.typ}" is not valid — expected "om+jwt"`,
      'token typ must be om+jwt'
    );
  }
  if (header.alg === 'none') {
    throw new OathMeshError(
      'algorithm_not_allowed',
      'algorithm "none" is rejected — this is a security redline',
      'use EdDSA or ES256'
    );
  }
  if (!['EdDSA', 'ES256'].includes(header.alg!)) {
    throw new OathMeshError(
      'algorithm_not_allowed',
      `algorithm "${header.alg}" is not allowed`,
      'use EdDSA (preferred) or ES256'
    );
  }

  // Step 03–04: Extract issuer and check trust
  let claims: Record<string, unknown>;
  try {
    claims = decodeJwt(token) as Record<string, unknown>;
  } catch {
    throw new OathMeshError('claim_missing', 'failed to decode token payload', 'provide a valid base64url-encoded token payload');
  }

  const iss = claims.iss as string | undefined;
  if (!iss) {
    throw new OathMeshError('claim_missing:iss', 'missing iss claim', 'include iss when minting');
  }
  if (!trustedIssuers.includes(iss)) {
    throw new OathMeshError(
      'issuer_untrusted',
      `issuer "${iss}" is not in the trusted issuers list`,
      'add the issuer URL to your trustedIssuers configuration'
    );
  }

  // Step 11 (moved early): required claims.
  const sub = requiredStringClaim(claims, 'sub');
  const aud = requiredStringClaim(claims, 'aud');
  const act = requiredStringClaim(claims, 'act');
  requiredNumberClaim(claims, 'iat');
  requiredNumberClaim(claims, 'exp');
  const jti = requiredStringClaim(claims, 'jti');

  // Step 11.5: subject format validation.
  if (!SUBJECT_REGEX.test(sub)) {
    throw new OathMeshError(
      'claim_missing:sub',
      `invalid subject format: "${sub}"`,
      'subject must match schema (svc://, agent://, job://, tool://, user:// plus allowed chars)'
    );
  }

  // Step 05–06: Load JWKS and verify signature
  const jwks = getJWKS(iss);

  let payload: Record<string, unknown>;
  try {
    const result = await jwtVerify(token, jwks, {
      audience,
      issuer: iss,
      clockTolerance: CLOCK_SKEW_SECONDS,
    });
    payload = result.payload as Record<string, unknown>;
  } catch (err) {
    throw mapJoseError(err as Error, audience);
  }

  // Step 07: post-signature issuer trust re-check.
  if (!trustedIssuers.includes(iss)) {
    throw new OathMeshError('issuer_untrusted', `issuer "${iss}" failed post-signature trust check`, 'verify trustedIssuers configuration');
  }

  // Step 08-10: temporal and audience checks aligned with Go semantics.
  const now = Math.floor(Date.now() / 1000);
  const exp = Number(payload.exp);
  if (!Number.isFinite(exp) || exp > MAX_EXP_UNIX) {
    throw new OathMeshError('token_expired', 'token expiry is invalid or too far in the future', 'mint a token with sane exp');
  }
  if (now > exp + CLOCK_SKEW_SECONDS) {
    throw new OathMeshError('token_expired', 'token has expired', 'mint a new token — Oath Tokens are short-lived');
  }

  const iat = Number(payload.iat);
  if (!Number.isFinite(iat) || iat > now + CLOCK_SKEW_SECONDS) {
    throw new OathMeshError('token_expired', 'token issued-at is in the future', 'check clock synchronization between issuer and receiver');
  }

  if (payload.nbf !== undefined) {
    const nbf = Number(payload.nbf);
    if (Number.isFinite(nbf) && nbf > now + CLOCK_SKEW_SECONDS) {
      throw new OathMeshError('token_expired', 'token not-before is in the future', 'token cannot be used yet');
    }
  }

  if (payload.aud !== audience) {
    throw new OathMeshError(
      'audience_mismatch',
      `token was minted for "${String(payload.aud)}" but received by "${audience}"`,
      `mint with --aud ${audience}`
    );
  }

  // Ensure signed payload still carries required claims.
  if (!payload.sub) throw new OathMeshError('claim_missing:sub', 'missing sub claim', 'include sub when minting');
  if (!payload.act) throw new OathMeshError('claim_missing:act', 'missing act claim', 'include act when minting');
  if (!payload.jti) throw new OathMeshError('claim_missing:jti', 'missing jti claim', 'jti is auto-generated by the issuer');

  // Step 12: request hash binding.
  if (typeof payload.rqh === 'string' && payload.rqh !== '' && config.requestHash) {
    const expectedHash = `sha256:${createHash('sha256').update(config.requestHash).digest('hex')}`;
    if (payload.rqh !== expectedHash) {
      throw new OathMeshError(
        'binding_mismatch',
        `request hash mismatch: token has "${payload.rqh}" but request hash is "${expectedHash}"`,
        'ensure the request body has not been modified since the token was minted'
      );
    }
  }

  // Step 12b: Enforce rqh if RequireRequestBinding is set
  if (config.requireRequestBinding && !payload.rqh) {
    throw new OathMeshError(
      'binding_required',
      'token missing rqh (request hash) claim',
      'mint a token with rqh= sha256:<canonical-request> for write/mutate operations'
    );
  }

  // Step 13: Check replay cache (if configured)
  if (config.replayCache) {
    const replayed = await config.replayCache.check(payload.jti as string);
    if (replayed) {
      throw new OathMeshError(
        'replay_detected',
        `token ${payload.jti as string} has already been used`,
        'each Oath Token can only be used once — mint a new token'
      );
    }
    await config.replayCache.add(payload.jti as string);
  }

  // Step 13.5: optional revocation list.
  if (config.revocationList) {
    const revoked = await config.revocationList.isRevoked(payload.sub as string);
    if (revoked) {
      throw new OathMeshError(
        'subject_revoked',
        `subject ${payload.sub as string} has been revoked`,
        'mint a token for an active subject'
      );
    }
  }

  // Step 14: Evaluate policy (if configured)
  if (config.policyEvaluator) {
    const source = normalizeSource(payload.src);
    const policyInput: PolicyInput = {
      iss,
      sub: payload.sub as string,
      aud: aud,
      act: payload.act as string,
      scope: normalizeScope(payload.scope),
      env: payload.env as string | undefined,
      tenant: payload.tenant as string | undefined,
      srcType: source?.type,
      srcRepo: source?.repo,
      srcWflow: source?.workflow,
    };
    const decision = await config.policyEvaluator.evaluate(policyInput);
    if (decision.outcome === 'deny') {
      throw new OathMeshError(
        'policy_denied',
        decision.denyReason || 'policy evaluation denied',
        'check policy rules for this request'
      );
    }
  }

  // Build verified context
  const source = normalizeSource(payload.src);
  return {
    principal: {
      issuer: iss,
      subject: payload.sub as string,
    },
    action: payload.act as string,
    tokenId: jti,
    environment: (payload.env as string) || '',
    tenant: payload.tenant as string | undefined,
    scope: normalizeScope(payload.scope),
    reason: payload.reason as string | undefined,
    source,
  };
}

/**
 * Map a jose library error to the OathMesh error taxonomy.
 */
function mapJoseError(err: Error, audience: string): OathMeshError {
  const msg = err.message || '';
  if (msg.includes('expired') || err.name === 'JWTExpired') {
    return new OathMeshError('token_expired', 'token has expired', 'mint a new token — Oath Tokens are short-lived');
  }
  if (msg.includes('audience') || err.name === 'JWTClaimValidationFailed') {
    if (msg.includes('audience')) {
      return new OathMeshError('audience_mismatch', 'token audience does not match', `mint with --aud ${audience}`);
    }
    if (msg.includes('issuer')) {
      return new OathMeshError('issuer_untrusted', 'issuer verification failed', 'check trustedIssuers configuration');
    }
  }
  if (msg.includes('signature') || err.name === 'JWSSignatureVerificationFailed') {
    return new OathMeshError('signature_invalid', 'JWS signature verification failed', 'check that the token was signed by a trusted issuer');
  }
  if (msg.includes('issuer')) {
    return new OathMeshError('issuer_untrusted', 'issuer verification failed', 'check trustedIssuers configuration');
  }
  return new OathMeshError('verification_failed', msg || 'token verification failed', 'check token format and issuer availability');
}

function requiredStringClaim(claims: Record<string, unknown>, name: RequiredStringClaimName): string {
  const value = claims[name];
  if (typeof value !== 'string' || value === '') {
    throw new OathMeshError(`claim_missing:${name}`, `missing ${name} claim`, `include ${name} when minting`);
  }
  return value;
}

function requiredNumberClaim(claims: Record<string, unknown>, name: RequiredNumberClaimName): number {
  const value = claims[name];
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    throw new OathMeshError(`claim_missing:${name}`, `missing ${name} claim`, `include ${name} when minting`);
  }
  return value;
}

function normalizeScope(raw: unknown): string[] | undefined {
  if (!Array.isArray(raw)) {
    return undefined;
  }
  const scope = raw.filter((v): v is string => typeof v === 'string');
  return scope.length > 0 ? scope : undefined;
}

function normalizeSource(raw: unknown): VerifiedCallerContext['source'] | undefined {
  if (!raw || typeof raw !== 'object') {
    return undefined;
  }
  const src = raw as Record<string, unknown>;
  if (typeof src.type !== 'string' || src.type === '') {
    return undefined;
  }
  return {
    type: src.type,
    repo: typeof src.repo === 'string' ? src.repo : undefined,
    workflow: typeof src.workflow === 'string' ? src.workflow : undefined,
    runId: typeof src.run_id === 'string'
      ? src.run_id
      : (typeof src.runId === 'string' ? src.runId : undefined),
    sha: typeof src.sha === 'string' ? src.sha : undefined,
  };
}
