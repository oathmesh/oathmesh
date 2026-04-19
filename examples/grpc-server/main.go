package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"

	"github.com/oathmesh/oathmesh/internal/audit"
	"github.com/oathmesh/oathmesh/internal/policy"
	"github.com/oathmesh/oathmesh/internal/verify"
	sdkmiddleware "github.com/oathmesh/oathmesh/sdk/go/middleware"
	"github.com/oathmesh/oathmesh/internal/middleware"
)

// UserServiceServer implements the UserService gRPC server
type UserServiceServer struct {
	// UnimplementedUserServiceServer must be embedded for future compatibility
}

// GetUser retrieves a user by ID
func (s *UserServiceServer) GetUser(ctx context.Context, req interface{}) (interface{}, error) {
	// Extract verified claims from context (injected by middleware)
	// vcc := middleware.VerifiedCallerFrom(ctx)

	// For this example, return a hardcoded user
	return map[string]interface{}{
		"id":    "user-123",
		"name":  "Alice",
		"email": "alice@example.com",
		"role":  "admin",
	}, nil
}

// ListUsers returns all users (streaming)
func (s *UserServiceServer) ListUsers(req interface{}, stream interface{}) error {
	// In a real implementation, this would stream multiple users
	return nil
}

// Health returns the health status
func (s *UserServiceServer) Health(ctx context.Context, req interface{}) (interface{}, error) {
	return map[string]interface{}{
		"status": "ok",
	}, nil
}

func main() {
	// Read configuration from environment variables
	audience := os.Getenv("OATHMESH_AUDIENCE")
	if audience == "" {
		audience = "grpc-example-service"
	}

	trustedIssuersStr := os.Getenv("OATHMESH_TRUSTED_ISSUERS")
	if trustedIssuersStr == "" {
		trustedIssuersStr = "https://issuer.example.com"
	}

	policyPath := os.Getenv("OATHMESH_POLICY_PATH")
	port := os.Getenv("OATHMESH_PORT")
	if port == "" {
		port = "50051"
	}

	// Parse trusted issuers
	issuers := strings.Split(trustedIssuersStr, ",")
	for i, v := range issuers {
		issuers[i] = strings.TrimSpace(v)
	}

	// Setup policy engine if provided
	var evaluator verify.PolicyEvaluator
	if policyPath != "" {
		pe, err := policy.NewWatchedPolicyEngine(policyPath, slog.Default())
		if err != nil {
			log.Fatalf("failed to init policy engine: %v", err)
		}
		evaluator = pe
	}

	// Create verifier config
	verifierCfg := &verify.VerifierConfig{
		Audience:        audience,
		TrustedIssuers:  issuers,
		JWKSProvider:    verify.NewJWKSCache(verify.DefaultJWKSCacheTTL, nil),
		ReplayCache:     verify.NewMemoryReplayCache(),
		PolicyEvaluator: evaluator,
		AuditSink:       audit.NewStdoutAuditSink(),
	}

	// Create rate limiter (1000 tokens per minute per subject)
	rateLimiter := middleware.NewSimpleRateLimiter(1000)

	// Create gRPC server with OathMesh middleware
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(sdkmiddleware.OathMeshUnaryInterceptor(verifierCfg, rateLimiter)),
		grpc.StreamInterceptor(sdkmiddleware.OathMeshStreamInterceptor(verifierCfg, rateLimiter)),
	}

	server := grpc.NewServer(opts...)
	defer server.Stop()

	// Register service (in real implementation, would register actual proto service)
	// pb.RegisterUserServiceServer(server, &UserServiceServer{})

	// Listen on the specified port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	log.Printf("gRPC server listening on port %s with OathMesh authentication enabled", port)
	log.Printf("Audience: %s", audience)
	log.Printf("Trusted issuers: %v", issuers)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
