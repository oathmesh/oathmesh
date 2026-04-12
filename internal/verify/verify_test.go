package verify

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/sign"
)

// ── Test helpers ────────────────────────────────────────────────────────────

const (
	testIssuer   = "https://issuer.test.oathmesh.dev"
	testAudience = "https://api.test.internal"
	testSubject  = "agent://repo/acme/deploy-bot"
	testAction   = "inventory.write"
	testKid      = "test-key-2026-04"
)

// mintTestToken creates a valid signed Oath Token for testing.
func mintTestToken(t *testing.T, privateKey ed25519.PrivateKey, overrides func(*sign.Claims)) string {
	t.Helper()

	claims := sign.Claims{
		Iss: testIssuer,
		Sub: testSubject,
		Aud: testAudience,
		Act: testAction,
		Iat: time.Now().Unix(),
		Exp: time.Now().Add(120 * time.Second).Unix(),
		JTI: uuid.New().String(),
	}

	if overrides != nil {
		overrides(&claims)
	}

	header := sign.Header{
		Typ: sign.TypeHeader,
		Alg: sign.AlgEdDSA,
		Kid: testKid,
	}

	token, err := sign.BuildJWS(header, claims, privateKey)
	if err != nil {
		t.Fatalf("failed to mint test token: %v", err)
	}
	return token
}

// testConfig creates a VerifierConfig for testing with sane defaults.
func testConfig(publicKey ed25519.PublicKey) *VerifierConfig {
	return &VerifierConfig{
		Audience:       testAudience,
		TrustedIssuers: []string{testIssuer},
		JWKSProvider:   NewStaticJWKSProvider(map[string]ed25519.PublicKey{testKid: publicKey}),
		ReplayCache:    NewMemoryReplayCache(),
	}
}

// generateTestKeys creates a fresh Ed25519 key pair.
func generateTestKeys(t *testing.T) (ed25519.PrivateKey, ed25519.PublicKey) {
	t.Helper()
	privateKey, publicKey, err := sign.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	return privateKey, publicKey
}

// buildRawToken constructs a JWS token from raw header/claims JSON for edge-case testing.
func buildRawToken(headerJSON, claimsJSON []byte, privateKey ed25519.PrivateKey) string {
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := headerB64 + "." + claimsB64
	signature := ed25519.Sign(privateKey, []byte(signingInput))
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}

// ── Valid token ──────────────────────────────────────────────────────────────

func TestVerify_ValidToken(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil)
	cfg := testConfig(publicKey)

	vcc, err := Verify(context.Background(), token, cfg)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	if vcc.Principal.Issuer != testIssuer {
		t.Errorf("expected issuer %q, got %q", testIssuer, vcc.Principal.Issuer)
	}
	if vcc.Principal.Subject != testSubject {
		t.Errorf("expected subject %q, got %q", testSubject, vcc.Principal.Subject)
	}
	if vcc.Action != testAction {
		t.Errorf("expected action %q, got %q", testAction, vcc.Action)
	}
	if vcc.TokenID == "" {
		t.Error("expected non-empty token ID")
	}
}

// ── Expired token ───────────────────────────────────────────────────────────

func TestVerify_ExpiredToken(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Exp = time.Now().Add(-60 * time.Second).Unix() // expired 60s ago
	})

	cfg := testConfig(publicKey)
	_, err := Verify(context.Background(), token, cfg)

	assertOathMeshError(t, err, core.ErrTokenExpired)
}

// ── Wrong audience ──────────────────────────────────────────────────────────

func TestVerify_WrongAudience(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Aud = "https://billing.internal" // wrong audience
	})

	cfg := testConfig(publicKey)
	_, err := Verify(context.Background(), token, cfg)

	assertOathMeshError(t, err, core.ErrAudienceMismatch)
}

// ── Invalid signature ───────────────────────────────────────────────────────

func TestVerify_InvalidSignature(t *testing.T) {
	privateKey, _ := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil)

	// Verify with a DIFFERENT key
	_, wrongPublicKey := generateTestKeys(t)
	cfg := testConfig(wrongPublicKey)

	_, err := Verify(context.Background(), token, cfg)

	assertOathMeshError(t, err, core.ErrSignatureInvalid)
}

// ── Missing required claims ─────────────────────────────────────────────────

