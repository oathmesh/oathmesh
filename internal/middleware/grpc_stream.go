package middleware

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/oathmesh/oathmesh/internal/verify"
)

// StreamInterceptor returns a gRPC stream interceptor that verifies OathMesh tokens at connection start.
func StreamInterceptor(
	verifierCfg *verify.VerifierConfig,
	rateLimiter RateLimiter,
) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		ctx := ss.Context()

		// Extract token from gRPC metadata
		token, err := extractTokenFromMetadata(ctx)
		if err != nil {
			recordMetric(info.FullMethod, codes.Unauthenticated, time.Since(start))
			return status.Error(codes.Unauthenticated, fmt.Sprintf("missing authorization token: %v", err))
		}

		// Verify the token
		vcc, err := verify.Verify(ctx, token, verifierCfg)
		if err != nil {
			// Map OathMesh errors to gRPC codes
			code, msg := mapErrorToGRPCCode(err)
			recordMetric(info.FullMethod, code, time.Since(start))
			return status.Error(code, msg)
		}

		// Check rate limit
		allowed, msg := rateLimiter.Allow(vcc.Principal.Subject)
		if !allowed {
			recordMetric(info.FullMethod, codes.ResourceExhausted, time.Since(start))
			return status.Error(codes.ResourceExhausted, msg)
		}

		// Inject verified claims into context
		newCtx := WithVerifiedCaller(ctx, vcc)

		// Create a wrappedServerStream to use the new context
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		// Call the handler with the enriched context
		err = handler(srv, wrapped)

		// Record metric
		if err != nil {
			code := codes.Internal
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			}
			recordMetric(info.FullMethod, code, time.Since(start))
		} else {
			recordMetric(info.FullMethod, codes.OK, time.Since(start))
		}

		return err
	}
}

// wrappedServerStream is a wrapper around grpc.ServerStream that uses a custom context.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context with injected verified claims.
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
