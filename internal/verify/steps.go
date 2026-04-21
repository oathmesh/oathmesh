package verify

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/metrics"
	"github.com/oathmesh/oathmesh/internal/sign"
)

var subjectRegex = regexp.MustCompile(`^(agent|svc|job|tool|user)://[a-zA-Z0-9/_.-]{1,256}$`)

// vctx holds all mutable state accumulated across the verification pipeline.
// Each step reads from and writes to this context. The struct is stack-allocated
// per Verify() call — no pooling, no reuse, no partial-zero bugs.
type vctx struct {
	ctx       context.Context
	token     string
	parts     []string
	header    sign.Header
	claims    sign.Claims
	publicKey ed25519.PublicKey
	jwksAlg   string
	now       time.Time
	clockSkew time.Duration
	expTime   time.Time
	iatTime   time.Time
	cfg       *VerifierConfig
	nowFn     func() time.Time
}

// pipelineStep pairs a step function with its metadata for the loop in Verify().
type pipelineStep struct {
	name string
	step int
	fn   func(v *vctx) *core.OathMeshError
}

// pipeline defines the fixed 14-step verification sequence.
// Steps are ordered cheapest-first: structural/claim checks before
// network calls (JWKS), signature verification after key retrieval,
// policy evaluation last. Step numbers match the OathMesh protocol spec.
var pipeline = []pipelineStep{
	{"parse_structure", 1, stepParseStructure},
	{"validate_header", 2, stepValidateHeader},
	{"decode_payload", 3, stepDecodePayload},
	{"check_issuer_pre", 4, stepCheckIssuerPre},
	{"required_claims", 11, stepRequiredClaims},
	{"subject_format", 11, stepSubjectFormat},
	{"load_jwks", 5, stepLoadJWKS},
	{"verify_signature", 6, stepVerifySignature},
	{"check_issuer_post", 7, stepCheckIssuerPost},
	{"check_expiry", 8, stepCheckExpiry},
	{"check_timing", 9, stepCheckTiming},
	{"check_audience", 10, stepCheckAudience},
	{"check_binding", 12, stepCheckBinding},
	{"check_replay", 13, stepCheckReplay},
	{"evaluate_policy", 14, stepEvaluatePolicy},
}

// ── Step 01: Parse structure ────────────────────────────────────────
func stepParseStructure(v *vctx) *core.OathMeshError {
	v.parts = strings.Split(v.token, ".")
	if len(v.parts) != 3 {
		return core.NewOathMeshErrorAt(1,
			core.ErrClaimMissing,
			fmt.Sprintf("invalid token format: expected 3 segments, got %d", len(v.parts)),
			"provide a valid OathMesh token in compact JWS format (header.payload.signature)",
		)
	}
	return nil
}

// ── Step 02: Decode and validate header ─────────────────────────────
func stepValidateHeader(v *vctx) *core.OathMeshError {
	headerJSON, err := base64.RawURLEncoding.DecodeString(v.parts[0])
	if err != nil {
		return core.NewOathMeshErrorAt(2,
			core.ErrClaimMissing,
			"failed to decode token header",
			"provide a valid base64url-encoded token header",
		)
	}

	if err := json.Unmarshal(headerJSON, &v.header); err != nil {
		return core.NewOathMeshErrorAt(2,
			core.ErrClaimMissing,
			"failed to parse token header JSON",
			"provide a valid JSON token header",
		)
	}

	// alg: "none" is always rejected immediately — security-critical
	if strings.EqualFold(v.header.Alg, "none") {
		return core.NewOathMeshErrorAt(2,
			core.ErrAlgorithmNotAllowed,
			"algorithm \"none\" is not allowed — this is a security violation",
			"use EdDSA (preferred) or ES256",
		)
	}

	if !AllowedAlgorithms[v.header.Alg] {
		return core.NewOathMeshErrorAt(2,
			core.ErrAlgorithmNotAllowed,
			fmt.Sprintf("algorithm %q is not allowed", v.header.Alg),
			"use EdDSA (preferred) or ES256",
		)
	}

	if v.header.Typ != sign.TypeHeader {
		return core.NewOathMeshErrorAt(2,
			core.ErrClaimMissing,
			fmt.Sprintf("token type %q is not valid — expected %q", v.header.Typ, sign.TypeHeader),
			"set typ header to \"om+jwt\"",
		)
	}

	return nil
}

