---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
skill: "auth"
triggers:
  - "implementing token minting or signing"
  - "implementing token verification"
  - "implementing OIDC bootstrap"
  - "working on key management or rotation"
  - "designing caller authentication to the issuer"
  - "implementing JWKS endpoints"
dependencies:
  - "skills/identity-resolution.md (for subject URI formatting)"
  - "skills/data-modeling.md (for claim schemas)"
  - "decisions/ADR-001-token-format.md (frozen format decisions)"
  - "decisions/ADR-002-auth-strategy.md (auth design decisions)"
---

# Skill: Authentication & Token Lifecycle

This skill covers all aspects of Oath Token creation, signing, verification, and the authentication mechanisms that callers use to obtain tokens from the issuer.

## Core Concepts

### Oath Token Lifecycle

```
1. Caller authenticates to Issuer (bootstrap identity)
2. Caller requests token (specifying audience, action, source)
3. Issuer validates request and caller identity
4. Issuer mints JWT: sets claims, signs with private key
5. Issuer returns token to caller (TTL: 120s default, 300s max)
6. Caller attaches token: Authorization: Oathmesh <token>
7. Receiver extracts token from header
8. Verifier validates: signature → issuer → expiry → iat → audience → claims → binding → replay
9. Policy engine evaluates verified context against rules
10. Receiver allows/denies and emits audit event
```

### Signing

- **Preferred algorithm**: EdDSA (Ed25519) — fast, small signatures, no nonce
- **Fallback algorithm**: ES256 (ECDSA P-256) — wider library support
- **Key format**: JWK with `kid` for rotation support
- **Key storage**: Local file for dev, KMS/HSM for production
- **Never**: RSA (too large), HS256/HS384/HS512 (symmetric — defeats the purpose)

### JWT Structure

**Header** (always present):
```json
{
  "typ": "om+jwt",
  "alg": "EdDSA",
  "kid": "issuer-key-2026-01"
}
```

**Payload** (required claims):
```json
{
  "iss": "https://issuer.oathmesh.dev",
  "sub": "agent://repo/acme/deploy-bot",
  "aud": "https://inventory.internal",
  "act": "inventory.write",
  "iat": 1774911000,
  "exp": 1774911120,
  "jti": "550e8400-e29b-41d4-a716-446655440000"
}
```

The `act` claim is OathMesh-specific and not standard JWT. It carries the requested operation family. It must always be present. The `scope` optional claim, if present, is the full set of permitted operations — `act` should be a member of `scope`.

## Caller-to-Issuer Authentication

### GitHub Actions OIDC (Golden Path)

1. GitHub Actions workflow requests OIDC token from `https://token.actions.githubusercontent.com`
2. Workflow sends OIDC token to OathMesh Issuer's `/v1/token` endpoint
3. Issuer validates the GitHub OIDC token:
   - Verify signature against GitHub's JWKS
   - Verify `iss` is `https://token.actions.githubusercontent.com`
   - Verify `aud` matches the issuer's configured audience for GitHub
   - Extract claims: `repository`, `workflow`, `run_id`, `sha`, `ref`
4. Issuer constructs the `sub` claim from GitHub claims (e.g., `job://github-actions/{org}/{repo}/{workflow}`)
5. Issuer populates `src` with provenance from GitHub claims
6. Issuer mints and returns Oath Token

### Kubernetes ServiceAccount Token

1. Workload presents its Kubernetes ServiceAccount JWT
2. Issuer validates against the cluster's OIDC discovery endpoint
3. Issuer constructs `sub` claim (e.g., `svc://{namespace}/{service-account}`)
4. Issuer populates `src` with Kubernetes provenance

### Pre-Shared Client Credentials (Dev Mode Only)

1. For local development, the issuer accepts a client ID + client secret pair
2. Client credentials are configured in the issuer's local config file
3. This method is disabled by default in production builds
4. The `sub` claim is taken directly from the client configuration

## Verification Implementation

The verifier must execute these checks in this exact order. Failing at any step produces a specific error code and short-circuits — no further checks are performed.

| Step | Check | Error Code | HTTP Status |
|---|---|---|---|
| 1 | Parse token (valid JWT structure) | `invalid_token` | 401 |
| 2 | Extract `iss` claim | `missing_issuer` | 401 |
| 3 | Look up issuer in trusted issuers list | `untrusted_issuer` | 401 |
| 4 | Fetch/refresh JWKS for issuer | `jwks_unavailable` | 503 |
| 5 | Match `kid` and verify signature | `invalid_signature` | 401 |
| 6 | Verify `alg` is in allowed algorithms | `disallowed_algorithm` | 401 |
| 7 | Verify `iss` exact match with trusted config | `issuer_mismatch` | 401 |
| 8 | Verify `exp` > current time | `token_expired` | 401 |
| 9 | Verify `iat` ≤ current time (with clock skew tolerance) | `token_not_yet_valid` | 401 |
| 10 | Verify `aud` matches receiver identity | `audience_mismatch` | 403 |
| 11 | Verify required claims present (`sub`, `act`, `jti`) | `missing_required_claim` | 401 |
| 12 | Verify `rqh` binding if present | `binding_mismatch` | 403 |
| 13 | Check `jti` against replay cache (if enabled) | `replay_detected` | 403 |
| 14 | Evaluate local policy | `policy_denied` | 403 |

### Clock Skew

Allow configurable clock skew tolerance: default 30 seconds, max 60 seconds. Applied to both `iat` and `exp` checks.

### JWKS Caching

- Cache JWKS for 5 minutes by default (configurable)
- On `kid` miss: force-refresh JWKS once, then fail — never retry infinitely
- Respect `Cache-Control` headers from JWKS endpoint
- Cache must support concurrent reads during refresh

## Key Rotation Protocol

1. Generate new key pair, assign new `kid` (format: `issuer-key-YYYY-MM`)
2. Add new public key to JWKS endpoint alongside existing key(s)
3. Wait at least 2× JWKS cache TTL for propagation (default: 10 minutes)
4. Switch signing to new private key
5. Keep old public key in JWKS for at least max-TTL (300 seconds) after last token signed with it
6. Remove old public key from JWKS
7. Securely destroy old private key

This requires human approval per `rules/security-redlines.md` Category 1.
