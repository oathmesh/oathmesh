package main

import (
	"context"
	"strings"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/oathmesh/oathmesh/internal/config"
	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/verify"
)

// Config represents the Kong plugin configuration.
// It will be populated by Kong.
type Config struct {
	Audience string `json:"audience"`
	Issuers  string `json:"issuers"`
}

// OathMeshPlugin represents our Kong plugin instance.
type OathMeshPlugin struct {
	VerifierConfig *verify.VerifierConfig
}

// New constructs a new OathMeshPlugin instance.
func New() interface{} {
	return &OathMeshPlugin{}
}

// Access is executed by Kong for every incoming request.
func (conf *Config) Access(kong *pdk.PDK) {
	// 1. Build VerifierConfig (in a real scenario, we'd cache this or build on init)
	// For simplicity in the plugin execution, we build it per request using the plugin config.
	issuers := strings.Split(conf.Issuers, ",")
	for i, v := range issuers {
		issuers[i] = strings.TrimSpace(v)
	}

	verifierCfg := &verify.VerifierConfig{
		Audience:       conf.Audience,
		TrustedIssuers: issuers,
		JWKSProvider:   verify.NewJWKSCache(verify.DefaultJWKSCacheTTL, nil),
		ReplayCache:    verify.NewMemoryReplayCache(),
	}

	// 2. Extract Authorization header
	authHeader, err := kong.Request.GetHeader("authorization")
	if err != nil || authHeader == "" {
		kong.Response.Exit(401, []byte("missing authorization header"), map[string][]string{})
		return
	}

	// 3. Parse Token
	token := authHeader
	switch {
	case strings.HasPrefix(token, "Bearer "):
		token = strings.TrimPrefix(token, "Bearer ")
	case strings.HasPrefix(token, "OathMesh "):
		token = strings.TrimPrefix(token, "OathMesh ")
	default:
		kong.Response.Exit(401, []byte("invalid authorization header format"), map[string][]string{})
		return
	}

	// 4. Verify Token
	vcc, err := verify.Verify(context.Background(), token, verifierCfg)
	if err != nil {
		statusCode := 401
		if oe, ok := err.(*core.OathMeshError); ok {
			if oe.Code == core.ErrPolicyDenied || oe.Code == core.ErrSubjectRevoked {
				statusCode = 403
			}
		}
		kong.Response.Exit(statusCode, []byte(err.Error()), map[string][]string{})
		return
	}

	// 5. Inject verified context to Upstream
	_ = kong.ServiceRequest.SetHeader("x-oathmesh-subject", vcc.Principal.Subject)
	_ = kong.ServiceRequest.SetHeader("x-oathmesh-action", vcc.Action)
	_ = kong.ServiceRequest.SetHeader("x-oathmesh-issuer", vcc.Principal.Issuer)

	// Allow request to proceed to upstream
}

var (
	Version  = "0.3.0"
	Priority = 1000
)

func main() {
	// Initialize default configuration
	_ = config.LoadFromEnv()

	_ = server.StartServer(New, Version, Priority)
}
