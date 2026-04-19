# Issuer Configuration Reference

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `OATHMESH_ISSUER` | `http://localhost:4000` | Canonical issuer URL. Parsed as a URL at startup. Must use `https://` when `OATHMESH_ENV` is not `development`. |
| `OATHMESH_KMS_KEY_ID` | — | AWS KMS key ID/ARN for signing. If set, local private key env vars are not required. |
| `OATHMESH_PRIVATE_KEY` | — | Ed25519 private key in PKCS8 PEM format (full string including headers). |
| `OATHMESH_PRIVATE_KEY_B64` | — | Ed25519 private key Base64 encoded. Useful for CI/CD environments. |
| `OATHMESH_PRIVATE_KEY_FILE` | — | Path to PEM file. Used only when `OATHMESH_PRIVATE_KEY` and `OATHMESH_PRIVATE_KEY_B64` are unset. |
| `OATHMESH_PORT` | `4000` | HTTP listen port. |
| `OATHMESH_CONFIG_FILE` | — | Path to Pkl configuration file. |
| `OATHMESH_TTL_DEFAULT` | `120` | Default token TTL in seconds. |
| `OATHMESH_TTL_MAX` | `300` | Maximum token TTL (hard ceiling). |
| `OATHMESH_TTL_WRITE` | `60` | Recommended TTL for write/mutate actions. |
| `OATHMESH_RATE_LIMIT_RPM` | `100` | Requests per minute per source IP on mint endpoint. |
| `OATHMESH_RATE_LIMIT_BURST` | `20` | Burst size above steady-state rate. |
| `OATHMESH_AUDIT_SINK` | `stdout` | Audit output: `stdout`, `file`, or `noop` (tests only). |
| `OATHMESH_AUDIT_FILE` | — | File path when `OATHMESH_AUDIT_SINK=file`. |
| `OATHMESH_AUDIT_HMAC_KEY` | — | Optional HMAC key for audit chaining in gateway mode. |
| `OATHMESH_JWKS_CACHE_TTL` | `60` | JWKS in-memory cache TTL in seconds. |
| `OATHMESH_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error`. |
| `OATHMESH_ENV` | `development` | Environment mode. Any value other than `development` enables non-development validation paths. |
| `REDIS_URL` | — | Redis connection URL for replay cache. |
| `DATABASE_URL` | — | PostgreSQL connection URL (production). |

## Gateway Mode Variables

| Variable | Description |
|---|---|
| `OATHMESH_GATEWAY_UPSTREAM` | Upstream service URL to proxy to. |
| `OATHMESH_GATEWAY_AUDIENCE` | Expected audience for incoming tokens. |
| `OATHMESH_GATEWAY_ISSUERS` | Comma-separated list of trusted issuer URLs. |
| `OATHMESH_GATEWAY_POLICY` | Path to Pkl policy file for gateway verification. Required only when `oathmesh serve --gateway` and `OATHMESH_ENV` is not `development`. |

## Validation Rules (from runtime code)

- Startup fails if `OATHMESH_ISSUER` is not parseable as a URL.
- If `OATHMESH_ENV != development`, `OATHMESH_ISSUER` must use `https://`.
- Signing configuration must provide either:
  - `OATHMESH_KMS_KEY_ID`, or
  - one of `OATHMESH_PRIVATE_KEY`, `OATHMESH_PRIVATE_KEY_B64`, `OATHMESH_PRIVATE_KEY_FILE`.
- `OATHMESH_TTL_MAX` must be between `1` and `300`.
- In gateway mode, `OATHMESH_GATEWAY_UPSTREAM`, `OATHMESH_GATEWAY_AUDIENCE`, and `OATHMESH_GATEWAY_ISSUERS` are required.
- In gateway mode with `OATHMESH_ENV != development`, `OATHMESH_GATEWAY_POLICY` is required.

## Pkl Configuration Schema

The issuer supports configuration via a Pkl file (`internal/config/issuer.pkl`):

```pkl
class IssuerConfig {
  issuer: String                          // Canonical issuer URL
  port: Int(isBetween(1, 65535)) = 4000   // HTTP listen port
  privateKeySource: String("env"|"file"|"b64") = "env"
  privateKeyFile: String?                 // Path when source = "file"
  jwksCacheTTLSeconds: Int = 60
  keyRotation: KeyRotationConfig = new {}
  ttl: TTLConfig = new {}
  rateLimit: RateLimitConfig = new {}
  audit: AuditConfig = new {}
  replayCache: ReplayCacheConfig = new {}
}
```

## Key Management

| Parameter | Value |
|---|---|
| Algorithm | Ed25519 (ES256 only with ADR approval) |
| `kid` format | `"issuer-key-YYYY-MM"` |
| Rotation | New key published alongside old key in JWKS |
| Overlap | Both keys served for 24 hours (configurable) |
| Signing source | `OATHMESH_KMS_KEY_ID` or local key via `OATHMESH_PRIVATE_KEY` / `OATHMESH_PRIVATE_KEY_B64` / `OATHMESH_PRIVATE_KEY_FILE`. |
| JWKS | Public keys only. Private key material never appears. |

## Rate Limiting

| Parameter | Default |
|---|---|
| Algorithm | Token bucket per source IP |
| Rate | 100 requests per minute |
| Burst | 20 requests |
| Response | 429 Too Many Requests (no OathMesh error body) |
