// Package verify implements the OathMesh 14-step token verification pipeline.
//
// OathMesh authenticates the caller. The receiver authorizes the request.
//
// Every conformant receiver MUST execute all 14 verification steps in the
// order defined by the OathMesh protocol spec. Steps are ordered cheapest-first:
// structural/claim checks before network calls (JWKS fetch), signature verification
// after key retrieval, policy evaluation last.
package verify

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/sign"
)

// Verify performs all 14 OathMesh verification steps and returns a
// VerifiedCallerContext on success, or an OathMeshError on failure.
//
// Execution order (optimized for cheap-checks-first):
//
//	01 → parse structure
//	02 → alg check / none rejection
//	03 → decode payload, extract iss
//	04 → issuer allowlist check
//	11 → required claims present (before expensive JWKS fetch)
//	05 → JWKS fetch / cache
//	06 → signature verify + alg confusion check
//	07 → iss exact match post-sig
//	08 → exp check
//	09 → iat check
//	10 → aud check
//	12 → rqh binding
//	13 → replay cache
//	14 → policy stub
func Verify(ctx context.Context, token string, cfg *VerifierConfig) (*core.VerifiedCallerContext, error) {
	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	clockSkew := cfg.ClockSkew
	if clockSkew == 0 {
		clockSkew = 10 * time.Second
	}

	// ── Step 01: Parse structure ────────────────────────────────────────
	// Verify exactly three base64url segments separated by "."
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, emitAndReturn(ctx, cfg, nil, core.NewOathMeshError(
			core.ErrClaimMissing,
			fmt.Sprintf("invalid token format: expected 3 segments, got %d", len(parts)),
			"provide a valid OathMesh token in compact JWS format (header.payload.signature)",
		))
	}

	// ── Step 02: Decode and validate header ─────────────────────────────
	// typ MUST be "om+jwt"; alg MUST be in allowed list.
	// If alg is "none": REJECT immediately — do not proceed to any other step.
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, emitAndReturn(ctx, cfg, nil, core.NewOathMeshError(
			core.ErrClaimMissing,
			"failed to decode token header",
			"provide a valid base64url-encoded token header",
		))
	}

	var header sign.Header
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, emitAndReturn(ctx, cfg, nil, core.NewOathMeshError(
			core.ErrClaimMissing,
			"failed to parse token header JSON",
			"provide a valid JSON token header",
		))
	}

	// alg: "none" is always rejected immediately — security-critical
	if strings.EqualFold(header.Alg, "none") {
		return nil, emitAndReturn(ctx, cfg, nil, core.NewOathMeshError(
			core.ErrAlgorithmNotAllowed,
			"algorithm \"none\" is not allowed — this is a security violation",
			"use EdDSA (preferred) or ES256",
		))
	}

	if !AllowedAlgorithms[header.Alg] {
		return nil, emitAndReturn(ctx, cfg, nil, core.NewOathMeshError(
			core.ErrAlgorithmNotAllowed,
			fmt.Sprintf("algorithm %q is not allowed", header.Alg),
			"use EdDSA (preferred) or ES256",
		))
	}

	if header.Typ != sign.TypeHeader {
		return nil, emitAndReturn(ctx, cfg, nil, core.NewOathMeshError(
			core.ErrClaimMissing,
			fmt.Sprintf("token type %q is not valid — expected %q", header.Typ, sign.TypeHeader),
			"set typ header to \"om+jwt\"",
		))
	}

	// ── Step 03: Decode payload; extract iss claim ──────────────────────
	// Do not use header for issuer routing — always from payload.
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, emitAndReturn(ctx, cfg, nil, core.NewOathMeshError(
			core.ErrClaimMissing,
			"failed to decode token payload",
			"provide a valid base64url-encoded token payload",
		))
	}

	var claims sign.Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, emitAndReturn(ctx, cfg, nil, core.NewOathMeshError(
			core.ErrClaimMissing,
			"failed to parse token payload JSON",
			"provide a valid JSON token payload",
		))
	}

	// ── Step 04: Check iss against trusted issuers list ─────────────────
	// Explicit allowlist only — no wildcards, no auto-discovery.
	if !isTrustedIssuer(claims.Iss, cfg.TrustedIssuers) {
		return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
			core.ErrIssuerUntrusted,
			fmt.Sprintf("issuer %q is not in the trusted issuers list", claims.Iss),
			"add the issuer to the trusted issuers configuration or verify the token was minted by a known issuer",
		))
	}

	// ── Step 11 (moved): Verify all required claims present ─────────────
	// Fail fast before the expensive JWKS network fetch.
	// Required: iss, sub, aud, act, iat, exp, jti
	if err := checkRequiredClaims(&claims); err != nil {
		return nil, emitAndReturn(ctx, cfg, &claims, err)
	}

	// ── Step 05: Load JWKS from trusted issuer ──────────────────────────
	// Use in-memory cache (default TTL 5min, fetch timeout 5s).
	// Refresh on kid miss — if kid not in cache, fetch once.
	// If still missing after refresh: reject with issuer_untrusted.
	jwksProvider := cfg.JWKSProvider
	if jwksProvider == nil {
		jwksProvider = NewJWKSCache(DefaultJWKSCacheTTL)
	}

	publicKey, jwksAlg, err := jwksProvider.GetKey(claims.Iss, header.Kid)
	if err != nil {
		return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
			core.ErrIssuerUntrusted,
			fmt.Sprintf("failed to load key %q from issuer %q: %v", header.Kid, claims.Iss, err),
			"verify the issuer is running and serving JWKS at /.well-known/jwks.json",
		))
	}

	// ES256 deprecation warning
	if jwksAlg == "ES256" {
		log.Printf("WARN: ES256 key detected in JWKS. ES256 requires a secure nonce; Ed25519 is recommended. Plan migration to EdDSA. ES256 support will be removed in v2.0.")
	}

	// ── Step 06: Verify JWS signature ───────────────────────────────────
	// alg in token header MUST match alg registered for that kid in JWKS.
	// Algorithm confusion attack: reject if mismatch.
	if header.Alg != jwksAlg {
		return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
			core.ErrAlgorithmNotAllowed,
			fmt.Sprintf("algorithm confusion: token header says %q but JWKS key is registered as %q", header.Alg, jwksAlg),
			"ensure the token was signed with the algorithm registered for this key ID",
		))
	}

	signingInput := parts[0] + "." + parts[1]
	signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
			core.ErrSignatureInvalid,
			"failed to decode token signature",
			"provide a valid base64url-encoded signature",
		))
	}

	if !ed25519.Verify(publicKey, []byte(signingInput), signatureBytes) {
		return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
			core.ErrSignatureInvalid,
			"JWS signature verification failed",
			"ensure the token was signed with the correct private key matching the published JWKS",
		))
	}

	// ── Step 07: Verify iss claim exact string match ────────────────────
	// Post-signature check: confirm iss matches trusted issuer config.
	// This re-checks after signature validation to prevent payload tampering.
	if !isTrustedIssuer(claims.Iss, cfg.TrustedIssuers) {
		return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
			core.ErrIssuerUntrusted,
			fmt.Sprintf("issuer %q failed post-signature trust check", claims.Iss),
			"verify the token was issued by a trusted issuer",
		))
	}

	now := nowFn()

	// ── Step 08: Verify expiry ──────────────────────────────────────────
	// time.Now() < exp (clock skew tolerance: max 10 seconds)
	expTime := time.Unix(claims.Exp, 0)
	if now.After(expTime.Add(clockSkew)) {
		return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
			core.ErrTokenExpired,
			fmt.Sprintf("token expired at %s (current time: %s, skew tolerance: %s)",
				expTime.Format(time.RFC3339), now.Format(time.RFC3339), clockSkew),
			"mint a new token — Oath Tokens are intentionally short-lived",
		))
	}

	// ── Step 09: Verify issued-at ───────────────────────────────────────
	// iat <= time.Now() + 10s (reject future-issued tokens)
	iatTime := time.Unix(claims.Iat, 0)
	if iatTime.After(now.Add(clockSkew)) {
		return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
			core.ErrTokenExpired,
			fmt.Sprintf("token issued-at %s is too far in the future (current time: %s, skew tolerance: %s)",
				iatTime.Format(time.RFC3339), now.Format(time.RFC3339), clockSkew),
			"check clock synchronization between issuer and receiver",
		))
	}

	// ── Step 10: Verify audience ────────────────────────────────────────
	// aud exactly matches receiver's configured audience.
	// No glob, no prefix, no suffix — exact match only.
	if claims.Aud != cfg.Audience {
		return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
			core.ErrAudienceMismatch,
			fmt.Sprintf("token was minted for %q but received by %q", claims.Aud, cfg.Audience),
			fmt.Sprintf("mint a new token with aud set to %q", cfg.Audience),
		))
	}

	// ── Step 12: Verify request hash binding (if present) ───────────────
	// If rqh claim present: verify sha256(canonical_request) == rqh value.
	if claims.RQH != "" && cfg.RequestHash != "" {
		expectedHash := "sha256:" + sha256Hex(cfg.RequestHash)
		if claims.RQH != expectedHash {
			return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
				core.ErrBindingMismatch,
				fmt.Sprintf("request hash mismatch: token has %q but request hash is %q", claims.RQH, expectedHash),
				"ensure the request body has not been modified since the token was minted",
			))
		}
	}

	// ── Step 12b: Enforce rqh if RequireRequestBinding is set ───────────
	// If config requires binding but token has no rqh claim → reject.
	if cfg.RequireRequestBinding && claims.RQH == "" {
		return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
			core.ErrBindingRequired,
			"token missing rqh (request hash) claim",
			"mint a token with rqh= sha256:<canonical-request> for write/mutate operations",
		))
	}

	// ── Step 13: Check replay cache ─────────────────────────────────────
	// If jti seen before within token TTL window → reject immediately.
	if cfg.ReplayCache != nil {
		remainingTTL := expTime.Sub(now)
		if remainingTTL < time.Second {
			remainingTTL = time.Second
		}

		replayed, err := cfg.ReplayCache.Check(ctx, claims.JTI, remainingTTL)
		if err != nil {
			return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
				core.ErrReplayDetected,
				fmt.Sprintf("replay cache error: %v", err),
				"check replay cache backend connectivity",
			))
		}
		if replayed {
			return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
				core.ErrReplayDetected,
				fmt.Sprintf("token %s has already been used", claims.JTI),
				"mint a new token — each Oath Token can only be used once",
			))
		}
	}

	// ── Step 14: Evaluate policy ────────────────────────────────────────
	// First matching rule wins; if no rule matches → deny.
	if cfg.PolicyEvaluator != nil {
		policyInput := &PolicyInput{
			Iss:   claims.Iss,
			Sub:   claims.Sub,
			Aud:   claims.Aud,
			Act:   claims.Act,
			Scope: claims.Scope,
			Env:   claims.Env,
		}
		if claims.Src != nil {
			policyInput.SrcType = claims.Src.Type
			policyInput.SrcRepo = claims.Src.Repo
			policyInput.SrcWflow = claims.Src.Workflow
		}

		decision := cfg.PolicyEvaluator.Evaluate(policyInput)
		if decision.Outcome == "deny" {
			return nil, emitAndReturn(ctx, cfg, &claims, core.NewOathMeshError(
				core.ErrPolicyDenied,
				fmt.Sprintf("policy denied by rule %q", decision.RuleName),
				"check policy rules or request different permissions",
			))
		}
	}

	// Build VerifiedCallerContext
	var src *core.Source
	if claims.Src != nil {
		src = &core.Source{
			Type:     claims.Src.Type,
			Repo:     claims.Src.Repo,
			Workflow: claims.Src.Workflow,
			RunID:    claims.Src.RunID,
			SHA:      claims.Src.SHA,
		}
	}

	vcc := &core.VerifiedCallerContext{
		Principal: core.Principal{
			Issuer:  claims.Iss,
			Subject: claims.Sub,
		},
		Action:    claims.Act,
		Scope:     claims.Scope,
		Reason:    claims.Reason,
		Source:    src,
		TokenID:   claims.JTI,
		IssuedAt:  iatTime,
		ExpiresAt: expTime,
		Env:       claims.Env,
	}

	// Emit audit event — allow
	emitAudit(ctx, cfg, &claims, "allow", "verification_passed")

	return vcc, nil
}

