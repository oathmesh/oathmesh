# Issuer Configuration Reference

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `OATHMESH_ISSUER` | — | Canonical issuer URL. Must be HTTPS in production. |
| `OATHMESH_PRIVATE_KEY` | — | Ed25519 private key in PKCS8 PEM format (full string including headers). |
| `OATHMESH_PRIVATE_KEY_FILE` | — | Dev-only: path to PEM file. Ignored if `OATHMESH_PRIVATE_KEY` is set. |
| `OATHMESH_PORT` | `4000` | HTTP listen port. |
| `OATHMESH_CONFIG_FILE` | — | Path to Pkl configuration file. |
| `OATHMESH_TTL_DEFAULT` | `120` | Default token TTL in seconds. |
| `OATHMESH_TTL_MAX` | `300` | Maximum token TTL (hard ceiling). |
| `OATHMESH_TTL_WRITE` | `60` | Recommended TTL for write/mutate actions. |
| `OATHMESH_RATE_LIMIT_RPM` | `100` | Requests per minute per source IP on mint endpoint. |
| `OATHMESH_RATE_LIMIT_BURST` | `20` | Burst size above steady-state rate. |
| `OATHMESH_AUDIT_SINK` | `stdout` | Audit output: `stdout`, `file`, or `noop` (tests only). |
| `OATHMESH_AUDIT_FILE` | — | File path when `OATHMESH_AUDIT_SINK=file`. |
| `OATHMESH_JWKS_CACHE_TTL` | `300` | JWKS in-memory cache TTL in seconds. |
| `OATHMESH_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error`. |
| `REDIS_URL` | — | Redis connection URL for replay cache. |
| `DATABASE_URL` | — | PostgreSQL connection URL (production). |

## Gateway Mode Variables

| Variable | Description |
|---|---|
| `OATHMESH_GATEWAY_UPSTREAM` | Upstream service URL to proxy to. |
| `OATHMESH_GATEWAY_AUDIENCE` | Expected audience for incoming tokens. |
| `OATHMESH_GATEWAY_ISSUERS` | Comma-separated list of trusted issuer URLs. |
| `OATHMESH_GATEWAY_POLICY` | Path to Pkl policy file for gateway verification. |

## Pkl Configuration Schema

The issuer supports configuration via a Pkl file (`internal/config/issuer.pkl`):

```pkl
class IssuerConfig {
  issuer: String                          // Canonical issuer URL
  port: Int(isBetween(1, 65535)) = 4000   // HTTP listen port
  privateKeySource: String("env"|"file") = "env"
  privateKeyFile: String?                 // Path when source = "file"
  jwksCacheTTLSeconds: Int = 300
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
| Private key source | `OATHMESH_PRIVATE_KEY` env var (PEM PKCS8). Never hardcoded. |
| JWKS | Public keys only. Private key material never appears. |

## Rate Limiting

| Parameter | Default |
|---|---|
| Algorithm | Token bucket per source IP |
| Rate | 100 requests per minute |
| Burst | 20 requests |
| Response | 429 Too Many Requests (no OathMesh error body) |
