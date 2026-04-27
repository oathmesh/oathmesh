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
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/metrics"
	"github.com/oathmesh/oathmesh/internal/sign"
)

// Verify performs all 14 OathMesh verification steps and returns a
// VerifiedCallerContext on success, or an OathMeshError on failure.
//
// Steps are executed as a func slice pipeline (see steps.go).
// Each step operates on a shared vctx and returns nil on success or
// an OathMeshError on failure. The loop stops on the first error.
func Verify(ctx context.Context, token string, cfg *VerifierConfig) (*core.VerifiedCallerContext, error) {
	metrics.VerificationsTotal.Inc()

	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	clockSkewLeeway := cfg.ClockSkewLeeway
	if clockSkewLeeway == 0 {
		clockSkewLeeway = 30 * time.Second
	}

	v := &vctx{
		ctx:             ctx,
		token:           token,
		cfg:             cfg,
		nowFn:           nowFn,
		clockSkewLeeway: clockSkewLeeway,
	}

	// Execute the pipeline — stop on first error.
	for _, s := range pipeline {
		if err := s.fn(v); err != nil {
			err.Step = s.step
			return nil, emitAndReturn(ctx, cfg, &v.claims, err)
		}
	}

	// Build VerifiedCallerContext from validated claims.
	var src *core.Source
	if v.claims.Src != nil {
		src = &core.Source{
			Type:     v.claims.Src.Type,
			Repo:     v.claims.Src.Repo,
			Workflow: v.claims.Src.Workflow,
			RunID:    v.claims.Src.RunID,
			SHA:      v.claims.Src.SHA,
		}
	}

	vcc := &core.VerifiedCallerContext{
		Principal: core.Principal{
			Issuer:  v.claims.Iss,
			Subject: v.claims.Sub,
		},
		Action:    v.claims.Act,
		Scope:     v.claims.Scope,
		Reason:    v.claims.Reason,
		Source:    src,
		TokenID:   v.claims.JTI,
		IssuedAt:  v.iatTime,
		ExpiresAt: v.expTime,
		Env:       v.claims.Env,
		Tenant:    v.claims.Tenant,
	}

	// Emit audit event — allow
	emitAudit(ctx, cfg, &v.claims, "allow", "verification_passed")

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
	metrics.VerificationErrors.Inc()
	switch err.Code {
	case core.ErrPolicyDenied:
		metrics.PolicyDenials.Inc()
	case core.ErrReplayDetected:
		metrics.ReplaysDetected.Inc()
	case core.ErrSignatureInvalid,
		core.ErrIssuerUntrusted,
		core.ErrTokenExpired,
		core.ErrAudienceMismatch,
		core.ErrAlgorithmNotAllowed,
		core.ErrClaimMissing,
		core.ErrBindingMismatch,
		core.ErrBindingRequired,
		core.ErrSubjectRevoked,
		core.ErrTokenMalformed,
		core.ErrVerificationFailed:
		// counted by VerificationErrors.Add(1) above
	}

	emitAudit(ctx, cfg, claims, "deny", string(err.Code)+": "+err.Message)
	return err
}
