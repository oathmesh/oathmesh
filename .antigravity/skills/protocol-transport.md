---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
skill: "protocol-transport"
triggers:
  - "implementing HTTP transport for Oath Tokens"
  - "implementing gRPC transport"
  - "designing gateway mode"
  - "working on trust header normalization"
  - "implementing the metadata discovery endpoint"
dependencies:
  - "skills/auth.md (for token format and verification)"
  - "skills/api-design.md (for endpoint design patterns)"
  - "decisions/ADR-001-token-format.md (frozen transport decisions)"
---

# Skill: Protocol Transport & Gateway

This skill covers how Oath Tokens are transmitted between callers and receivers, how the gateway mode works, and how trust boundaries are managed at the network layer.

OathMesh is NOT a service mesh. It does not manage traffic routing, load balancing, circuit breaking, or connection pooling. It operates at the application layer for the sole purpose of transmitting and verifying signed identity.

## HTTP Transport

### Canonical Header

```
Authorization: Oathmesh eyJhbGciOiJFZERTQSIsInR5cCI6Im9tK2p3dCIsImtpZCI6Imlzc3Vlci1rZXktMjAyNi0wMSJ9...
```

The `Oathmesh` scheme is preferred. It explicitly distinguishes Oath Tokens from OAuth2 Bearer tokens, preventing middleware confusion.

### Bearer Compatibility Mode

```
Authorization: Bearer eyJhbGciOiJFZERTQSIsInR5cCI6Im9tK2p3dCJ9...
```

Only use Bearer mode when integrating with infrastructure that only supports `Authorization: Bearer`. The verifier must still check `typ: om+jwt` to distinguish Oath Tokens from standard OAuth2 JWTs.

Bearer compatibility is opt-in per receiver configuration:
```yaml
transport:
  accept_bearer: true  # default: false
```

### Token Extraction Priority

When the receiver extracts the token, check in this order:
1. `Authorization: Oathmesh <token>` — always checked first
2. `Authorization: Bearer <token>` — only if `accept_bearer: true`
3. No other locations (no query parameters, no cookies, no custom headers)

## gRPC Transport

Use gRPC metadata:
```
authorization: Oathmesh <token>
```

The verifier is implemented as a gRPC interceptor that extracts and verifies the token before the handler runs.

## Issuer Discovery

### Metadata Endpoint

Every issuer serves metadata at:
```
GET /.well-known/oathmesh-issuer
```

Response:
```json
{
  "issuer": "https://issuer.oathmesh.dev",
  "jwks_uri": "https://issuer.oathmesh.dev/.well-known/jwks.json",
  "alg_values_supported": ["EdDSA", "ES256"],
  "max_ttl_seconds": 300,
  "token_type": "om+jwt",
  "binding_modes_supported": ["none", "request-hash"],
  "version": "1.0"
}
```

### Requirements

- Served over HTTPS (except in local dev mode)
- `issuer` field must exactly match the `iss` claim in tokens — character-for-character, including trailing-slash presence/absence
- `jwks_uri` must be reachable from verifiers
- Response must include `Cache-Control` header (suggested: `max-age=3600`)
- Content-Type: `application/json`

## Gateway Mode

The OathMesh gateway is a lightweight reverse proxy that:

1. Receives incoming requests with `Authorization: Oathmesh <token>`
2. Verifies the token using the standard verification flow
3. Evaluates policy
4. If allowed, forwards the request to the upstream service
5. Strips the `Authorization` header from the forwarded request
6. Adds verified caller context as trusted internal headers
7. Emits an audit event

### Trust Header Format

After verification, the gateway adds these headers to the forwarded request:

```
X-Oathmesh-Subject: agent://repo/acme/deploy-bot
X-Oathmesh-Issuer: https://issuer.oathmesh.dev
X-Oathmesh-Action: inventory.write
X-Oathmesh-Token-Id: 550e8400-e29b-41d4-a716-446655440000
X-Oathmesh-Verified: true
```

### Trust Header Security

The gateway MUST:
- Strip any incoming `X-Oathmesh-*` headers before verification — prevent injection
- Only add `X-Oathmesh-*` headers after successful verification
- Use an HMAC-signed `X-Oathmesh-Signature` header that downstream services can verify to ensure headers came from a trusted gateway (shared HMAC key between gateway and upstream)
- Document this trust boundary in the deployment guide

Upstream services that sit behind the gateway can choose to:
1. Trust `X-Oathmesh-*` headers (simpler, requires network-level trust)
2. Verify the `X-Oathmesh-Signature` header (recommended for zero-trust internal networks)

### Gateway Configuration

```yaml
gateway:
  listen: ":8443"
  tls:
    cert_file: "/etc/oathmesh/tls/cert.pem"
    key_file: "/etc/oathmesh/tls/key.pem"
  upstream: "http://inventory-service:8080"
  policy_file: "/etc/oathmesh/policy.yaml"
  trusted_issuers:
    - issuer: "https://issuer.oathmesh.dev"
      jwks_uri: "https://issuer.oathmesh.dev/.well-known/jwks.json"
  trust_headers:
    hmac_key_env: "OATHMESH_GATEWAY_HMAC_KEY"
  audit:
    output: "stdout"
    format: "json"
```

## Request Binding

Optional mechanism for high-security scenarios. The caller computes a hash of the request and includes it in the token:

```
rqh = SHA-256(method + path + body)
```

The verifier recomputes the hash from the received request and compares. If they don't match, the token is rejected with `binding_mismatch`.

Implementation notes:
- Body hashing should use the raw bytes, not a parsed/re-serialized form
- The hash algorithm is always SHA-256 (not configurable in v1)
- The `rqh` claim format is `sha256:<hex-encoded-hash>`
- Request binding is opt-in per receiver configuration
