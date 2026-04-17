# Token Format

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  <b>Oath Token structure, headers, and claims reference.</b>
</p>

---

> 📖 **New to OathMesh?** Start with [Concepts](../docs/concepts.md).

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
  "iss": "https://issuer.oathmesh.dev",
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
