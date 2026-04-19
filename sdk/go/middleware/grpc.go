package middleware

import (
	"google.golang.org/grpc"

	"github.com/oathmesh/oathmesh/internal/middleware"
	"github.com/oathmesh/oathmesh/internal/verify"
)

// OathMeshUnaryInterceptor returns a configured gRPC unary interceptor for OathMesh token verification.
// Use this with grpc.NewServer:
//
//	cfg := &verify.VerifierConfig{
//		Audience:       "my-service",
//		TrustedIssuers: []string{"https://issuer.example.com"},
//		// ... other config
//	}
//	rateLimiter := middleware.NewSimpleRateLimiter(1000)
//	server := grpc.NewServer(
//		grpc.UnaryInterceptor(middleware.OathMeshUnaryInterceptor(cfg, rateLimiter)),
//	)
func OathMeshUnaryInterceptor(
	cfg *verify.VerifierConfig,
	rateLimiter middleware.RateLimiter,
) grpc.UnaryServerInterceptor {
	if rateLimiter == nil {
		rateLimiter = middleware.NewSimpleRateLimiter(1000)
	}
	return middleware.UnaryInterceptor(cfg, rateLimiter)
}

// OathMeshStreamInterceptor returns a configured gRPC stream interceptor for OathMesh token verification.
// Use this with grpc.NewServer:
//
//	cfg := &verify.VerifierConfig{
//		Audience:       "my-service",
//		TrustedIssuers: []string{"https://issuer.example.com"},
//		// ... other config
//	}
//	rateLimiter := middleware.NewSimpleRateLimiter(1000)
//	server := grpc.NewServer(
//		grpc.StreamInterceptor(middleware.OathMeshStreamInterceptor(cfg, rateLimiter)),
//	)
func OathMeshStreamInterceptor(
	cfg *verify.VerifierConfig,
	rateLimiter middleware.RateLimiter,
) grpc.StreamServerInterceptor {
	if rateLimiter == nil {
		rateLimiter = middleware.NewSimpleRateLimiter(1000)
	}
	return middleware.StreamInterceptor(cfg, rateLimiter)
}
