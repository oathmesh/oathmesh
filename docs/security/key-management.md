# Key Management

## Algorithm

OathMesh uses **Ed25519** (via Go's `crypto/ed25519` stdlib package) as the primary signing algorithm. Ed25519 was chosen for:

- **Performance:** ~15,000 sign operations/second on commodity hardware
- **Key size:** 32-byte private key, 32-byte public key — compact JWKS
- **Security margin:** 128-bit security level, resistant to all known attacks
- **Deterministic signatures:** No nonce required, eliminating nonce-reuse vulnerabilities
- **stdlib availability:** No third-party dependency for the most critical security path

**ES256** (P-256 ECDSA) is accepted as a secondary algorithm for receivers that cannot verify Ed25519. New issuers should always use Ed25519. Switching to ES256 requires a new ADR.

## Key Identifier Format

```
issuer-key-YYYY-MM
```

Example: `issuer-key-2026-04`

The `kid` is included in every token header and every JWKS key entry. Receivers use it to locate the correct public key for signature verification.

## Key Loading

### Production

Load the private key from the `OATHMESH_PRIVATE_KEY` environment variable. The value is the full PEM PKCS8 string including headers:

```bash
export OATHMESH_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIJ...
-----END PRIVATE KEY-----"
```

### Development

For local development only, load from a file:

```bash
openssl genpkey -algorithm Ed25519 -out private.pem
export OATHMESH_PRIVATE_KEY_FILE=./private.pem
```

`OATHMESH_PRIVATE_KEY_FILE` is ignored if `OATHMESH_PRIVATE_KEY` is set.

### Security Rules

- Private keys are **never** hardcoded in source code
- Private keys are **never** logged under any circumstances
- Private keys are **never** returned in any HTTP response
- Private keys are **never** committed to version control
- `.gitignore` excludes `*.pem` and `*.key`

## Key Rotation

### Process

1. Generate a new Ed25519 key pair
2. Deploy the new key to the issuer
3. The issuer publishes **both** the new key and the old key in JWKS
4. During the overlap period (default: 24 hours), receivers can verify tokens signed with either key
5. After the overlap period, the old key is removed from JWKS

### During Rotation

```
JWKS contains:
  issuer-key-2026-04  ← new (signing)
  issuer-key-2026-03  ← old (verification only, deprecated)
```

### Cache Behavior

- JWKS is cached in memory with a default TTL of 5 minutes
- If a `kid` is not in cache, the verifier fetches JWKS once and retries
- If still missing after refresh: reject with `issuer_untrusted`

### Key Compromise Residual Window

After revoking a key, receivers that cached JWKS will accept tokens signed with the old key for up to `OATHMESH_JWKS_CACHE_TTL` seconds (default 300s, i.e., 5 minutes). This is a residual risk during emergency key compromise.

**For emergency revocation:**
1. Set `OATHMESH_JWKS_CACHE_TTL=0` on all receivers before rotating the key
2. Rotate the key (issuer publishes new key in JWKS)
3. Wait up to 300s for all in-flight tokens to expire
4. Restore `OATHMESH_JWKS_CACHE_TTL` to its normal value (300s)

Alternatively, restart all receiver processes to clear their JWKS cache immediately.

### KMS Guidance

For production deployments, store private keys in a Hardware Security Module (HSM) or Key Management Service (KMS):

- **AWS:** AWS KMS with Ed25519 support or Secrets Manager
- **GCP:** Cloud KMS or Secret Manager
- **Azure:** Azure Key Vault

The issuer reads the key at startup from the environment variable. Use your platform's secret injection mechanism to populate `OATHMESH_PRIVATE_KEY` from your KMS.
