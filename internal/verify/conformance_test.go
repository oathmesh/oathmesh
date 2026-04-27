package verify

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/sign"
)

func TestConformance_token_parsing_validation_failures(t *testing.T) {
	_, publicKey := generateTestKeys(t)
	cfg := testConfig(publicKey)

	_, err := Verify(context.Background(), "not-a-token", cfg)
	assertOathMeshError(t, err, core.ErrTokenMalformed)
}

func TestConformance_issuer_check_untrusted(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Iss = "https://evil.issuer.local"
	})

	cfg := testConfig(publicKey)
	_, err := Verify(context.Background(), token, cfg)
	assertOathMeshError(t, err, core.ErrIssuerUntrusted)
}

func TestConformance_audience_check_mismatch(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Aud = "https://wrong.audience.local"
	})

	cfg := testConfig(publicKey)
	_, err := Verify(context.Background(), token, cfg)
	assertOathMeshError(t, err, core.ErrAudienceMismatch)
}

func TestConformance_replay_detection_semantics(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil)
	cfg := testConfig(publicKey)

	if _, err := Verify(context.Background(), token, cfg); err != nil {
		t.Fatalf("expected first verification to pass, got: %v", err)
	}

	_, err := Verify(context.Background(), token, cfg)
	assertOathMeshError(t, err, core.ErrReplayDetected)
}

func TestConformance_revocation_subject_revoked(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil)
	cfg := testConfig(publicKey)
	cfg.RevocationList = &MemoryRevocationList{
		revocations: map[string]time.Time{
			testSubject: time.Now().Add(-1 * time.Minute),
		},
	}

	_, err := Verify(context.Background(), token, cfg)
	assertOathMeshError(t, err, core.ErrSubjectRevoked)
}

func TestConformance_alg_none_rejection(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
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

func TestConformance_subject_format_validation(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Sub = "invalid-subject"
	})
	cfg := testConfig(publicKey)

	_, err := Verify(context.Background(), token, cfg)
	if err == nil {
		t.Fatal("expected invalid subject error")
	}
	ome, ok := err.(*core.OathMeshError)
	if !ok {
		t.Fatalf("expected OathMeshError, got %T", err)
	}
	if ome.Code != core.ErrorCode(string(core.ErrClaimMissing)+":sub") {
		t.Fatalf("expected claim_missing:sub, got %s", ome.Code)
	}
}

func TestConformance_binding_required_semantics(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil) // no rqh by default
	cfg := testConfig(publicKey)
	cfg.RequireRequestBinding = true

	_, err := Verify(context.Background(), token, cfg)
	assertOathMeshError(t, err, core.ErrBindingRequired)
}

func TestConformance_iat_future_rejection(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Iat = time.Now().Add(60 * time.Second).Unix()
		c.Exp = time.Now().Add(120 * time.Second).Unix()
	})
	cfg := testConfig(publicKey)

	_, err := Verify(context.Background(), token, cfg)
	assertOathMeshError(t, err, core.ErrTokenExpired)
}
