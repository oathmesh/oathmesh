# Token Format

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  <b>Oath Token structure, headers, and claims reference.</b>
</p>

---

> 📖 **New to OathMesh?** Start with [Concepts](../concepts.md).

## Visual: Oath Token Structure

```text
Oath Token (Compact JWS Format)
═══════════════════════════════════════════════════════════════

  eyJ0eXAi...        .        eyJpc3Mi...        .        <signature>
  └─ HEADER ─┘               └─ PAYLOAD ─┘                └─ SIG ─┘
      (JSON)                     (JSON)                    (raw bytes
      Base64url                  Base64url                  base64url)

                            │ DECODED │

┌─────────────────────────────────────────────────────────────┐
│                        HEADER                               │
├─────────────────────────────────────────────────────────────┤
│ {                                                           │
│   "typ": "om+jwt"          ← Token type (MUST be om+jwt)  │
│   "alg": "EdDSA"           ← Signing algorithm             │
│   "kid": "issuer-2026-04"  ← Key ID for JWKS lookup        │
│ }                                                           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                       PAYLOAD (CLAIMS)                      │
├─────────────────────────────────────────────────────────────┤
│ {                                                           │
│   "iss": "https://issuer.oathmesh.tech"    ← Issuer        │
│   "sub": "agent://repo/acme/deploy-bot"   ← Subject        │
│   "aud": "https://inventory.internal"     ← Audience       │
│   "act": "inventory.write"                ← Action         │
│   "iat": 1774911000                       ← Issued-at      │
│   "exp": 1774911120                       ← Expiry (120s)  │
│   "jti": "550e8400-e29b-41d4-a716..."     ← Unique ID      │
│   "scope": ["inventory.read", "..."]      ← Permissions    │
│   "src": {...}                            ← Source context │
│   "env": "prod"                           ← Environment    │
│   "rqh": "sha256:e3b0c4..."               ← Request hash   │
│ }                                                           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      SIGNATURE                              │
├─────────────────────────────────────────────────────────────┤
│ Ed25519(base64url(header) + "." + base64url(payload))      │
│                                                             │
│ Verified at Step 06 using issuer's public key (JWKS)       │
│ Algorithm MUST match kid in JWKS (prevent algorithm        │
│ confusion attacks)                                          │
└─────────────────────────────────────────────────────────────┘
```

## Visual: Decode + Verify Path

```text
Authorization: OathMesh <header.payload.signature>
  |
  +--> split by "." into 3 segments
        |
        +--> decode header (typ, alg, kid)
        +--> decode payload (iss, sub, aud, act, iat, exp, jti, ...)
        +--> keep signature bytes
  |
  +--> use iss + kid to load JWKS key
  |
  +--> verify signature over:
        base64url(header) + "." + base64url(payload)
  |
  +--> continue with claim and policy checks
```

## Overview

An Oath Token is a signed compact JWS consisting of three base64url-encoded segments separated by `.`:

```
<header>.<payload>.<signature>
```

## Header

```json
{
  "typ": "om+jwt",
  "alg": "EdDSA",
  "kid": "issuer-key-2026-04"
}
```

| Field | Required | Value |
|---|---|---|
| `typ` | Yes | Must be `"om+jwt"`. Any other value is rejected at Step 02. |
| `alg` | Yes | `"EdDSA"` (primary). `"ES256"` accepted for compatibility. `"none"`, `"HS256"`, and `"RS256"` with keys <2048 bits are **always rejected**. |
| `kid` | Yes | Key identifier in format `"issuer-key-YYYY-MM"`. Used to locate the signing key in the issuer's JWKS. |

## Payload (Claims)

```json
{
  "iss": "https://issuer.oathmesh.tech",
  "sub": "agent://repo/acme/deploy-bot",
  "aud": "https://inventory.internal",
  "act": "inventory.write",
  "iat": 1774911000,
  "exp": 1774911120,
  "jti": "550e8400-e29b-41d4-a716-446655440000",
  "scope": ["inventory.read", "inventory.write"],
  "reason": "sync catalog after deploy",
  "src": {
    "type": "github_actions",
    "repo": "acme/storefront",
    "workflow": "deploy.yml",
    "run_id": "123456",
    "sha": "abc123def456"
  },
  "delegated_by": "user://mustafa",
  "env": "prod",
  "tenant": "acme-prod",
  "rqh": "sha256:e3b0c44298fc1c149afb"
}
```

See [Claim Reference](claim-reference.md) for details on each field.

## Signature

The signature is computed over `base64url(header) + "." + base64url(payload)` using the algorithm specified in the header (`EdDSA` by default). The signature is then base64url-encoded to form the third segment.

## Example

A complete Oath Token looks like:

```
eyJ0eXAiOiJvbStqd3QiLCJhbGciOiJFZERTQSIsImtpZCI6Imlzc3Vlci1rZXktMjAyNi0wNCJ9.eyJpc3MiOiJodHRwczovL2lzc3Vlci5vYXRobWVzaC5kZXYiLCJzdWIiOiJhZ2VudDovL3JlcG8vYWNtZS9kZXBsb3ktYm90IiwiYXVkIjoiaHR0cHM6Ly9pbnZlbnRvcnkuaW50ZXJuYWwiLCJhY3QiOiJpbnZlbnRvcnkud3JpdGUiLCJpYXQiOjE3NzQ5MTEwMDAsImV4cCI6MTc3NDkxMTEyMCwianRpIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIn0.<signature>
```

## HTTP Transport

The canonical transport header is:

```
Authorization: OathMesh <token>
```

**Note:** Legacy `Bearer` tokens are fully deprecated and will fail strict Step-01 header validation. This breaks backwards compatibility with legacy upstream integrations that failed to migrate.

---

## Related Documentation

| Document | Description |
|----------|-------------|
| [Verification Rules](verification-rules.md) | 14-step verification pipeline |
| [Claim Reference](claim-reference.md) | All claims detailed |
| [Error Taxonomy](error-taxonomy.md) | Error codes |
| [Audit Events](audit-events.md) | Audit logging |
