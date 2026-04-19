package middleware_test

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/middleware"
	"github.com/oathmesh/oathmesh/internal/sign"
	"github.com/oathmesh/oathmesh/internal/verify"
)

const (
	testIssuer   = "https://issuer.test.oathmesh.tech"
	testAudience = "https://api.test.internal"
	testKid      = "grpc-test-kid"
)

func mintValidToken(t *testing.T, subject string) (*verify.VerifierConfig, string) {
	t.Helper()

	privateKey, publicKey, err := sign.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keys: %v", err)
	}

	now := time.Now().Unix()
	token, err := sign.BuildJWS(
		sign.Header{Typ: sign.TypeHeader, Alg: sign.AlgEdDSA, Kid: testKid},
		sign.Claims{
			Iss:   testIssuer,
			Sub:   subject,
			Aud:   testAudience,
			Act:   "inventory.read",
			Iat:   now,
			Exp:   now + 120,
			JTI:   fmt.Sprintf("jti-%d", time.Now().UnixNano()),
			Scope: []string{"action:read:user"},
		},
		privateKey,
	)
	if err != nil {
		t.Fatalf("mint token: %v", err)
	}

	cfg := &verify.VerifierConfig{
		Audience:       testAudience,
		TrustedIssuers: []string{testIssuer},
		JWKSProvider:   verify.NewStaticJWKSProvider(map[string]ed25519.PublicKey{testKid: publicKey}),
	}

	return cfg, token
}

type fixedRateLimiter struct {
	allowed bool
	msg     string
}

func (f fixedRateLimiter) Allow(string) (bool, string) {
	return f.allowed, f.msg
}

func TestUnaryInterceptor_ValidToken_ContextInjectionPath(t *testing.T) {
	cfg, token := mintValidToken(t, "svc://integration/grpc-unary")
	interceptor := middleware.UnaryInterceptor(cfg, fixedRateLimiter{allowed: true})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "OathMesh "+token))
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/TestUnary"}

	called := false
	resp, err := interceptor(ctx, "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		vcc := middleware.VerifiedCallerFrom(ctx)
		if vcc == nil {
			t.Fatal("expected verified caller context")
		}
		if vcc.Principal.Subject != "svc://integration/grpc-unary" {
			t.Fatalf("unexpected subject: %s", vcc.Principal.Subject)
		}
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
	if resp != "ok" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestUnaryInterceptor_InvalidToken(t *testing.T) {
	cfg, _ := mintValidToken(t, "svc://integration/grpc-invalid")
	interceptor := middleware.UnaryInterceptor(cfg, fixedRateLimiter{allowed: true})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "OathMesh not-a-token"))
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/TestUnary"}

	resp, err := interceptor(ctx, nil, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler must not run when token is invalid")
		return nil, nil
	})

	if err == nil {
		t.Fatal("expected unauthenticated error")
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %v", resp)
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", status.Code(err))
	}
}

func TestUnaryInterceptor_MissingToken(t *testing.T) {
	cfg, _ := mintValidToken(t, "svc://integration/grpc-missing")
	interceptor := middleware.UnaryInterceptor(cfg, fixedRateLimiter{allowed: true})
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/TestUnary"}

	_, err := interceptor(context.Background(), nil, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler must not run when token is missing")
		return nil, nil
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", status.Code(err))
	}
}

func TestUnaryInterceptor_RateLimitExceeded(t *testing.T) {
	cfg, token := mintValidToken(t, "svc://integration/grpc-rate-limited")
	interceptor := middleware.UnaryInterceptor(cfg, fixedRateLimiter{allowed: false, msg: "rate limit exceeded"})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "OathMesh "+token))
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/TestUnary"}

	_, err := interceptor(ctx, nil, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler must not run when rate-limited")
		return nil, nil
	})
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("expected resource exhausted, got %v", status.Code(err))
	}
}

func TestStreamInterceptor_ValidToken_ContextInjectionPath(t *testing.T) {
	cfg, token := mintValidToken(t, "svc://integration/grpc-stream")
	interceptor := middleware.StreamInterceptor(cfg, fixedRateLimiter{allowed: true})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "OathMesh "+token))
	stream := &mockServerStream{ctx: ctx}
	info := &grpc.StreamServerInfo{FullMethod: "/test.Service/TestStream"}

	called := false
	err := interceptor(nil, stream, info, func(srv interface{}, ss grpc.ServerStream) error {
		called = true
		vcc := middleware.VerifiedCallerFrom(ss.Context())
		if vcc == nil {
			t.Fatal("expected verified caller context on stream")
		}
		if vcc.Principal.Subject != "svc://integration/grpc-stream" {
			t.Fatalf("unexpected subject: %s", vcc.Principal.Subject)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !called {
		t.Fatal("expected stream handler to run")
	}
}

func TestStreamInterceptor_InvalidToken(t *testing.T) {
	cfg, _ := mintValidToken(t, "svc://integration/grpc-stream-invalid")
	interceptor := middleware.StreamInterceptor(cfg, fixedRateLimiter{allowed: true})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "OathMesh not-a-token"))
	stream := &mockServerStream{ctx: ctx}
	info := &grpc.StreamServerInfo{FullMethod: "/test.Service/TestStream"}

	err := interceptor(nil, stream, info, func(srv interface{}, ss grpc.ServerStream) error {
		t.Fatal("handler must not run for invalid token")
		return nil
	})

	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", status.Code(err))
	}
}

func TestStreamInterceptor_MissingToken(t *testing.T) {
	cfg, _ := mintValidToken(t, "svc://integration/grpc-stream-missing")
	interceptor := middleware.StreamInterceptor(cfg, fixedRateLimiter{allowed: true})

	stream := &mockServerStream{ctx: context.Background()}
	info := &grpc.StreamServerInfo{FullMethod: "/test.Service/TestStream"}

	err := interceptor(nil, stream, info, func(srv interface{}, ss grpc.ServerStream) error {
		t.Fatal("handler must not run without token")
		return nil
	})

	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", status.Code(err))
	}
}

func TestVerifiedCallerFrom_WithoutInjectedContext(t *testing.T) {
	if got := middleware.VerifiedCallerFrom(context.Background()); got != nil {
		t.Fatalf("expected nil context, got %+v", got)
	}

	vcc := &core.VerifiedCallerContext{Principal: core.Principal{Issuer: testIssuer, Subject: "svc://integration/claims"}}
	ctx := context.WithValue(context.Background(), "oathmesh:verified_claims", vcc)

	got := middleware.VerifiedCallerFrom(ctx)
	if got == nil || got.Principal.Subject != "svc://integration/claims" {
		t.Fatal("expected to retrieve injected verified caller context")
	}
}

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func (m *mockServerStream) SendHeader(metadata.MD) error { return nil }
func (m *mockServerStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockServerStream) SetTrailer(metadata.MD)       {}
func (m *mockServerStream) SendMsg(interface{}) error    { return nil }
func (m *mockServerStream) RecvMsg(interface{}) error    { return nil }
