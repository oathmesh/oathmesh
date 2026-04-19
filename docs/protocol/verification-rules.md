# Verification Rules — The 14 Steps

<p align="center">
  <b>The mandatory verification pipeline every receiver must execute.</b>
</p>

---

> 📖 **New to OathMesh?** Start with [Concepts](../concepts.md) and [Token Format](token-format.md).

All receivers **must** perform all 14 verification steps in this exact order. No step may be skipped.

## Visual: 14-Step Pipeline

```text
Authorization: OathMesh <token>
  |
  +--> [01] Parse structure (3 segments)
  +--> [02] Validate header (typ/alg, reject alg=none)
  +--> [03] Decode payload (extract iss)
  +--> [04] Verify iss is trusted
  +--> [05] Fetch/load JWKS ({iss}/.well-known/jwks.json)
  +--> [06] Verify signature (kid + alg match key)
  +--> [07] Re-check iss after signature
  +--> [08] Check exp with 10s skew
  +--> [09] Check iat with 10s skew
  +--> [10] Check aud exact match
  +--> [11] Check required claims
  +--> [12] Check rqh hash binding (optional)
  +--> [13] Check replay cache (jti)
  +--> [13.5] Check revocation list (optional)
  +--> [14] Evaluate policy
            |
            +--> allow => 200 + audit.allow
            +--> deny  => 401 + audit.deny
```

## Visual: Verification Decision Gate

```text
Token + Request
   |
   v
[Structural checks 01-04]
   | pass
   v
[Cryptographic checks 05-07]
   | pass
   v
[Temporal + audience + claims checks 08-11]
   | pass
   v
[Binding + replay + revocation checks 12-13.5]
   | pass
   v
[Policy step 14]
   | allow                      | deny
   v                            v
200 (request proceeds)      401 (request blocked)
audit.allow emitted         audit.deny emitted
```

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
Fetch the issuer's JWKS from `{iss}/.well-known/jwks.json`. Use an in-memory cache with a default TTL of 60 seconds. If the `kid` from the token header is not in cache, fetch once. If still missing after refresh: reject with `issuer_untrusted`.

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
- **MemoryReplayCache:** 256-Shard Lock Table — atomic check-and-set eliminating global mutex contention and preventing TOCTOU replay bursts.
- **RedisReplayCache:** `SET jti EX <remaining_ttl> NX` — atomic check-and-set, no race

## Step 13.5 — Check Revocation (Optional)
If a RevocationList is configured, verify the `sub` against the list. If revoked, reject with `subject_revoked`.

## Step 14 — Evaluate Policy
Evaluate the Pkl policy rules in order dynamically resolving contexts including `sub`, `act`, `env`, and `tenant` boundaries. First matching rule wins. If no rule matches: deny. Emit an audit event regardless of outcome (allow **or** deny — this is never conditional).

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
| `binding_mismatch` | 12 | `rqh` present but hash does not match |
| `replay_detected` | 13 | `jti` seen before in replay cache |
| `subject_revoked` | 13.5 | Subject appears on active revocation list |
| `policy_denied` | 14 | No rule matched or explicit deny |
