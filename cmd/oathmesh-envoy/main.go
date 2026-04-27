package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/oathmesh/oathmesh/internal/config"
	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/verify"
)

type envoyExtAuthzServer struct {
	verifierCfg *verify.VerifierConfig
}

// Check implements the Envoy ext_authz Authorization gRPC service.
func (s *envoyExtAuthzServer) Check(ctx context.Context, req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	// Extract the Authorization header
	var authHeader string
	if req.Attributes != nil && req.Attributes.Request != nil && req.Attributes.Request.Http != nil {
		headers := req.Attributes.Request.Http.Headers
		if val, ok := headers["authorization"]; ok {
			authHeader = val
		}
	}

	if authHeader == "" {
		return s.deny(codes.Unauthenticated, "missing authorization header"), nil
	}

	// Support both "Bearer" and "OathMesh" prefixes
	token := authHeader
	switch {
	case strings.HasPrefix(token, "Bearer "):
		token = strings.TrimPrefix(token, "Bearer ")
	case strings.HasPrefix(token, "OathMesh "):
		token = strings.TrimPrefix(token, "OathMesh ")
	default:
		return s.deny(codes.Unauthenticated, "invalid authorization header format"), nil
	}

	// Verify the token
	vcc, err := verify.Verify(ctx, token, s.verifierCfg)
	if err != nil {
		code := codes.Unauthenticated
		if oe, ok := err.(*core.OathMeshError); ok {
			if oe.Code == core.ErrPolicyDenied || oe.Code == core.ErrSubjectRevoked {
				code = codes.PermissionDenied
			}
		}
		return s.deny(code, err.Error()), nil
	}

	// Token is valid. Inject verified claims as context headers into the upstream request.
	return s.allow(vcc), nil
}

func (s *envoyExtAuthzServer) deny(grpcCode codes.Code, msg string) *authv3.CheckResponse {
	// Map gRPC codes to HTTP status for Envoy
	httpCode := typev3.StatusCode_Unauthorized
	if grpcCode == codes.PermissionDenied {
		httpCode = typev3.StatusCode_Forbidden
	}

	return &authv3.CheckResponse{
		Status: status.New(grpcCode, msg).Proto(),
		HttpResponse: &authv3.CheckResponse_DeniedResponse{
			DeniedResponse: &authv3.DeniedHttpResponse{
				Status: &typev3.HttpStatus{
					Code: httpCode,
				},
				Body: msg,
			},
		},
	}
}

func (s *envoyExtAuthzServer) allow(vcc *core.VerifiedCallerContext) *authv3.CheckResponse {
	return &authv3.CheckResponse{
		Status: status.New(codes.OK, "OK").Proto(),
		HttpResponse: &authv3.CheckResponse_OkResponse{
			OkResponse: &authv3.OkHttpResponse{
				Headers: []*corev3.HeaderValueOption{
					{
						Header: &corev3.HeaderValue{
							Key:   "x-oathmesh-subject",
							Value: vcc.Principal.Subject,
						},
						AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					},
					{
						Header: &corev3.HeaderValue{
							Key:   "x-oathmesh-action",
							Value: vcc.Action,
						},
						AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					},
					{
						Header: &corev3.HeaderValue{
							Key:   "x-oathmesh-issuer",
							Value: vcc.Principal.Issuer,
						},
						AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					},
				},
			},
		},
	}
}

func main() {
	log.Println("Starting OathMesh Envoy ext_authz server...")

	// Load configuration
	cfg := config.LoadFromEnv()

	issuers := strings.Split(cfg.GatewayIssuers, ",")
	for i, v := range issuers {
		issuers[i] = strings.TrimSpace(v)
	}

	verifierCfg := &verify.VerifierConfig{
		Audience:        cfg.GatewayAudience,
		TrustedIssuers:  issuers,
		JWKSProvider:    verify.NewJWKSCache(verify.DefaultJWKSCacheTTL, nil),
		ReplayCache:     verify.NewMemoryReplayCache(),
	}

	port := os.Getenv("ENVOY_AUTHZ_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	authv3.RegisterAuthorizationServer(grpcServer, &envoyExtAuthzServer{verifierCfg: verifierCfg})

	log.Printf("Listening for Envoy ext_authz requests on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
