package verify

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oathmesh/oathmesh/internal/sign"
)

func makeJWKSFixture(t *testing.T, kid string) ([]byte, ed25519.PublicKey) {
	t.Helper()
	_, pub, err := sign.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}
	jwks, err := sign.BuildJWKS(map[string]ed25519.PublicKey{
		kid: pub,
	})
	if err != nil {
		t.Fatalf("build jwks: %v", err)
	}
	data, err := json.Marshal(jwks)
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}
	return data, pub
}

func TestJWKSCache_TrustedIssuerRegistrationMode(t *testing.T) {
	const kid = "kid-backward-compat"
	jwksBody, pub := makeJWKSFixture(t, kid)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/jwks.json" {
			t.Fatalf("expected /.well-known/jwks.json path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksBody)
	}))
	defer srv.Close()

	cache := NewJWKSCache(60*time.Second, nil)
	if err := cache.RegisterTrustedIssuers([]string{srv.URL}); err != nil {
		t.Fatalf("RegisterTrustedIssuers failed: %v", err)
	}
	key, alg, err := cache.GetKey(srv.URL, kid)
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}
	if alg != "EdDSA" {
		t.Fatalf("expected EdDSA alg, got %s", alg)
	}
	if !bytes.Equal(key, pub) {
		t.Fatal("resolved key does not match expected public key")
	}
}

func TestJWKSCache_UnregisteredIssuerDoesNotFetch(t *testing.T) {
	const kid = "kid-unregistered"
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cache := NewJWKSCache(60*time.Second, nil)
	_, _, err := cache.GetKey(srv.URL, kid)
	if err == nil {
		t.Fatal("expected error for unregistered issuer")
	}
	if !strings.Contains(err.Error(), "no trusted JWKS endpoint configured") {
		t.Fatalf("expected trusted endpoint error, got: %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected 0 outbound requests for unregistered issuer, got %d", requests)
	}
}

func TestJWKSCache_EndpointsModeUsesIssuerKey(t *testing.T) {
	const kid = "kid-endpoints"
	jwksBody, pub := makeJWKSFixture(t, kid)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(jwksBody)
	}))
	defer srv.Close()

	cache := NewJWKSCache(60*time.Second, map[string]string{
		"https://issuer-a.local": srv.URL + "/.well-known/jwks.json",
	})

	key, alg, err := cache.GetKey("https://issuer-a.local", kid)
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}
	if alg != "EdDSA" {
		t.Fatalf("expected EdDSA alg, got %s", alg)
	}
	if !bytes.Equal(key, pub) {
		t.Fatal("resolved key does not match expected public key")
	}
}

func TestJWKSCache_UnconfiguredIssuerReturnsError(t *testing.T) {
	jwksBody, _ := makeJWKSFixture(t, "some-other-kid")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(jwksBody)
	}))
	defer srv.Close()

	cache := NewJWKSCache(60*time.Second, map[string]string{
		"default": srv.URL + "/.well-known/jwks.json",
	})
	_, _, err := cache.GetKey("https://issuer-missing.local", "kid")
	if err == nil {
		t.Fatal("expected error for missing key in default endpoint response")
	}
}
