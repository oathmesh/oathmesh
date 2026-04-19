package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/metrics"
	"github.com/oathmesh/oathmesh/internal/verify"
)

// UnaryInterceptor returns a gRPC unary interceptor that verifies OathMesh tokens.
func UnaryInterceptor(
	verifierCfg *verify.VerifierConfig,
	rateLimiter RateLimiter,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Extract token from gRPC metadata
		token, err := extractTokenFromMetadata(ctx)
		if err != nil {
			recordMetric(info.FullMethod, codes.Unauthenticated, time.Since(start))
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("missing authorization token: %v", err))
		}

		// Verify the token
		vcc, err := verify.Verify(ctx, token, verifierCfg)
		if err != nil {
			// Map OathMesh errors to gRPC codes
			code, msg := mapErrorToGRPCCode(err)
			recordMetric(info.FullMethod, code, time.Since(start))
			return nil, status.Error(code, msg)
		}

		// Check rate limit
		allowed, msg := rateLimiter.Allow(vcc.Principal.Subject)
		if !allowed {
			recordMetric(info.FullMethod, codes.ResourceExhausted, time.Since(start))
			return nil, status.Error(codes.ResourceExhausted, msg)
		}

		// Inject verified claims into context
		newCtx := WithVerifiedCaller(ctx, vcc)

		// Call the handler with the enriched context
		result, err := handler(newCtx, req)

		// Record success metric
		if err != nil {
			// gRPC error occurred in handler
			recordMetric(info.FullMethod, codes.Internal, time.Since(start))
		} else {
			recordMetric(info.FullMethod, codes.OK, time.Since(start))
		}

		return result, err
	}
}

// extractTokenFromMetadata extracts the authorization token from gRPC metadata.
// Expects "authorization" metadata key with value "Bearer <token>" or "OathMesh <token>".
func extractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", fmt.Errorf("missing authorization header")
	}

	authHeader := authHeaders[0]

	// Support both "Bearer" and "OathMesh" prefixes
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}
	if strings.HasPrefix(authHeader, "OathMesh ") {
		return strings.TrimPrefix(authHeader, "OathMesh "), nil
	}

	return "", fmt.Errorf("invalid authorization header format")
}

// mapErrorToGRPCCode maps OathMeshError to gRPC error codes.
func mapErrorToGRPCCode(err error) (codes.Code, string) {
	if oe, ok := err.(*core.OathMeshError); ok {
		switch oe.Code {
		case core.ErrClaimMissing, core.ErrSignatureInvalid, core.ErrIssuerUntrusted, core.ErrTokenExpired:
			return codes.Unauthenticated, oe.Message
		case core.ErrAudienceMismatch:
			return codes.Unauthenticated, oe.Message
		case core.ErrAlgorithmNotAllowed:
			return codes.Unauthenticated, oe.Message
		case core.ErrPolicyDenied:
			return codes.PermissionDenied, oe.Message
		case core.ErrSubjectRevoked:
			return codes.PermissionDenied, oe.Message
		case core.ErrReplayDetected:
			return codes.Unauthenticated, oe.Message
		case core.ErrBindingMismatch, core.ErrBindingRequired:
			return codes.Unauthenticated, oe.Message
		default:
			return codes.Internal, "verification failed"
		}
	}

	return codes.Internal, "internal verification error"
}

// recordMetric records gRPC request metrics to Prometheus.
func recordMetric(fullMethod string, code codes.Code, duration time.Duration) {
	// Extract method name from full method (e.g., "/package.Service/Method")
	methodName := fullMethod
	if idx := strings.LastIndex(fullMethod, "/"); idx >= 0 {
		methodName = fullMethod[idx+1:]
	}

	// Record request count
	metrics.GRPCRequestsTotal.WithLabelValues(methodName, code.String()).Inc()

	// Record request duration
	metrics.GRPCRequestDuration.WithLabelValues(methodName).Observe(duration.Seconds())
}

// VerifiedCallerFrom extracts the verified caller context from a gRPC context.
func VerifiedCallerFrom(ctx context.Context) *core.VerifiedCallerContext {
	vcc, ok := ctx.Value(contextKeyVerifiedClaims).(*core.VerifiedCallerContext)
	if !ok {
		return nil
	}
	return vcc
}
