# Verification Rules — The 14 Steps

All receivers **must** perform all 14 verification steps in this exact order. No step may be skipped.

## Step 01 — Parse Structure
Verify the token contains exactly three base64url-encoded segments separated by `.`. Any other structure is immediately rejected.

## Step 02 — Decode and Validate Header
Decode the first segment as JSON. Verify:
- `typ` must be `"om+jwt"`
- `alg` must be in the allowed algorithms list (`EdDSA`, `ES256`)
- **If `alg` is `"none"`: REJECT immediately.** Do not proceed to any other step. This prevents algorithm downgrade attacks.

## Step 03 — Decode Payload
Decode the second segment as JSON. Extract the `iss` (issuer) claim from the payload. **Do not use the header for issuer routing** — the header is untrusted until signature verification completes.

## Step 04 — Check Issuer Against Trusted List
Verify `iss` appears in the receiver's explicitly configured trusted issuers list. No wildcards. No auto-discovery. No fallback. Unknown issuers are always rejected with `issuer_untrusted`.

## Step 05 — Load JWKS
Fetch the issuer's JWKS from `{iss}/.well-known/jwks.json`. Use an in-memory cache with a default TTL of 5 minutes. If the `kid` from the token header is not in cache, fetch once. If still missing after refresh: reject with `issuer_untrusted`.

**Implementation requirement:** Use a dedicated `http.Client` with `Timeout: 5 * time.Second`. Never use `http.DefaultClient` (it has no timeout).

## Step 06 — Verify JWS Signature
Using the public key identified by `kid` from the JWKS, verify the JWS signature. The `alg` in the token header **must** match the `alg` registered for that key in the JWKS — this prevents algorithm confusion attacks.

## Step 07 — Verify Issuer Claim
Verify the `iss` claim in the payload is an exact string match against the trusted issuer configuration. This re-checks after signature verification to prevent issuer spoofing.

## Step 08 — Verify Expiry
Verify `time.Now() < exp`. Clock skew tolerance: maximum 10 seconds. Tokens more than 10 seconds past expiry are rejected with `token_expired`.

## Step 09 — Verify Issued-At
Verify `iat <= time.Now() + 10s`. Reject future-issued tokens to prevent clock manipulation attacks.

## Step 10 — Verify Audience
Verify `aud` exactly matches the receiver's configured audience. No glob matching. No prefix matching. No suffix matching. Exact string comparison only. This prevents confused deputy attacks.

## Step 11 — Verify Required Claims
Verify all required claims are present: `iss`, `sub`, `aud`, `act`, `iat`, `exp`, `jti`. Any missing claim triggers rejection with `claim_missing:{claim}`.

## Step 12 — Verify Request Binding (Optional)
If the `rqh` claim is present: compute `sha256(canonical_request)` and verify it matches the `rqh` value. Mismatch triggers `binding_mismatch`. If `rqh` is absent, this step is a no-op.

## Step 13 — Check Replay Cache
Check whether `jti` has been seen before within the token's TTL window. If seen: reject immediately with `replay_detected`. If not seen: record the `jti` with a TTL equal to the remaining token lifetime.

Implementation options:
- **MemoryReplayCache:** `sync.RWMutex` — reads under `RLock`, writes under `Lock`
- **RedisReplayCache:** `SET jti EX <remaining_ttl> NX` — atomic check-and-set, no race

## Step 14 — Evaluate Policy
Evaluate the Pkl policy rules in order. First matching rule wins. If no rule matches: deny. Emit an audit event regardless of outcome (allow **or** deny — this is never conditional).

## Design Rationale

Steps are ordered by cost: cheapest structural checks first (1–4), signature verification after structure is validated (6), expensive policy evaluation last (14). This minimizes work on obviously invalid tokens.

## Error Code Quick Reference

| Error Code | Step | Trigger |
|---|---|---|
| `signature_invalid` | 06 | JWS signature verification failed |
| `issuer_untrusted` | 04, 07 | `iss` not in trusted issuers list |
| `token_expired` | 08 | `time.Now() > exp + 10s` |
| `audience_mismatch` | 10 | `aud` does not match configured audience |
| `algorithm_not_allowed` | 02 | `alg` not in allowed list |
| `claim_missing:{claim}` | 11 | Required claim absent |
| `replay_detected` | 13 | `jti` seen before in replay cache |
| `policy_denied` | 14 | No rule matched or explicit deny |
| `binding_mismatch` | 12 | `rqh` present but hash does not match |