// checkRequiredClaims verifies all required claims are present.
// Required: iss, sub, aud, act, iat, exp, jti.
func checkRequiredClaims(claims *sign.Claims) *core.OathMeshError {
	missing := []string{}

	if claims.Iss == "" {
		missing = append(missing, "iss")
	}
	if claims.Sub == "" {
		missing = append(missing, "sub")
	}
	if claims.Aud == "" {
		missing = append(missing, "aud")
	}
	if claims.Act == "" {
		missing = append(missing, "act")
	}
	if claims.Iat == 0 {
		missing = append(missing, "iat")
	}
	if claims.Exp == 0 {
		missing = append(missing, "exp")
	}
	if claims.JTI == "" {
		missing = append(missing, "jti")
	}

	if len(missing) > 0 {
		claim := missing[0]
		return core.NewOathMeshError(
			core.ErrorCode(fmt.Sprintf("%s:%s", core.ErrClaimMissing, claim)),
			fmt.Sprintf("required claim %q is missing", claim),
			fmt.Sprintf("include the %q claim in the token payload", claim),
		)
	}
	return nil
}

// isTrustedIssuer checks if the issuer is in the trusted list.
// Exact string match only — no wildcards, no auto-discovery.
func isTrustedIssuer(issuer string, trusted []string) bool {
	for _, t := range trusted {
		if t == issuer {
			return true
		}
	}
	return false
}

