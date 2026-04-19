package verify

import (
	"context"
	"testing"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/sign"
)

func TestConformance_token_parsing_validation_failures(t *testing.T) {
	_, publicKey := generateTestKeys(t)
	cfg := testConfig(publicKey)

	_, err := Verify(context.Background(), "not-a-token", cfg)
	assertOathMeshError(t, err, core.ErrClaimMissing)
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