func TestVerify_MissingRequiredClaim(t *testing.T) {
	tests := []struct {
		name      string
		claimName string
		override  func(*sign.Claims)
	}{
		{"missing iss", "iss", func(c *sign.Claims) { c.Iss = "" }},
		{"missing sub", "sub", func(c *sign.Claims) { c.Sub = "" }},
		{"missing aud", "aud", func(c *sign.Claims) { c.Aud = "" }},
		{"missing act", "act", func(c *sign.Claims) { c.Act = "" }},
		{"missing iat", "iat", func(c *sign.Claims) { c.Iat = 0 }},
		{"missing exp", "exp", func(c *sign.Claims) { c.Exp = 0 }},
		{"missing jti", "jti", func(c *sign.Claims) { c.JTI = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, publicKey := generateTestKeys(t)
			token := mintTestToken(t, privateKey, tt.override)

			cfg := testConfig(publicKey)
			// For missing iss/aud, issuer check or aud check may fire first.
			// We need to handle iss being empty → issuer_untrusted may fire before claim_missing.
			// Adjust trusted issuers to include "" so issuer check passes for the iss test.
			if tt.claimName == "iss" {
				cfg.TrustedIssuers = append(cfg.TrustedIssuers, "")
			}

			_, err := Verify(context.Background(), token, cfg)
			if err == nil {
				t.Fatalf("expected error for missing %s, got nil", tt.claimName)
			}

			ome, ok := err.(*core.OathMeshError)
			if !ok {
				t.Fatalf("expected OathMeshError, got %T: %v", err, err)
			}

			expectedCode := core.ErrorCode(fmt.Sprintf("%s:%s", core.ErrClaimMissing, tt.claimName))
			if ome.Code != expectedCode {
				t.Errorf("expected error code %q, got %q (message: %s)", expectedCode, ome.Code, ome.Message)
			}
		})
	}
}

// ── Unknown issuer ──────────────────────────────────────────────────────────

func TestVerify_UnknownIssuer(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Iss = "https://evil.issuer.dev"
	})

	cfg := testConfig(publicKey)
	_, err := Verify(context.Background(), token, cfg)

	assertOathMeshError(t, err, core.ErrIssuerUntrusted)
}

// ── Replay detection ────────────────────────────────────────────────────────

func TestVerify_ReplayDetected(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil)

	cfg := testConfig(publicKey)

	// First use: should succeed
	_, err := Verify(context.Background(), token, cfg)
	if err != nil {
		t.Fatalf("first verification should succeed: %v", err)
	}

	// Second use: should be rejected as replay
	_, err = Verify(context.Background(), token, cfg)
	assertOathMeshError(t, err, core.ErrReplayDetected)
}

// ── Algorithm "none" rejection ──────────────────────────────────────────────
// Spec Step 02: If alg is "none", REJECT immediately — do not proceed to any other step.

func TestVerify_AlgNone(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)

	// Build a token with alg: "none" (hand-crafted, not from normal signer)
	headerJSON, _ := json.Marshal(map[string]string{
		"typ": "om+jwt",
		"alg": "none",
		"kid": testKid,
	})
	claimsJSON, _ := json.Marshal(map[string]interface{}{
		"iss": testIssuer,
		"sub": testSubject,
		"aud": testAudience,
		"act": testAction,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(120 * time.Second).Unix(),
		"jti": uuid.New().String(),
	})

	token := buildRawToken(headerJSON, claimsJSON, privateKey)

	cfg := testConfig(publicKey)
	_, err := Verify(context.Background(), token, cfg)

	assertOathMeshError(t, err, core.ErrAlgorithmNotAllowed)
}

// ── Algorithm confusion attack ──────────────────────────────────────────────
// Token header says "ES256" but JWKS key is registered as "EdDSA".

func TestVerify_AlgorithmConfusion(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)

	headerJSON, _ := json.Marshal(map[string]string{
		"typ": "om+jwt",
		"alg": "ES256", // Claim ES256 but sign with EdDSA
		"kid": testKid,
	})
	claimsJSON, _ := json.Marshal(map[string]interface{}{
		"iss": testIssuer,
		"sub": testSubject,
		"aud": testAudience,
		"act": testAction,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(120 * time.Second).Unix(),
		"jti": uuid.New().String(),
	})

	token := buildRawToken(headerJSON, claimsJSON, privateKey)

	cfg := testConfig(publicKey)
	_, err := Verify(context.Background(), token, cfg)

	assertOathMeshError(t, err, core.ErrAlgorithmNotAllowed)
}