// sha256Hex returns the hex-encoded SHA-256 hash of the input.
func sha256Hex(input string) string {
	h := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", h)
}

// emitAudit sends an audit event to the configured sink.
// Called on every verification attempt — allow AND deny.
func emitAudit(ctx context.Context, cfg *VerifierConfig, claims *sign.Claims, outcome, reason string) {
	if cfg.AuditSink == nil {
		return
	}

	event := &core.AuditEvent{
		Event:     "oathmesh.verify",
		Outcome:   outcome,
		Reason:    reason,
		Timestamp: time.Now(),
	}

	if claims != nil {
		event.JTI = claims.JTI
		event.Sub = claims.Sub
		event.Aud = claims.Aud
		event.Act = claims.Act
		event.Iss = claims.Iss
		event.Env = claims.Env
		if claims.Src != nil {
			event.Source = &core.Source{
				Type:     claims.Src.Type,
				Repo:     claims.Src.Repo,
				Workflow: claims.Src.Workflow,
				RunID:    claims.Src.RunID,
				SHA:      claims.Src.SHA,
			}
		}
	}

	// Best-effort emit — don't fail verification due to audit sink error
	_ = cfg.AuditSink.Emit(ctx, event)
}

// emitAndReturn emits a deny audit event and returns the error.
// Used to ensure every verification failure is audited.
func emitAndReturn(ctx context.Context, cfg *VerifierConfig, claims *sign.Claims, err *core.OathMeshError) *core.OathMeshError {
	emitAudit(ctx, cfg, claims, "deny", string(err.Code)+": "+err.Message)
	return err
}
