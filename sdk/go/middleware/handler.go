package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/verify"
)

type callerContextKey struct{}

// OathMeshMiddleware creates an HTTP middleware that verifies incoming
// OathMesh tokens against the provided VerifierConfig.
// It injects the VerifiedCallerContext into the request Context.
func OathMeshMiddleware(cfg *verify.VerifierConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			token := extractToken(authz)
			if token == "" {
				sendError(w, r, core.NewOathMeshError(
					core.ErrorCode("claim_missing:token"),
					"missing or invalid Authorization header",
					"provide a token in the format 'Authorization: OathMesh <token>'",
				))
				return
			}

			vcc, err := verify.Verify(r.Context(), token, cfg)
			if err != nil {
				if oe, ok := err.(*core.OathMeshError); ok {
					sendError(w, r, oe)
					return
				}
				sendError(w, r, core.NewOathMeshError(
					core.ErrorCode("verification_failed"),
					"internal verification error",
					"check server logs",
				))
				return
			}

			// Add VerifiedCallerContext to the request context
			ctx := context.WithValue(r.Context(), callerContextKey{}, vcc)
			
			// Optional: strip Authorization from the request reaching the handler
			// so the upstream app doesn't accidentally log or forward the raw token.
			r.Header.Del("Authorization")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CallerFrom extracts the VerifiedCallerContext from the request context.
// Returns nil if the middleware didn't run or verification failed.
func CallerFrom(ctx context.Context) *core.VerifiedCallerContext {
	vcc, ok := ctx.Value(callerContextKey{}).(*core.VerifiedCallerContext)
	if !ok {
		return nil
	}
	return vcc
}

func sendError(w http.ResponseWriter, r *http.Request, oe *core.OathMeshError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	
	id := r.Header.Get("X-Request-Id")
	if id == "" {
		id = "unknown"
	}
	oe.ReqID = id

	_ = json.NewEncoder(w).Encode(oe)
}

// extractToken pulls the raw token from an Authorization header.
// Supports "OathMesh <token>" (canonical).
func extractToken(header string) string {
	if strings.HasPrefix(header, "OathMesh ") {
		return strings.TrimPrefix(header, "OathMesh ")
	}
	return ""
}