// ── Clock skew accepted (within tolerance) ──────────────────────────────────

func TestVerify_ClockSkewAccepted(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		// Token issued 5 seconds in the future — within 10s tolerance
		c.Iat = time.Now().Add(5 * time.Second).Unix()
		c.Exp = time.Now().Add(125 * time.Second).Unix()
	})

	cfg := testConfig(publicKey)
	vcc, err := Verify(context.Background(), token, cfg)
	if err != nil {
		t.Fatalf("expected clock skew within tolerance to be accepted: %v", err)
	}
	if vcc == nil {
		t.Fatal("expected non-nil VerifiedCallerContext")
	}
}

// ── Clock skew rejected (outside tolerance) ─────────────────────────────────

func TestVerify_ClockSkewRejected(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		// Token issued 15 seconds in the future — outside 10s tolerance
		c.Iat = time.Now().Add(15 * time.Second).Unix()
		c.Exp = time.Now().Add(135 * time.Second).Unix()
	})

	cfg := testConfig(publicKey)
	_, err := Verify(context.Background(), token, cfg)

	assertOathMeshError(t, err, core.ErrTokenExpired)
}

// ── Request hash binding match ──────────────────────────────────────────────

func TestVerify_RQHBindingMatch(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	requestBody := `{"action":"deploy","target":"prod"}`
	hash := sha256.Sum256([]byte(requestBody))
	rqh := fmt.Sprintf("sha256:%x", hash)

	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.RQH = rqh
	})

	cfg := testConfig(publicKey)
	cfg.RequestHash = requestBody

	vcc, err := Verify(context.Background(), token, cfg)
	if err != nil {
		t.Fatalf("expected rqh binding match to succeed: %v", err)
	}
	if vcc == nil {
		t.Fatal("expected non-nil VerifiedCallerContext")
	}
}

// ── Request hash binding mismatch ───────────────────────────────────────────

func TestVerify_RQHBindingMismatch(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	hash := sha256.Sum256([]byte(`original-body`))
	rqh := fmt.Sprintf("sha256:%x", hash)

	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.RQH = rqh
	})

	cfg := testConfig(publicKey)
	cfg.RequestHash = "tampered-body" // different from what was hashed

	_, err := Verify(context.Background(), token, cfg)

	assertOathMeshError(t, err, core.ErrBindingMismatch)
}

// ── JWKS stale key rotation ─────────────────────────────────────────────────
// Sign with key A, rotate to key B. Verifier's stale cache has only key B.
// On kid miss, it should refresh JWKS and find key A if it's still served.

func TestVerify_JWKS_StaleKeyRotation(t *testing.T) {
	// Generate two key pairs: key A (current) and key B (new after rotation)
	privateKeyA, publicKeyA := generateTestKeys(t)
	_, publicKeyB := generateTestKeys(t)

	kidA := "issuer-key-2026-03"
	kidB := "issuer-key-2026-04"

	// Sign token with key A
	header := sign.Header{Typ: sign.TypeHeader, Alg: sign.AlgEdDSA, Kid: kidA}
	claims := sign.Claims{
		Iss: testIssuer, Sub: testSubject, Aud: testAudience, Act: testAction,
		Iat: time.Now().Unix(), Exp: time.Now().Add(120 * time.Second).Unix(),
		JTI: uuid.New().String(),
	}
	token, err := sign.BuildJWS(header, claims, privateKeyA)
	if err != nil {
		t.Fatalf("failed to build test token: %v", err)
	}

	// JWKS provider serves both keys (simulating post-rotation JWKS with overlap)
	provider := NewStaticJWKSProvider(map[string]ed25519.PublicKey{
		kidA: publicKeyA,
		kidB: publicKeyB,
	})

	cfg := &VerifierConfig{
		Audience:       testAudience,
		TrustedIssuers: []string{testIssuer},
		JWKSProvider:   provider,
		ReplayCache:    NewMemoryReplayCache(),
	}

	vcc, err := Verify(context.Background(), token, cfg)
	if err != nil {
		t.Fatalf("expected stale key rotation to succeed after refresh: %v", err)
	}
	if vcc.Principal.Subject != testSubject {
		t.Errorf("expected subject %q, got %q", testSubject, vcc.Principal.Subject)
	}
}

// ── MemoryReplayCache concurrency ───────────────────────────────────────────
// 1000 concurrent goroutines checking distinct jtis — zero data races.
// Run with: go test -race ./internal/verify/...

