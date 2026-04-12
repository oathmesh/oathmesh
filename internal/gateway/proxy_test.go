package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
)

func TestInjectHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	req.Header.Set("Authorization", "OathMesh token123")
	
	vcc := &core.VerifiedCallerContext{
		Principal: core.Principal{
			Subject: "agent://test/bot",
			Issuer:  "https://issuer.local",
		},
		Action:  "read",
		TokenID: "jti-123",
		ExpiresAt: time.Now().Add(time.Minute),
		Env:     "staging",
	}

	InjectHeaders(req, vcc)

	if req.Header.Get("Authorization") != "" {
		t.Errorf("expected Authorization header to be stripped, got %q", req.Header.Get("Authorization"))
	}

	if req.Header.Get(HeaderSubject) != "agent://test/bot" {
		t.Errorf("wrong subject header: %q", req.Header.Get(HeaderSubject))
	}
	if req.Header.Get(HeaderAction) != "read" {
		t.Errorf("wrong action header: %q", req.Header.Get(HeaderAction))
	}
	if req.Header.Get(HeaderTokenID) != "jti-123" {
		t.Errorf("wrong token id header: %q", req.Header.Get(HeaderTokenID))
	}
	if req.Header.Get(HeaderIssuer) != "https://issuer.local" {
		t.Errorf("wrong issuer header: %q", req.Header.Get(HeaderIssuer))
	}
	if req.Header.Get(HeaderEnv) != "staging" {
		t.Errorf("wrong env header: %q", req.Header.Get(HeaderEnv))
	}
}

func TestInjectHeaders_NoEnv(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	req.Header.Set("Authorization", "OathMesh token123")
	// Pre-existing env header that should be cleared
	req.Header.Set(HeaderEnv, "prod")
	
	vcc := &core.VerifiedCallerContext{
		Principal: core.Principal{
			Subject: "job://test",
			Issuer:  "https://issuer.local",
		},
		Action:  "deploy",
		TokenID: "jti-456",
		// Env left empty
	}

	InjectHeaders(req, vcc)

	if req.Header.Get(HeaderEnv) != "" {
		t.Errorf("expected Env header to be stripped since it's empty in context, got %q", req.Header.Get(HeaderEnv))
	}
}

func TestProxyConfig(t *testing.T) {
	_, err := NewProxy(Config{
		UpstreamURL: "http://127.0.0.1:8080",
	})
	if err != nil {
		t.Fatalf("expected no error configuring proxy, got: %v", err)
	}

	_, err = NewProxy(Config{
		UpstreamURL: "://invalid-url",
	})
	if err == nil {
		t.Fatal("expected error on invalid URL, got nil")
	}
}
