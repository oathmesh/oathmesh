package issuer

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

// MintAuth is middleware that protects mint/exchange endpoints with a
// pre-shared key. The caller must include the header:
//
//	Authorization: Bearer <OATHMESH_MINT_SECRET>
//
// If OATHMESH_MINT_SECRET is empty, the middleware rejects ALL requests
// with 503 Service Unavailable — fail-closed by design.
//
// Public endpoints (JWKS, discovery, healthz, metrics) are NOT protected
// by this middleware; they are mounted on a separate route group.
func MintAuth(next http.Handler) http.Handler {
	secret := os.Getenv("OATHMESH_MINT_SECRET")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" {
			// Fail-closed: if no secret is configured, reject everything.
			http.Error(w, `{"error":"mint_auth_unconfigured","message":"OATHMESH_MINT_SECRET is not set — mint endpoint is disabled","fix":"set the OATHMESH_MINT_SECRET environment variable"}`, http.StatusServiceUnavailable)
			return
		}

		authz := r.Header.Get("Authorization")
		if authz == "" {
			http.Error(w, `{"error":"mint_auth_required","message":"Authorization header is required for mint endpoints","fix":"include 'Authorization: Bearer <OATHMESH_MINT_SECRET>' header"}`, http.StatusUnauthorized)
			return
		}

		// Accept "Bearer <secret>" format
		token := strings.TrimPrefix(authz, "Bearer ")
		if token == authz {
			// No "Bearer " prefix found
			http.Error(w, `{"error":"mint_auth_invalid","message":"Authorization header must use Bearer scheme","fix":"use 'Authorization: Bearer <secret>' format"}`, http.StatusUnauthorized)
			return
		}

		// Constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(token), []byte(secret)) != 1 {
			http.Error(w, `{"error":"mint_auth_denied","message":"invalid mint secret","fix":"check your OATHMESH_MINT_SECRET value"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
