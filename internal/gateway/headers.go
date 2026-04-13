package gateway

import (
	"net/http"

	"github.com/MustafaMahmoudAtta111/oathmesh/internal/core"
)

const (
	HeaderSubject = "X-OathMesh-Subject"
	HeaderAction  = "X-OathMesh-Action"
	HeaderTokenID = "X-OathMesh-Token-Id"
	HeaderIssuer  = "X-OathMesh-Issuer"
	HeaderEnv     = "X-OathMesh-Env"
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
}
