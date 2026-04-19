package verify

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/sign"
)

// ── Memory Revocation List Tests ────────────────────────────────────────────

func TestMemoryRevocationListIsRevoked_SubjectNotRevoked(t *testing.T) {
	rl := &MemoryRevocationList{
		revocations: make(map[string]time.Time),
	}

	revoked, err := rl.IsRevoked(context.Background(), "agent://test/subject-1")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if revoked {
		t.Error("expected subject to not be revoked")
	}
}

func TestMemoryRevocationListIsRevoked_SubjectRevoked(t *testing.T) {
	rl := &MemoryRevocationList{
		revocations: map[string]time.Time{
			"agent://test/subject-1": time.Now().Add(-5 * time.Minute),
		},
	}

	revoked, err := rl.IsRevoked(context.Background(), "agent://test/subject-1")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !revoked {
		t.Error("expected subject to be revoked")
	}
}

func TestMemoryRevocationListIsRevoked_DifferentSubjectNotAffected(t *testing.T) {
	rl := &MemoryRevocationList{
		revocations: map[string]time.Time{
			"agent://test/subject-1": time.Now().Add(-5 * time.Minute),
		},
	}

	// Different subject should not be affected
	revoked, err := rl.IsRevoked(context.Background(), "agent://test/subject-2")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if revoked {
		t.Error("expected different subject to not be revoked")
	}
}

func TestMemoryRevocationListIsRevoked_MultipleSubjects(t *testing.T) {
	rl := &MemoryRevocationList{
		revocations: map[string]time.Time{
			"agent://test/subject-1": time.Now().Add(-5 * time.Minute),
			"agent://test/subject-2": time.Now().Add(-10 * time.Minute),
			"agent://test/subject-3": time.Now().Add(-1 * time.Minute),
		},
	}

	tests := []struct {
		subject  string
		expected bool
	}{
		{"agent://test/subject-1", true},
		{"agent://test/subject-2", true},
		{"agent://test/subject-3", true},
		{"agent://test/subject-4", false},
		{"svc://api/service", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("subject_%s", tt.subject), func(t *testing.T) {
			revoked, err := rl.IsRevoked(context.Background(), tt.subject)
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
			if revoked != tt.expected {
				t.Errorf("expected revoked=%v, got %v", tt.expected, revoked)
			}
		})
	}
}

func TestMemoryRevocationList_FetchAndSync(t *testing.T) {
	// Start a mock issuer server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/revoked-subjects" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"revocations": [
					{"sub": "agent://test/subject-1", "revoked_at": "2024-01-01T00:00:00Z"},
					{"sub": "agent://test/subject-2", "revoked_at": "2024-01-02T00:00:00Z"}
				]
			}`)
		}
	}))
	defer server.Close()

	// Create revocation list with fast poll interval
	rl := NewMemoryRevocationList(server.URL, 100*time.Millisecond)
	defer rl.Close()

	// Wait for initial sync to complete
	time.Sleep(50 * time.Millisecond)

	// Check that revocations were fetched
	revoked1, err := rl.IsRevoked(context.Background(), "agent://test/subject-1")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !revoked1 {
		t.Error("expected subject-1 to be revoked")
	}

	revoked2, err := rl.IsRevoked(context.Background(), "agent://test/subject-2")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !revoked2 {
		t.Error("expected subject-2 to be revoked")
	}

	// Non-revoked subject should not be affected
	revoked3, err := rl.IsRevoked(context.Background(), "agent://test/subject-3")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if revoked3 {
		t.Error("expected subject-3 to not be revoked")
	}
}

func TestMemoryRevocationList_Close(t *testing.T) {
	// Start a mock issuer server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/revoked-subjects" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"revocations": []}`)
		}
	}))
	defer server.Close()

	rl := NewMemoryRevocationList(server.URL, 100*time.Millisecond)
	rl.Close()

	// After close, the polling should stop. Give it a moment to ensure it stops cleanly.
	time.Sleep(50 * time.Millisecond)
}

// ── Redis Revocation List Tests ─────────────────────────────────────────────

func TestRedisRevocationListIsRevoked_SubjectNotRevoked(t *testing.T) {
	t.Skip("Redis tests require a Redis server; skipping in unit test environment")
}

func TestRedisRevocationListIsRevoked_SubjectRevoked(t *testing.T) {
	t.Skip("Redis tests require a Redis server; skipping in unit test environment")
}

// ── Revocation Integration Tests ────────────────────────────────────────────
// These tests verify that revocation is properly integrated with the Verify() function.