// ── Step 03: Decode payload ─────────────────────────────────────────
func stepDecodePayload(v *vctx) *core.OathMeshError {
	payloadJSON, err := base64.RawURLEncoding.DecodeString(v.parts[1])
	if err != nil {
		return core.NewOathMeshErrorAt(3,
			core.ErrClaimMissing,
			"failed to decode token payload",
			"provide a valid base64url-encoded token payload",
		)
	}

	if err := json.Unmarshal(payloadJSON, &v.claims); err != nil {
		return core.NewOathMeshErrorAt(3,
			core.ErrClaimMissing,
			"failed to parse token payload JSON",
			"provide a valid JSON token payload",
		)
	}
	return nil
}

// ── Step 04: Check issuer (pre-signature) ───────────────────────────
func stepCheckIssuerPre(v *vctx) *core.OathMeshError {
	if !isTrustedIssuer(v.claims.Iss, v.cfg.TrustedIssuers) {
		return core.NewOathMeshErrorAt(4,
			core.ErrIssuerUntrusted,
			fmt.Sprintf("issuer %q is not in the trusted issuers list", v.claims.Iss),
			"add the issuer to the trusted issuers configuration or verify the token was minted by a known issuer",
		)
	}
	return nil
}

// ── Step 11: Required claims ────────────────────────────────────────
func stepRequiredClaims(v *vctx) *core.OathMeshError {
	if err := checkRequiredClaims(&v.claims); err != nil {
		err.Step = 11
		return err
	}
	return nil
}

// ── Step 11.5: Subject format ───────────────────────────────────────
func stepSubjectFormat(v *vctx) *core.OathMeshError {
	if !subjectRegex.MatchString(v.claims.Sub) {
		return core.NewOathMeshErrorAt(11,
			core.ErrorCode(fmt.Sprintf("%s:sub", core.ErrClaimMissing)),
			fmt.Sprintf("invalid subject format: %q", v.claims.Sub),
			"subject must match schema (e.g. svc://, agent://, user:// followed by allowed chars)",
		)
	}
	return nil
}

// ── Step 05: Load JWKS ──────────────────────────────────────────────
func stepLoadJWKS(v *vctx) *core.OathMeshError {
	jwksProvider := v.cfg.JWKSProvider
	if jwksProvider == nil {
		jwksProvider = NewJWKSCache(DefaultJWKSCacheTTL, nil)
	}
	if jwksCache, ok := jwksProvider.(*JWKSCache); ok {
		if err := jwksCache.RegisterTrustedIssuers(v.cfg.TrustedIssuers); err != nil {
			return core.NewOathMeshErrorAt(5,
				core.ErrIssuerUntrusted,
				fmt.Sprintf("failed to register trusted issuer JWKS endpoints: %v", err),
				"ensure trusted issuers are valid absolute URLs with scheme and host",
			)
		}
	}

	pubKey, alg, err := jwksProvider.GetKey(v.claims.Iss, v.header.Kid)
	if err != nil {
		return core.NewOathMeshErrorAt(5,
			core.ErrIssuerUntrusted,
			fmt.Sprintf("failed to load key %q from issuer %q: %v", v.header.Kid, v.claims.Iss, err),
			"verify the issuer is running and serving JWKS at /.well-known/jwks.json",
		)
	}
	v.publicKey = pubKey
	v.jwksAlg = alg

	// ES256 deprecation warning
	if alg == "ES256" {
		log.Printf("WARN: ES256 key detected in JWKS. ES256 requires a secure nonce; Ed25519 is recommended. Plan migration to EdDSA. ES256 support will be removed in v2.0.")
	}
	return nil
}

