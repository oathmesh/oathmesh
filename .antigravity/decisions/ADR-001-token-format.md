---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
---

# ADR-001: Token Format — JWT/JWS with om+jwt Type

**Status:** Accepted

## Context

OathMesh requires a signed token format for transmitting short-lived machine-call assertions from callers to receivers. The token must be:
- Verifiable without contacting the issuer (local verification)
- Implementable across Go, Node.js, and Python
- Small enough to fit in HTTP headers (< 2KB)
- Supports asymmetric cryptography (issuer signs, anyone verifies)
- Compatible with existing JWKS infrastructure

Reference: `oathmesh.txt` sections 9.1–9.5

## Options Considered

### Option A: Standard JWT/JWS (RFC 7519/7515)

- Uses the most widely supported token format
- Massive library ecosystem across all target languages
- Well-understood security properties
- Supports EdDSA and ES256
- Cons: "yet another JWT" — may be confused with OAuth2 tokens

### Option B: PASETO (Platform-Agnostic Security Tokens)

- Opinionated, avoids JWT's algorithm confusion pitfalls
- Strong defaults
- Cons: much smaller library ecosystem, especially for EdDSA
- Cons: not as widely adopted — harder for third parties to implement
- Cons: less tooling for debugging (no jwt.io equivalent)

### Option C: Custom Binary Format

- Maximum control over size and structure
- Cons: requires custom parsers in every language
- Cons: no existing tooling, debuggers, or community knowledge
- Cons: massive implementation burden for v1

## Decision

We will use **JWT/JWS (RFC 7519/7515) with a custom type header `om+jwt`** because:

1. JWT libraries exist in every target language with mature, audited implementations
2. JWKS infrastructure is well-understood and can be reused
3. The custom `typ: om+jwt` header distinguishes Oath Tokens from standard OAuth2 JWTs, preventing middleware confusion
4. EdDSA support is available in all major JWT libraries (Go: stdlib, Node: jose, Python: PyJWT+cryptography)
5. jwt.io and similar tools can decode Oath Tokens for debugging

The `om+jwt` type header is FROZEN — it will not change.

### Signing Algorithm

- **Primary:** EdDSA (Ed25519) — fast, small signatures (64 bytes), no nonce needed
- **Secondary:** ES256 (ECDSA P-256) — for environments where Ed25519 is not available
- Symmetric algorithms (HS256, HS384, HS512) are permanently excluded — they require shared secrets, which defeats OathMesh's purpose

### Custom Claim: `act`

The `act` (action) claim is a required claim not present in standard JWT. It carries the requested operation family (e.g., `inventory.write`). This claim is FROZEN.

**Relationship to `scope`:** `act` is the primary action the caller is requesting. `scope`, if present, is the full set of permitted operations. `act` must be a member of `scope` when `scope` is present.

### Subject URI Schemes

Subject identifiers use OathMesh-specific URI schemes: `svc://`, `agent://`, `job://`, `tool://`, `user://`. These schemes are FROZEN.

## Consequences

### Positive
- Fastest possible time to a working prototype
- Developers can debug tokens with existing JWT tools
- JWKS caching is a solved problem
- Third parties can implement verifiers with any JWT library

### Negative
- OathMesh tokens look like JWTs from the outside, requiring the `om+jwt` type check to distinguish them
- JWT's flexibility means implementers must be careful to reject unexpected algorithms

### Risks
- Algorithm confusion: mitigated by explicit algorithm allowlists in verifier configuration
- Token confusion with OAuth2: mitigated by the `Oathmesh` authorization scheme and `om+jwt` type header

## References

- `oathmesh.txt` sections 9.1–9.5 (token format specification)
- RFC 7519: JSON Web Token (JWT)
- RFC 7515: JSON Web Signature (JWS)
- RFC 7517: JSON Web Key (JWK)
- RFC 8037: CFRG Elliptic Curve Diffie-Hellman (ECDH) and Signatures in JOSE (EdDSA)
- `skills/auth.md` (implementation details)
- `context/glossary.md` (term definitions)