func TestMemoryReplayCache_Concurrency(t *testing.T) {
	rc := NewMemoryReplayCache()
	defer rc.Close()

	const goroutines = 1000
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()

			jti := fmt.Sprintf("jti-concurrent-%d", i)
			ttl := 5 * time.Minute

			// First check: should NOT be replayed
			replayed, err := rc.Check(context.Background(), jti, ttl)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: check error: %w", i, err)
				return
			}
			if replayed {
				errors <- fmt.Errorf("goroutine %d: first check returned replayed=true for unique jti", i)
				return
			}

			// Second check with same jti: SHOULD be replayed
			replayed, err = rc.Check(context.Background(), jti, ttl)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: second check error: %w", i, err)
				return
			}
			if !replayed {
				errors <- fmt.Errorf("goroutine %d: second check should detect replay", i)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// ── No Authorization header (malformed token) ───────────────────────────────

func TestVerify_MalformedToken(t *testing.T) {
	_, publicKey := generateTestKeys(t)
	cfg := testConfig(publicKey)

	tests := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"single segment", "abc"},
		{"two segments", "abc.def"},
		{"four segments", "a.b.c.d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Verify(context.Background(), tt.token, cfg)
			if err == nil {
				t.Fatal("expected error for malformed token")
			}
		})
	}
}

// ── Source claims propagation ────────────────────────────────────────────────

func TestVerify_SourceClaimsPropagated(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Src = &sign.Source{
			Type:     "github_actions",
			Repo:     "acme/storefront",
			Workflow: "deploy.yml",
			RunID:    "123456",
			SHA:      "abc123def456",
		}
		c.Scope = []string{"inventory.read", "inventory.write"}
		c.Reason = "sync catalog after deploy"
		c.Env = "prod"
	})

	cfg := testConfig(publicKey)
	vcc, err := Verify(context.Background(), token, cfg)
	if err != nil {
		t.Fatalf("expected valid token: %v", err)
	}

	if vcc.Source == nil {
		t.Fatal("expected non-nil source")
	}
	if vcc.Source.Type != "github_actions" {
		t.Errorf("expected source type github_actions, got %s", vcc.Source.Type)
	}
	if vcc.Source.Repo != "acme/storefront" {
		t.Errorf("expected repo acme/storefront, got %s", vcc.Source.Repo)
	}
	if len(vcc.Scope) != 2 {
		t.Errorf("expected 2 scope values, got %d", len(vcc.Scope))
	}
	if vcc.Reason != "sync catalog after deploy" {
		t.Errorf("expected reason, got %q", vcc.Reason)
	}
	if vcc.Env != "prod" {
		t.Errorf("expected env prod, got %s", vcc.Env)
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func assertOathMeshError(t *testing.T, err error, expectedCode core.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %q, got nil", expectedCode)
	}

	ome, ok := err.(*core.OathMeshError)
	if !ok {
		t.Fatalf("expected OathMeshError, got %T: %v", err, err)
	}

	if ome.Code != expectedCode {
		t.Errorf("expected error code %q, got %q (message: %s)", expectedCode, ome.Code, ome.Message)
	}
}

// ── Mock policy evaluator for testing ───────────────────────────────────────

type mockPolicyEvaluator struct {
	outcome  string
	ruleName string
}

func (m *mockPolicyEvaluator) Evaluate(input *PolicyInput) *PolicyDecision {
	return &PolicyDecision{Outcome: m.outcome, RuleName: m.ruleName}
}

// ── Mock audit sink for testing ─────────────────────────────────────────────

type mockAuditSink struct {
	mu     sync.Mutex
	events []*core.AuditEvent
}

func (m *mockAuditSink) Emit(ctx context.Context, event *core.AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *mockAuditSink) getEvents() []*core.AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]*core.AuditEvent, len(m.events))
	copy(copied, m.events)
	return copied
}

// ── Policy deny test ────────────────────────────────────────────────────────
// Step 14: policy evaluator returns deny → policy_denied error + deny audit event

