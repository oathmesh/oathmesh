# Threat Model

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  <b>Security model, threats, and mitigations.</b>
</p>

---

> 🔒 **Report vulnerabilities** per our [Security Policy](../../SECURITY.md). Do NOT open public issues for security issues.

## Scope

This document covers threats to OathMesh as a machine-call identity system. It does not cover user authentication, browser security, or infrastructure-level attacks (those belong to your cloud provider's threat model).

## Threat Categories

### 1. Token Replay

**Threat:** An attacker intercepts a valid Oath Token and replays it to gain unauthorized access.

**Mitigations:**
- Every token has a unique `jti` (UUID, cryptographically random)
- Replay cache (memory or Redis) rejects duplicate `jti` values within the token's TTL window
- Tokens have a maximum TTL of 300 seconds, limiting the replay window
- Optional `rqh` (request hash binding) ties a token to a specific request body

**Residual risk:** If TLS is not enforced between caller and receiver, tokens can be intercepted in transit. Always use HTTPS in production.

### 2. Confused Deputy

**Threat:** A token minted for Service A is presented to Service B.

**Mitigations:**
- Step 10 enforces exact `aud` (audience) matching — no globs, no prefix matching
- Each receiver is configured with exactly one audience URL
- Audience mismatches are immediately rejected with `audience_mismatch`

**Residual risk:** None if receivers are correctly configured. Misconfigured audience URLs create a confused deputy vulnerability.

### 3. Key Compromise

**Threat:** An attacker obtains the issuer's private key and mints arbitrary tokens.

**Mitigations:**
- Private keys are loaded from environment variables, never hardcoded
- Private keys are never logged, never returned in any HTTP response
- JWKS endpoint serves public keys only
- Key rotation with overlap period allows immediate rotation on compromise
- Short TTLs (≤300s) limit the damage window from a compromised key

**Response:** On key compromise:
1. Generate a new key pair immediately
2. Deploy the new key to the issuer
3. The old key will expire from JWKS caches within 60 seconds (default cache TTL)
4. All tokens signed with the old key will expire within 300 seconds

### 4. Algorithm Confusion

**Threat:** An attacker crafts a token with `alg: "none"` or substitutes a symmetric algorithm to bypass signature verification.

**Mitigations:**
- Step 02 rejects `alg: "none"` immediately—no other step runs
- `alg` in the token header must match `alg` registered for the `kid` in JWKS
- Only `EdDSA` and `ES256` are permitted; `HS256`, short-key `RS256` are always rejected

### 5. Issuer Spoofing

**Threat:** An attacker sets up a malicious issuer and tricks receivers into trusting it.

**Mitigations:**
- Trusted issuers are an explicit allowlist — no wildcards, no auto-discovery
- `iss` claim is verified twice: once before JWKS fetch (Step 04) and once after signature verification (Step 07)
- JWKS is fetched only from trusted issuer URLs

### 6. Man-in-the-Middle (JWKS Fetch)

**Threat:** An attacker intercepts the JWKS fetch and substitutes malicious keys.

**Mitigations:**
- JWKS URLs should use HTTPS in production
- JWKS fetch uses a dedicated HTTP client with 5-second timeout (prevents slow-loris)
- JWKS is cached in memory, reducing fetch frequency

### 7. Denial of Service

**Threat:** An attacker floods the mint endpoint to exhaust resources.

**Mitigations:**
- Token bucket rate limiting per source IP (default: 100 rpm, burst 20)
- Rate limiter returns 429 before any token processing
- JWKS fetch timeout prevents slow upstream from blocking verification

## Security Invariants

These properties must hold at all times:

1. Private key material never appears in logs, responses, or JWKS
2. Full token strings never appear in logs
3. `alg: "none"` is always rejected at Step 02
4. Audience matching is always exact — never glob, prefix, or suffix
5. Default policy is always deny — cannot be disabled
6. Audit events fire on every verification — allow and deny
7. TTL never exceeds 300 seconds
8. `jti` is always cryptographically random (UUID)

---

## Related Documentation

| Document | Description |
|----------|-------------|
| [Replay Defense](replay-defense.md) | Detailed replay attack mitigation |
| [Key Management](key-management.md) | Private key handling best practices |
| [Logging Guidance](logging-guidance.md) | Secure logging practices |
| [Security Policy](../../SECURITY.md) | Reporting vulnerabilities |
