package sign

import (
	"testing"

	"github.com/google/uuid"
)

func TestSignToken_TTLClamped(t *testing.T) {
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	tests := []struct {
		name        string
		ttlHint     int
		expectedTTL int
	}{
		{"default TTL", 0, DefaultTTL},
		{"custom TTL", 60, 60},
		{"TTL exceeds max", 500, MaxTTL},
		{"TTL at max", 300, 300},
		{"negative TTL", -10, DefaultTTL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := MintRequest{
				Sub: "agent://repo/acme/test",
				Aud: "https://api.example.com",
				Act: "test.read",
				TTL: tt.ttlHint,
			}

			token, err := SignToken(req, "https://issuer.example.com", privateKey, "test-key-001")
			if err != nil {
				t.Fatalf("failed to sign token: %v", err)
			}

			claims, err := UnverifiedClaims(token)
			if err != nil {
				t.Fatalf("failed to parse claims: %v", err)
			}

			actualTTL := int(claims.Exp - claims.Iat)
			if actualTTL != tt.expectedTTL {
				t.Errorf("expected TTL %d, got %d", tt.expectedTTL, actualTTL)
			}
		})
	}
}

func TestParsePrivateKeyPEM_ValidKey(t *testing.T) {
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	pemBytes, err := MarshalPrivateKeyToPEM(privateKey)
	if err != nil {
		t.Fatalf("failed to marshal key: %v", err)
	}

	parsedKey, err := ParsePrivateKeyPEM(string(pemBytes))
	if err != nil {
		t.Fatalf("failed to parse key: %v", err)
	}

	if parsedKey.Equal(privateKey) == false {
		t.Error("parsed key does not match original")
	}
}

func TestParsePrivateKeyPEM_InvalidKey(t *testing.T) {
	invalidPEM := []byte(`-----BEGIN PRIVATE KEY-----
invalid
-----END PRIVATE KEY-----`)

	_, err := ParsePrivateKeyPEM(string(invalidPEM))
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestJTI_IsUUID(t *testing.T) {
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	req := MintRequest{
		Sub: "agent://repo/acme/test",
		Aud: "https://api.example.com",
		Act: "test.read",
	}

	token, err := SignToken(req, "https://issuer.example.com", privateKey, "test-key-001")
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	claims, err := UnverifiedClaims(token)
	if err != nil {
		t.Fatalf("failed to parse claims: %v", err)
	}

	_, err = uuid.Parse(claims.JTI)
	if err != nil {
		t.Errorf("JTI is not a valid UUID: %v", err)
	}
}

func TestBuildJWKS(t *testing.T) {
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	_ = privateKey

	publicKeys := map[string]PublicKey{
		"test-key-001": publicKey,
	}

	jwks, err := BuildJWKS(publicKeys)
	if err != nil {
		t.Fatalf("failed to build JWKS: %v", err)
	}

	if len(jwks.Keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(jwks.Keys))
	}

	if jwks.Keys[0].Kid != "test-key-001" {
		t.Errorf("expected kid test-key-001, got %s", jwks.Keys[0].Kid)
	}

	if jwks.Keys[0].Kty != "OKP" {
		t.Errorf("expected KTY OKP, got %s", jwks.Keys[0].Kty)
	}
}