// ── Step 06: Verify JWS signature ───────────────────────────────────
func stepVerifySignature(v *vctx) *core.OathMeshError {
	if v.header.Alg != v.jwksAlg {
		return core.NewOathMeshErrorAt(6,
			core.ErrAlgorithmNotAllowed,
			fmt.Sprintf("algorithm confusion: token header says %q but JWKS key is registered as %q", v.header.Alg, v.jwksAlg),
			"ensure the token was signed with the algorithm registered for this key ID",
		)
	}

	signingInput := v.parts[0] + "." + v.parts[1]
	signatureBytes, err := base64.RawURLEncoding.DecodeString(v.parts[2])
	if err != nil {
		return core.NewOathMeshErrorAt(6,
			core.ErrSignatureInvalid,
			"failed to decode token signature",
			"provide a valid base64url-encoded signature",
		)
	}

	if !ed25519.Verify(v.publicKey, []byte(signingInput), signatureBytes) {
		return core.NewOathMeshErrorAt(6,
			core.ErrSignatureInvalid,
			"JWS signature verification failed",
			"ensure the token was signed with the correct private key matching the published JWKS",
		)
	}
	return nil
}

// ── Step 07: Verify issuer (post-signature) ─────────────────────────
func stepCheckIssuerPost(v *vctx) *core.OathMeshError {
	if !isTrustedIssuer(v.claims.Iss, v.cfg.TrustedIssuers) {
		return core.NewOathMeshErrorAt(7,
			core.ErrIssuerUntrusted,
			fmt.Sprintf("issuer %q failed post-signature trust check", v.claims.Iss),
			"verify the token was issued by a trusted issuer",
		)
	}
	v.now = v.nowFn()
	return nil
}

// ── Step 08: Verify expiry ──────────────────────────────────────────
func stepCheckExpiry(v *vctx) *core.OathMeshError {
	if v.claims.Exp > 4102444800 { // Year 2100
		return core.NewOathMeshErrorAt(8,
			core.ErrTokenExpired,
			"token expiry is artificially too far in the future",
			"mint a token with a sane expiry (before year 2100)",
		)
	}
	v.expTime = time.Unix(v.claims.Exp, 0)
	if v.now.After(v.expTime.Add(v.clockSkew)) {
		metrics.ClockSkewRejections.WithLabelValues("expired").Inc()
		return core.NewOathMeshErrorAt(8,
			core.ErrTokenExpired,
			fmt.Sprintf("token expired at %s (current time: %s, skew tolerance: %s)",
				v.expTime.Format(time.RFC3339), v.now.Format(time.RFC3339), v.clockSkew),
			"mint a new token — Oath Tokens are intentionally short-lived",
		)
	}
	return nil
}

// ── Step 09: Verify timing (iat, nbf) ───────────────────────────────
func stepCheckTiming(v *vctx) *core.OathMeshError {
	v.iatTime = time.Unix(v.claims.Iat, 0)
	if v.iatTime.After(v.now.Add(v.clockSkew)) {
		metrics.ClockSkewRejections.WithLabelValues("future_iat").Inc()
		return core.NewOathMeshErrorAt(9,
			core.ErrTokenExpired,
			fmt.Sprintf("token issued-at %s is too far in the future (current time: %s, skew tolerance: %s)",
				v.iatTime.Format(time.RFC3339), v.now.Format(time.RFC3339), v.clockSkew),
			"check clock synchronization between issuer and receiver",
		)
	}

	if v.claims.Nbf != 0 {
		nbfTime := time.Unix(v.claims.Nbf, 0)
		if nbfTime.After(v.now.Add(v.clockSkew)) {
			metrics.ClockSkewRejections.WithLabelValues("future_nbf").Inc()
			return core.NewOathMeshErrorAt(9,
				core.ErrTokenExpired,
				fmt.Sprintf("token not-before %s is in the future (current time: %s, skew tolerance: %s)",
					nbfTime.Format(time.RFC3339), v.now.Format(time.RFC3339), v.clockSkew),
				"token cannot be used yet",
			)
		}
	}
	return nil
}