func TestVerify_SubjectRevoked_MemoryRevocationList(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil)

	// Create a revocation list with the test subject revoked
	revocationList := &MemoryRevocationList{
		revocations: map[string]time.Time{
			testSubject: time.Now().Add(-5 * time.Minute),
		},
	}

	cfg := testConfig(publicKey)
	cfg.RevocationList = revocationList

	vcc, err := Verify(context.Background(), token, cfg)
	if err == nil {
		t.Fatalf("expected verification to fail for revoked subject, got success: %v", vcc)
	}

	if ome, ok := err.(*core.OathMeshError); ok {
		if ome.Code != core.ErrSubjectRevoked {
			t.Errorf("expected error code %s, got %s", core.ErrSubjectRevoked, ome.Code)
		}
	} else {
		t.Errorf("expected OathMeshError, got %T: %v", err, err)
	}
}

func TestVerify_SubjectNotRevoked_MemoryRevocationList(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	token := mintTestToken(t, privateKey, nil)

	// Create an empty revocation list
	revocationList := &MemoryRevocationList{
		revocations: make(map[string]time.Time),
	}

	cfg := testConfig(publicKey)
	cfg.RevocationList = revocationList

	vcc, err := Verify(context.Background(), token, cfg)
	if err != nil {
		t.Fatalf("expected verification to succeed, got error: %v", err)
	}

	if vcc.Principal.Subject != testSubject {
		t.Errorf("expected subject %q, got %q", testSubject, vcc.Principal.Subject)
	}
}

func TestVerify_DifferentSubjectNotAffected(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)

	// Mint token with a different subject
	token := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Sub = "agent://test/other-subject"
	})

	// Create a revocation list with testSubject revoked
	revocationList := &MemoryRevocationList{
		revocations: map[string]time.Time{
			testSubject: time.Now().Add(-5 * time.Minute),
		},
	}

	cfg := testConfig(publicKey)
	cfg.RevocationList = revocationList

	vcc, err := Verify(context.Background(), token, cfg)
	if err != nil {
		t.Fatalf("expected verification to succeed, got error: %v", err)
	}

	if vcc.Principal.Subject != "agent://test/other-subject" {
		t.Errorf("expected subject %q, got %q", "agent://test/other-subject", vcc.Principal.Subject)
	}
}

// ── Revocation Policy Tests ─────────────────────────────────────────────────
// These tests verify the revocation policy: ALL tokens for a revoked subject are invalid.

func TestRevocationPolicy_AllTokensInvalidAfterRevocation(t *testing.T) {
	// This test verifies that once a subject is revoked, ALL tokens (new or old)
	// are considered invalid, regardless of iat.
	privateKey, publicKey := generateTestKeys(t)

	now := time.Now()

	// Mint token issued 1 hour ago
	tokenIssuedLongAgo := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Iat = now.Add(-1 * time.Hour).Unix()
		c.Exp = now.Add(1 * time.Hour).Unix() // Still valid
	})

	// Mint token issued recently
	tokenIssuedRecently := mintTestToken(t, privateKey, func(c *sign.Claims) {
		c.Iat = now.Unix()
		c.Exp = now.Add(2 * time.Hour).Unix()
	})

	// Create a revocation list where the subject was revoked 30 minutes ago
	revokedAt := now.Add(-30 * time.Minute)
	revocationList := &MemoryRevocationList{
		revocations: map[string]time.Time{
			testSubject: revokedAt,
		},
	}

	cfg := testConfig(publicKey)
	cfg.RevocationList = revocationList

	// Both tokens should be rejected, even the one issued long before revocation
	_, err1 := Verify(context.Background(), tokenIssuedLongAgo, cfg)
	if err1 == nil {
		t.Error("expected old token to be rejected after revocation")
	}
	if ome, ok := err1.(*core.OathMeshError); ok {
		if ome.Code != core.ErrSubjectRevoked {
			t.Errorf("expected ErrSubjectRevoked, got %s", ome.Code)
		}
	}

	// Recently issued token should also be rejected
	_, err2 := Verify(context.Background(), tokenIssuedRecently, cfg)
	if err2 == nil {
		t.Error("expected recent token to be rejected after revocation")
	}
	if ome, ok := err2.(*core.OathMeshError); ok {
		if ome.Code != core.ErrSubjectRevoked {
			t.Errorf("expected ErrSubjectRevoked, got %s", ome.Code)
		}
	}
}
