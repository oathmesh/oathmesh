package issuer

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/oathmesh/oathmesh/internal/core"
	internalmiddleware "github.com/oathmesh/oathmesh/internal/middleware"
	"github.com/oathmesh/oathmesh/internal/sign"
	"github.com/oathmesh/oathmesh/internal/verify"
)

const (
	phase2MintSecret = "phase2-integration-secret"
	phase2Audience   = "https://inventory.integration.internal"
	phase2Issuer     = "https://issuer.integration.local"
	phase2Kid        = "phase2-integration-kid"
)

type integrationSigner struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	issuer     string
	kid        string
}

func newIntegrationSigner(t *testing.T) *integrationSigner {
	t.Helper()

	privateKey, publicKey, err := sign.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}

	return &integrationSigner{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     phase2Issuer,
		kid:        phase2Kid,
	}
}

func (s *integrationSigner) GetIssuer() string { return s.issuer }

func (s *integrationSigner) JWKS() (*sign.JWKS, error) {
	return sign.BuildJWKS(map[string]ed25519.PublicKey{s.kid: s.publicKey})
}

func (s *integrationSigner) SignToken(req sign.MintRequest) (string, error) {
	return sign.SignToken(req, s.issuer, s.privateKey, s.kid)
}

type staticRevocationList struct {
	revoked map[string]bool
}

func (s staticRevocationList) IsRevoked(_ context.Context, subject string) (bool, error) {
	return s.revoked[subject], nil
}

type fixedRateLimiter struct {
	allowed bool
	msg     string
}

func (f fixedRateLimiter) Allow(string) (bool, string) {
	return f.allowed, f.msg
}

func mintTokenViaIssuer(t *testing.T, server *Server, req MintRequest) string {
	t.Helper()

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal mint request: %v", err)
	}

	httpReq := httptest.NewRequest(http.MethodPost, "/v1/token", bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer "+phase2MintSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.router().ServeHTTP(rec, httpReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("mint request failed: status=%d body=%s", rec.Code, rec.Body.String())
	}

	var mintResp MintResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &mintResp); err != nil {
		t.Fatalf("decode mint response: %v", err)
	}
	if mintResp.Token == "" {
		t.Fatal("expected non-empty token in mint response")
	}

	return mintResp.Token
}

func verifierConfig(signer *integrationSigner) *verify.VerifierConfig {
	return &verify.VerifierConfig{
		Audience:       phase2Audience,
		TrustedIssuers: []string{signer.issuer},
		JWKSProvider: verify.NewStaticJWKSProvider(map[string]ed25519.PublicKey{
			signer.kid: signer.publicKey,
		}),
	}
}

func TestPhase2IssuerMintVerifyHappyPath(t *testing.T) {
	t.Setenv("OATHMESH_MINT_SECRET", phase2MintSecret)

	signer := newIntegrationSigner(t)
	server := NewServer(signer)

	token := mintTokenViaIssuer(t, server, MintRequest{
		Sub:   "svc://phase2/happy-path",
		Aud:   phase2Audience,
		Act:   "inventory.read",
		Scope: []string{"action:read:user"},
	})

	vcc, err := verify.Verify(context.Background(), token, verifierConfig(signer))
	if err != nil {
		t.Fatalf("expected token verification success, got %v", err)
	}
	if vcc.Principal.Subject != "svc://phase2/happy-path" {
		t.Fatalf("unexpected subject %q", vcc.Principal.Subject)
	}
}

func TestPhase2RevocationAndReplayBehavior(t *testing.T) {
	t.Setenv("OATHMESH_MINT_SECRET", phase2MintSecret)

	signer := newIntegrationSigner(t)
	server := NewServer(signer)

	token := mintTokenViaIssuer(t, server, MintRequest{
		Sub: "agent://phase2/security-agent",
		Aud: phase2Audience,
		Act: "inventory.write",
	})

	replayCfg := verifierConfig(signer)
	replayCfg.ReplayCache = verify.NewMemoryReplayCache()

	if _, err := verify.Verify(context.Background(), token, replayCfg); err != nil {
		t.Fatalf("first verify should succeed, got %v", err)
	}
	if _, err := verify.Verify(context.Background(), token, replayCfg); err == nil {
		t.Fatal("expected replay detection error on second verify")
	} else if ome, ok := err.(*core.OathMeshError); !ok || ome.Code != core.ErrReplayDetected {
		t.Fatalf("expected replay_detected, got %v", err)
	}

	revocationCfg := verifierConfig(signer)
	revocationCfg.RevocationList = staticRevocationList{
		revoked: map[string]bool{"agent://phase2/security-agent": true},
	}

	if _, err := verify.Verify(context.Background(), token, revocationCfg); err == nil {
		t.Fatal("expected subject_revoked error for revoked subject")
	} else if ome, ok := err.(*core.OathMeshError); !ok || ome.Code != core.ErrSubjectRevoked {
		t.Fatalf("expected subject_revoked, got %v", err)
	}
}

func TestPhase2GRPCMiddlewareAuthAndRateLimitBehavior(t *testing.T) {
	t.Setenv("OATHMESH_MINT_SECRET", phase2MintSecret)

	signer := newIntegrationSigner(t)
	server := NewServer(signer)
	token := mintTokenViaIssuer(t, server, MintRequest{
		Sub: "svc://phase2/grpc-client",
		Aud: phase2Audience,
		Act: "inventory.read",
	})

	cfg := verifierConfig(signer)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "OathMesh "+token))
	info := &grpc.UnaryServerInfo{FullMethod: "/phase2.Service/TestUnary"}

	allowInterceptor := internalmiddleware.UnaryInterceptor(cfg, fixedRateLimiter{allowed: true})
	called := false
	resp, err := allowInterceptor(ctx, "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		vcc := internalmiddleware.VerifiedCallerFrom(ctx)
		if vcc == nil || vcc.Principal.Subject != "svc://phase2/grpc-client" {
			t.Fatalf("expected verified caller with grpc subject, got %+v", vcc)
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("expected grpc auth success, got %v", err)
	}
	if !called || resp != "ok" {
		t.Fatalf("expected grpc handler to run with ok response, called=%v resp=%v", called, resp)
	}

	denyInterceptor := internalmiddleware.UnaryInterceptor(cfg, fixedRateLimiter{allowed: false, msg: "rate limit exceeded"})
	_, err = denyInterceptor(ctx, "request", info, func(context.Context, interface{}) (interface{}, error) {
		t.Fatal("handler must not execute when rate-limited")
		return nil, nil
	})
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("expected resource exhausted, got %v", status.Code(err))
	}
}
