package issuer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/MustafaMahmoudAtta111/oathmesh/internal/sign"
)

type mockKeySet struct{}

func (m mockKeySet) GetIssuer() string { return "https://issuer.local" }
func (m mockKeySet) JWKS() (*sign.JWKS, error) { return nil, nil }
func (m mockKeySet) SignToken(req sign.MintRequest) (string, error) { return "token", nil }

func TestServer_HealthzBypass(t *testing.T) {
	// 1. Create a proxy mock that always returns 401 (acting as a strictly enforcing gateway)
	gatewayHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	srv := NewServer(mockKeySet{})
	srv.SetGateway(gatewayHandler)

	r := srv.router()

	// 2. Test that GET /healthz bypasses the gateway entirely and returns 200 without Auth
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected GET /healthz to bypass auth and return 200, got %d", rec.Code)
	}

	// 3. Test that other random URLs hit the gateway handler (401)
	reqMiss := httptest.NewRequest(http.MethodGet, "/random-api/resource", nil)
	recMiss := httptest.NewRecorder()

	r.ServeHTTP(recMiss, reqMiss)

	if recMiss.Code != http.StatusUnauthorized {
		t.Errorf("expected random paths to hit gateway Auth (401), got %d", recMiss.Code)
	}
}