// ── Step 10: Verify audience ────────────────────────────────────────
func stepCheckAudience(v *vctx) *core.OathMeshError {
	if v.claims.Aud != v.cfg.Audience {
		return core.NewOathMeshErrorAt(10,
			core.ErrAudienceMismatch,
			fmt.Sprintf("token was minted for %q but received by %q", v.claims.Aud, v.cfg.Audience),
			fmt.Sprintf("mint a new token with aud set to %q", v.cfg.Audience),
		)
	}
	return nil
}

// ── Step 12: Verify request binding ─────────────────────────────────
func stepCheckBinding(v *vctx) *core.OathMeshError {
	if v.claims.RQH != "" && v.cfg.RequestHash != "" {
		expectedHash := "sha256:" + sha256Hex(v.cfg.RequestHash)
		if v.claims.RQH != expectedHash {
			return core.NewOathMeshErrorAt(12,
				core.ErrBindingMismatch,
				fmt.Sprintf("request hash mismatch: token has %q but request hash is %q", v.claims.RQH, expectedHash),
				"ensure the request body has not been modified since the token was minted",
			)
		}
	}

	if v.cfg.RequireRequestBinding && v.claims.RQH == "" {
		return core.NewOathMeshErrorAt(12,
			core.ErrBindingRequired,
			"token missing rqh (request hash) claim",
			"mint a token with rqh= sha256:<canonical-request> for write/mutate operations",
		)
	}
	return nil
}

// ── Step 13: Check replay + revocation ──────────────────────────────
func stepCheckReplay(v *vctx) *core.OathMeshError {
	if v.cfg.ReplayCache != nil {
		remainingTTL := v.expTime.Sub(v.now)
		if remainingTTL < time.Second {
			remainingTTL = time.Second
		}

		replayed, err := v.cfg.ReplayCache.Check(v.ctx, v.claims.JTI, remainingTTL)
		if err != nil {
			return core.NewOathMeshErrorAt(13,
				core.ErrReplayDetected,
				fmt.Sprintf("replay cache error: %v", err),
				"check replay cache backend connectivity",
			)
		}
		if replayed {
			return core.NewOathMeshErrorAt(13,
				core.ErrReplayDetected,
				fmt.Sprintf("token %s has already been used", v.claims.JTI),
				"mint a new token — each Oath Token can only be used once",
			)
		}
	}

	if v.cfg.RevocationList != nil {
		revoked, err := v.cfg.RevocationList.IsRevoked(v.ctx, v.claims.Sub)
		if err != nil {
			// Fail open on fetch errors (network partition tolerance)
		} else if revoked {
			return core.NewOathMeshErrorAt(13,
				core.ErrSubjectRevoked,
				fmt.Sprintf("subject %s has been revoked", v.claims.Sub),
				"mint a token for a valid, active subject",
			)
		}
	}
	return nil
}

// ── Step 14: Evaluate policy ────────────────────────────────────────
func stepEvaluatePolicy(v *vctx) *core.OathMeshError {
	if v.cfg.PolicyEvaluator != nil {
		policyInput := &PolicyInput{
			Iss:    v.claims.Iss,
			Sub:    v.claims.Sub,
			Aud:    v.claims.Aud,
			Act:    v.claims.Act,
			Scope:  v.claims.Scope,
			Env:    v.claims.Env,
			Tenant: v.claims.Tenant,
		}
		if v.claims.Src != nil {
			policyInput.SrcType = v.claims.Src.Type
			policyInput.SrcRepo = v.claims.Src.Repo
			policyInput.SrcWflow = v.claims.Src.Workflow
		}

		decision := v.cfg.PolicyEvaluator.Evaluate(policyInput)
		if decision.Outcome == "deny" {
			return core.NewOathMeshErrorAt(14,
				core.ErrPolicyDenied,
				fmt.Sprintf("policy denied by rule %q", decision.RuleName),
				"check policy rules or request different permissions",
			)
		}
	}
	return nil
}
