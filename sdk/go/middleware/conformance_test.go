package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oathmesh/oathmesh/internal/verify"
)

func TestConformance_middleware_auth_header_handling_semantics(t *testing.T) {
	cfg := &verify.VerifierConfig{}
	mw := OathMeshMiddleware(cfg)
	handler := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler should not be called")
	}))

	tests := []struct {
		name          string
		authorization string
	}{
		{name: "missing", authorization: ""},
		{name: "wrong_scheme", authorization: "Bearer abc123"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authorization != "" {
				req.Header.Set("Authorization", tc.authorization)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), "claim_missing:token") {
				t.Fatalf("expected claim_missing:token, got body=%s", rec.Body.String())
			}
		})
	}
}
