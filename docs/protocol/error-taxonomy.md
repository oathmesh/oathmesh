# Error Taxonomy

Every OathMesh error follows a structured format:

```json
{
  "error": "<code>",
  "message": "<human-readable cause>",
  "fix": "<actionable fix instruction>",
  "request_id": "<req-uuid>"
}
```

All errors return HTTP 401 Unauthorized.

## Error Codes

| Code | Verification Step | Trigger | Fix |
|---|---|---|---|
| `signature_invalid` | Step 06 | JWS signature verification failed | Check that the token was signed by a trusted issuer and the key has not been rotated |
| `issuer_untrusted` | Step 04, 07 | `iss` claim not in the receiver's explicit trusted issuers list | Add the issuer URL to the receiver's `trustedIssuers` configuration |
| `token_expired` | Step 08 | `time.Now() > exp + 10s` (past clock skew tolerance) | Mint a new token — Oath Tokens are short-lived and cannot be refreshed |
| `audience_mismatch` | Step 10 | `aud` does not exactly match the receiver's configured audience | Mint a new token with `--aud` set to the receiver's audience URL |
| `algorithm_not_allowed` | Step 02 | `alg` is `"none"`, `"HS256"`, or an unsupported algorithm | Use `EdDSA` (default) or `ES256` |
| `claim_missing:{claim}` | Step 11 | A required claim (`iss`, `sub`, `aud`, `act`, `iat`, `exp`, `jti`) is absent | Include all required claims when minting |
| `claim_missing:token` | Pre-Step 01 | No `Authorization` header or not in `OathMesh <token>` format | Add `Authorization: OathMesh <token>` header to the request |
| `replay_detected` | Step 13 | `jti` has been seen before within the token's TTL window | Mint a new token — each call requires a unique token |
| `policy_denied` | Step 14 | No policy rule matched, or an explicit deny rule matched | Update the Pkl policy to allow the caller's subject, action, and scope |
| `binding_mismatch` | Step 12 | `rqh` claim present but `sha256(canonical_request)` does not match | Ensure the request body matches the hash bound into the token at mint time |

## Error Response Examples

**Missing token:**
```json
{
  "error": "claim_missing:token",
  "message": "missing or invalid Authorization header",
  "fix": "provide a token in the format 'Authorization: OathMesh <token>'",
  "request_id": "req-abc-123"
}
```

**Audience mismatch:**
```json
{
  "error": "audience_mismatch",
  "message": "token was minted for https://billing.internal but received by https://inventory.internal",
  "fix": "mint a new token with aud set to https://inventory.internal",
  "request_id": "req-def-456"
}
```

**Replay detected:**
```json
{
  "error": "replay_detected",
  "message": "token jti 550e8400-... has already been used",
  "fix": "mint a new token for each request — Oath Tokens are single-use",
  "request_id": "req-ghi-789"
}
```

## Implementation Rules

- Never expose raw Go error strings externally — always map to a taxonomy code
- Every error includes a `fix` field with an actionable instruction
- `request_id` is included when available (from `X-Request-Id` header or generated)
- Errors are logged internally with full context; the external response contains only the taxonomy fields
