package issuer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oathmesh/oathmesh/internal/sign"
)

type mintTestSigner struct{}

func (m mintTestSigner) GetIssuer() string                              { return "https://issuer.local" }
func (m mintTestSigner) JWKS() (*sign.JWKS, error)                      { return &sign.JWKS{}, nil }
func (m mintTestSigner) SignToken(req sign.MintRequest) (string, error) { return "token", nil }

func TestMintHandler_ResponseIncludesExpiryAndType(t *testing.T) {
	t.Setenv("OATHMESH_MINT_SECRET", "test-secret")
	srv := NewServer(mintTestSigner{})

	body := `{"sub":"svc://tests/api","aud":"https://inventory.internal","act":"inventory.read","ttl_hint":60}`
	req := httptest.NewRequest(http.MethodPost, "/v1/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	rec := httptest.NewRecorder()

	srv.router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp MintResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token != "token" {
		t.Fatalf("expected token in response, got %q", resp.Token)
	}
	if resp.ExpiresIn != 60 {
		t.Fatalf("expected expires_in=60, got %d", resp.ExpiresIn)
	}
	if resp.TokenType != "OathMesh" {
		t.Fatalf("expected token_type=OathMesh, got %q", resp.TokenType)
	}
}

func TestMintHandler_ExpiryFallbackClamp(t *testing.T) {
	t.Setenv("OATHMESH_MINT_SECRET", "test-secret")
	srv := NewServer(mintTestSigner{})

	body := `{"sub":"svc://tests/api","aud":"https://inventory.internal","act":"inventory.read","ttl_hint":999}`
	req := httptest.NewRequest(http.MethodPost, "/v1/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	rec := httptest.NewRecorder()

	srv.router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp MintResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ExpiresIn != sign.MaxTTL {
		t.Fatalf("expected expires_in=%d, got %d", sign.MaxTTL, resp.ExpiresIn)
	}
}
