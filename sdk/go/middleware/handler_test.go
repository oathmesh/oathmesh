package middleware

import (
	"context"
	"crypto/ed25519"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/sign"
	"github.com/oathmesh/oathmesh/internal/verify"
)

// A stub keyset so we can mint real tokens for the mock
type staticKeyset struct {
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

func (s staticKeyset) GetIssuer() string { return "https://issuer.local" }
func (s staticKeyset) GetKid() string    { return "issuer-key-2024-01" }
func (s staticKeyset) GetAllPublicKeys() map[string]ed25519.PublicKey {
	return map[string]ed25519.PublicKey{"issuer-key-2024-01": s.pub}
}
func (s staticKeyset) SignToken(req sign.MintRequest) (string, error) {
	return sign.SignToken(req, "https://issuer.local", s.priv, "issuer-key-2024-01")
}

func TestOathMeshMiddleware_MissingToken(t *testing.T) {
	cfg := &verify.VerifierConfig{}
	mw := OathMeshMiddleware(cfg)
	
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "claim_missing:token") {
		t.Errorf("expected claim_missing:token error")
	}
}

func TestCallerFrom_NilGraceful(t *testing.T) {
	// Must return nil safely if no context was set
	vcc := CallerFrom(context.Background())
	if vcc != nil {
		t.Fatalf("expected nil for empty context, got %+v", vcc)
	}

	// Test populated context
	inVCC := &core.VerifiedCallerContext{
		TokenID: "123",
	}
	ctx := context.WithValue(context.Background(), callerContextKey{}, inVCC)
	outVCC := CallerFrom(ctx)
	if outVCC == nil || outVCC.TokenID != "123" {
		t.Fatalf("failed to extract populated context")
	}
}

func TestOathMeshMiddleware_SuccessAndRace(t *testing.T) {
	priv, pub, err := sign.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	ks := staticKeyset{priv: priv, pub: pub}

	cfg := &verify.VerifierConfig{
		Audience:       "https://api.internal",
		TrustedIssuers: []string{"https://issuer.local"},
		JWKSProvider:   verify.NewStaticJWKSProvider(ks.GetAllPublicKeys()),
		ReplayCache:    verify.NewMemoryReplayCache(),
	}

	mw := OathMeshMiddleware(cfg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		caller := CallerFrom(r.Context())
		if caller == nil {
			t.Errorf("expected caller context to be set")
		} else if caller.Principal.Subject != "user://local" {
			t.Errorf("expected subject user://local, got %s", caller.Principal.Subject)
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Run multiple parallel requests to effectively trigger race detector
	for i := 0; i < 10; i++ {
		t.Run("concurrency", func(t *testing.T) {
			t.Parallel()
			
			// We must mint a new token for each request because MemoryReplayCache 
			// will explicitly block reused exact JTIs.
			newToken, _ := ks.SignToken(sign.MintRequest{
				Sub: "user://local",
				Aud: "https://api.internal",
				Act: "read",
				TTL: 60,
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "OathMesh "+newToken)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
			}
		})
	}
}
