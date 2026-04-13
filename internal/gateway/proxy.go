package gateway

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/verify"
)

// Config defines the configuration for the OathMesh Gateway reverse proxy.
type Config struct {
	UpstreamURL string
	VerifyConfig *verify.VerifierConfig
}

// NewProxy creates a reverse proxy that enforces OathMesh verification
// before forwarding requests to the upstream URL.
func NewProxy(cfg Config) (http.Handler, error) {
	target, err := url.Parse(cfg.UpstreamURL)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Keep original behavior of SingleHostReverseProxy but allow us
	// to inject our verification middleware before proxy.ServeHTTP.
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Extract Token
		authz := r.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(authz, "OathMesh ") {
			sendError(w, r, core.NewOathMeshError(
				core.ErrorCode("claim_missing:token"),
				"missing or invalid Authorization header",
				"provide a token in the format 'Authorization: OathMesh <token>'",
			))
			return
		}

		token := strings.TrimPrefix(authz, "OathMesh ")

		// Step 2-14: Verification & Policy Evaluation
		vcc, err := verify.Verify(r.Context(), token, cfg.VerifyConfig)
		if err != nil {
			if oe, ok := err.(*core.OathMeshError); ok {
				sendError(w, r, oe)
				return
			}
			sendError(w, r, core.NewOathMeshError(
				core.ErrorCode("verification_failed"),
				fmt.Sprintf("internal verification error: %v", err),
				"check gateway configuration and logs",
			))
			return
		}

		// Step 15: Inject headers and forward
		InjectHeaders(r, vcc)
		
		// Adjust request URL and Host to target the upstream
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.Host = target.Host

		proxy.ServeHTTP(w, r)
	}), nil
}

func sendError(w http.ResponseWriter, r *http.Request, oe *core.OathMeshError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	
	// Ensure a Request ID is always present for traceability
	if oe.ReqID == "" {
		oe.ReqID = getRequestID(r)
	}

	if err := json.NewEncoder(w).Encode(oe); err != nil {
		slog.Error("failed to encode error response", "error", err)
	}
}

// getRequestID tries to find a correlation ID from headers.
func getRequestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-Id"); id != "" {
		return id
	}
	return "unknown"
}
