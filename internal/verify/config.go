package verify

import (
	"context"
	"crypto/ed25519"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
)

// AllowedAlgorithms is the set of JWS algorithms accepted by OathMesh.
// EdDSA is the primary algorithm; ES256 is accepted for receivers that
// cannot verify Ed25519. "none", "HS256", and weak RSA keys are always rejected.
var AllowedAlgorithms = map[string]bool{
	"EdDSA": true,
	"ES256": true,
}

// VerifierConfig holds all configuration for the Verify() function.
// OathMesh authenticates the caller. The receiver authorizes the request.
type VerifierConfig struct {
	// Audience is the receiver's own canonical URL.
	// Tokens with a different aud claim are rejected.
	Audience string

	// TrustedIssuers is the explicit allowlist of issuer URLs.
	// No wildcards, no auto-discovery. Unknown issuers are always rejected.
	TrustedIssuers []string

	// JWKSProvider fetches public keys for a given issuer URL and key ID.
	// If nil, a default HTTP-based JWKS provider is created.
	JWKSProvider JWKSProvider

	// ReplayCache checks and records jti values to prevent replay attacks.
	// If nil, replay checking is skipped (not recommended in production).
	ReplayCache core.ReplayCache

	// AuditSink receives audit events for every verification attempt.
	// If nil, no audit events are emitted.
	AuditSink core.AuditSink

	// RevocationList checks if a subject has been explicitly revoked.
	// If nil, revocation checking is skipped.
	RevocationList RevocationList

	// PolicyEvaluator evaluates policy rules against verified token claims.
	// If nil, all authenticated tokens are allowed (no policy enforcement).
	// Wire this to a policy.PolicyEngine for production use.
	PolicyEvaluator PolicyEvaluator

	// ClockSkewLeeway is the maximum allowed clock difference between issuer and receiver.
	// Default: 30 seconds.
	ClockSkewLeeway time.Duration

	// RequestHash is the canonical request string to verify against the rqh claim.
	// If empty, rqh binding check is skipped even if the token has an rqh claim.
	RequestHash string

	// RequireRequestBinding enforces that tokens MUST include an rqh claim.
	// When true, tokens without rqh are rejected with error "binding_required".
	// Recommended for all write/mutate endpoints to prevent tampering.
	// Default: false (for backward compatibility).
	RequireRequestBinding bool

	// Now is a function that returns the current time.
	// Defaults to time.Now if nil. Exposed for testing.
	Now func() time.Time
}

// JWKSProvider abstracts JWKS key retrieval.
// Implementations include the HTTP-fetching JWKSCache and static test providers.
//
// SECURITY: The issuerKey is a trusted lookup key (often an issuer URL string that
// has already passed trusted-issuer allowlist checks), not arbitrary user input.
// The provider resolves this key to a full URL from server-controlled configuration.
// This design prevents SSRF by ensuring user input is never used in URL construction.
type JWKSProvider interface {
	// GetKey returns the public key for the given issuer key and key ID.
	// The issuerKey is a lookup key that maps to a full issuer URL via config.
	// If the key is not in cache, the provider should fetch fresh JWKS
	// from the resolved issuer's /.well-known/jwks.json endpoint.
	GetKey(issuerKey string, kid string) (ed25519.PublicKey, string, error)
}

// RevocationList abstracts subject revocation checks before policy evaluation.
type RevocationList interface {
	// IsRevoked returns true if the subject has been revoked.
	// OathMesh revocation policy: revocation invalidates ALL tokens for a subject,
	// regardless of when they were issued.
	IsRevoked(ctx context.Context, subject string) (bool, error)
}

// PolicyEvaluator evaluates OathMesh policy rules against token claims.
// Implementations should be thread-safe for concurrent use.
type PolicyEvaluator interface {
	Evaluate(input *PolicyInput) *PolicyDecision
}

// PolicyInput holds the claims needed for policy evaluation.
type PolicyInput struct {
	Iss      string
	Sub      string
	Aud      string
	Act      string
	Scope    []string
	Env      string
	Tenant   string
	SrcType  string
	SrcRepo  string
	SrcWflow string
}

// PolicyDecision represents the result of a policy evaluation.
type PolicyDecision struct {
	// Outcome is "allow" or "deny".
	Outcome string
	// RuleName is the name of the matched rule.
	RuleName string
}
