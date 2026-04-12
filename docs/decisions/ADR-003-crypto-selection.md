# ADR-003: Cryptographic Primitive Selection

## Status

Accepted

## Date

2026-04-12

## Context

OathMesh needs a signing algorithm that:
- Provides strong security for machine-to-machine authentication
- Has minimal implementation complexity
- Is widely supported by verifiers
- Avoids known weak algorithms (no HMAC, no small RSA keys)

## Decision

**Primary Algorithm**: EdDSA (Ed25519)
- Go standard library: `crypto/ed25519`
- No third-party JWT library needed
- Strong security: equivalent to ~128-bit symmetric key
- Fast verification
- Small signatures (64 bytes)

**Secondary Algorithm**: ES256 (ECDSA P-256)
- Accepted for receivers that cannot verify Ed25519
- MUST NOT be used for new issuers without a new ADR
- Go standard library: `crypto/ecdsa`

**Explicitly Forbidden**:
- `HS256` (HMAC-SHA256) — not suitable for identity
- `RS256` with key < 2048 bits
- `none` — always rejected at verification Step 02

### Key Format

- **Private key**: Ed25519 in PKCS8 PEM format
- **Load from**: `OATHMESH_PRIVATE_KEY` env var (PEM string with header/footer)
- **Dev fallback**: `OATHMESH_PRIVATE_KEY_FILE` path
- **kid format**: `issuer-key-YYYY-MM` (e.g., `issuer-key-2026-04`)

## Threat Analysis

| Threat | Mitigation |
|--------|------------|
| Key compromise | Short TTL (max 300s), key rotation |
| Algorithm confusion | alg in token header verified against JWKS key entry |
| Replay | jti + replay cache |
| Token theft | Short TTL, optional request binding |

## Consequences

- **Positive**: No CVE surface from crypto libraries
- **Positive**: Single implementation path for signing
- **Negative**: Some verifiers may only support ES256 initially

## Review

Security Agent approved crypto/ed25519 as signing primitive on 2026-04-12.