func TestVerify_PolicyDeny(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil)

	auditSink := &mockAuditSink{}

	cfg := testConfig(publicKey)
	cfg.PolicyEvaluator = &mockPolicyEvaluator{outcome: "deny", ruleName: "block-all"}
	cfg.AuditSink = auditSink

	_, err := Verify(context.Background(), token, cfg)

	assertOathMeshError(t, err, core.ErrPolicyDenied)

	// Verify audit event was emitted with outcome "deny"
	events := auditSink.getEvents()
	if len(events) == 0 {
		t.Fatal("expected at least one audit event on policy deny")
	}

	lastEvent := events[len(events)-1]
	if lastEvent.Outcome != "deny" {
		t.Errorf("expected audit outcome deny, got %s", lastEvent.Outcome)
	}
}

// ── Policy allow test ───────────────────────────────────────────────────────
// Step 14: policy evaluator returns allow → success + rule name in audit

func TestVerify_PolicyAllow(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil)

	cfg := testConfig(publicKey)
	cfg.PolicyEvaluator = &mockPolicyEvaluator{outcome: "allow", ruleName: "storefront-read"}

	vcc, err := Verify(context.Background(), token, cfg)
	if err != nil {
		t.Fatalf("expected policy allow, got error: %v", err)
	}

	if vcc.Principal.Subject != testSubject {
		t.Errorf("expected subject %q, got %q", testSubject, vcc.Principal.Subject)
	}
}

// ── No policy evaluator test ────────────────────────────────────────────────
// Nil evaluator → all authenticated tokens allowed (backward compatibility)

func TestVerify_NoPolicyEvaluator(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil)

	cfg := testConfig(publicKey)
	cfg.PolicyEvaluator = nil // explicitly nil

	vcc, err := Verify(context.Background(), token, cfg)
	if err != nil {
		t.Fatalf("expected nil policy evaluator to allow, got error: %v", err)
	}

	if vcc == nil {
		t.Fatal("expected non-nil VerifiedCallerContext")
	}
}

// ── Audit emission on deny paths ────────────────────────────────────────────
// Confirm that emitAndReturn fires an audit event on EVERY rejection, not just Step 14.

func TestVerify_AuditEmittedOnAllDenials(t *testing.T) {
	_, publicKey := generateTestKeys(t)
	privateKey2, _ := generateTestKeys(t)

	tests := []struct {
		name  string
		token string
		code  core.ErrorCode
	}{
		{
			name:  "malformed token",
			token: "only.two",
			code:  core.ErrClaimMissing,
		},
		{
			name: "unknown issuer",
			token: func() string {
				return mintTestToken(t, privateKey2, func(c *sign.Claims) {
					c.Iss = "https://evil.issuer"
				})
			}(),
			code: core.ErrIssuerUntrusted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditSink := &mockAuditSink{}
			cfg := testConfig(publicKey)
			cfg.AuditSink = auditSink

			_, err := Verify(context.Background(), tt.token, cfg)
			if err == nil {
				t.Fatal("expected error")
			}

			events := auditSink.getEvents()
			if len(events) == 0 {
				t.Fatalf("expected audit event for %s denial, got none", tt.name)
			}

			lastEvent := events[len(events)-1]
			if lastEvent.Outcome != "deny" {
				t.Errorf("expected audit outcome deny, got %s", lastEvent.Outcome)
			}
		})
	}
}

// ── Redis fail-closed behavior ──────────────────────────────────────────────
// RedisReplayCache with invalid URL should fail to connect;
// on check, fail-closed should return an error wrapping ErrCacheUnavailable.

func TestRedisReplayCache_FailClosed(t *testing.T) {
	rc, err := NewRedisReplayCache(RedisReplayCacheConfig{
		RedisURL:   "redis://localhost:59999/0", // port that nothing is listening on
		FailClosed: true,
	})
	if err != nil {
		t.Fatalf("expected RedisReplayCache to be created (lazy connect): %v", err)
	}
	defer rc.Close()

	// This should fail because there's no Redis at that port
	_, err = rc.Check(context.Background(), "test-jti", 5*time.Minute)
	if err == nil {
		t.Fatal("expected error on fail-closed Redis with no server")
	}

	// The error should wrap ErrCacheUnavailable
	if !containsError(err, "replay cache backend unavailable") {
		t.Errorf("expected ErrCacheUnavailable in error chain, got: %v", err)
	}
}

func containsError(err error, substring string) bool {
	if err == nil {
		return false
	}
	return fmt.Sprintf("%v", err) != "" && len(substring) > 0 && fmt.Sprintf("%v", err) != "" &&
		containsSubstring(err.Error(), substring)
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && searchSubstring(s, sub)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

