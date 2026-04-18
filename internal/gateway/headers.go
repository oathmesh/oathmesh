package gateway

import (
	"net/http"

	"github.com/oathmesh/oathmesh/internal/core"
)

const (
	HeaderSubject   = "X-OathMesh-Subject"
	HeaderAction    = "X-OathMesh-Action"
	HeaderAlgorithm = "X-OathMesh-Algorithm"
	HeaderTokenID   = "X-OathMesh-Token-Id" //nolint:gosec // Not a credential, just a custom HTTP header name
	HeaderIssuer  = "X-OathMesh-Issuer"
	HeaderEnv     = "X-OathMesh-Env"
	HeaderTenant  = "X-OathMesh-Tenant"
)

// InjectHeaders removes the Authorization header so the raw Oath Token is
// never forwarded upstream. It then injects the context headers with the
// verified caller information.
func InjectHeaders(req *http.Request, vcc *core.VerifiedCallerContext) {
	// Strip the raw token unconditionally
	req.Header.Del("Authorization")

	// Inject context headers
	req.Header.Set(HeaderSubject, vcc.Principal.Subject)
	req.Header.Set(HeaderAction, vcc.Action)
	req.Header.Set(HeaderTokenID, vcc.TokenID)
	req.Header.Set(HeaderIssuer, vcc.Principal.Issuer)

	if vcc.Env != "" {
		req.Header.Set(HeaderEnv, vcc.Env)
	} else {
		req.Header.Del(HeaderEnv)
	}

	if vcc.Tenant != "" {
		req.Header.Set(HeaderTenant, vcc.Tenant)
	} else {
		req.Header.Del(HeaderTenant)
	}
}
