package config

import (
	"strings"
	"testing"
)

func TestValidate_RequiresSigningKeyWithoutKMS(t *testing.T) {
	t.Setenv("OATHMESH_ENV", "production")
	t.Setenv("OATHMESH_ISSUER", "https://issuer.example.com")
	t.Setenv("OATHMESH_KMS_KEY_ID", "")
	t.Setenv("OATHMESH_PRIVATE_KEY", "")
	t.Setenv("OATHMESH_PRIVATE_KEY_B64", "")
	t.Setenv("OATHMESH_PRIVATE_KEY_FILE", "")

	cfg := LoadFromEnv()
	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected validation error when no signing key is configured")
	}
	if !strings.Contains(err.Error(), "OATHMESH_PRIVATE_KEY") {
		t.Fatalf("expected key requirement error, got: %v", err)
	}
}

func TestValidate_AllowsKMSWithoutLocalKey(t *testing.T) {
	t.Setenv("OATHMESH_ENV", "production")
	t.Setenv("OATHMESH_ISSUER", "https://issuer.example.com")
	t.Setenv("OATHMESH_KMS_KEY_ID", "test-kms-key")
	t.Setenv("OATHMESH_PRIVATE_KEY", "")
	t.Setenv("OATHMESH_PRIVATE_KEY_B64", "")
	t.Setenv("OATHMESH_PRIVATE_KEY_FILE", "")

	cfg := LoadFromEnv()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected KMS mode to pass validation without local key, got: %v", err)
	}
}

func TestValidate_RequiresHTTPSOutsideDevelopment(t *testing.T) {
	t.Setenv("OATHMESH_ENV", "production")
	t.Setenv("OATHMESH_ISSUER", "http://issuer.example.com")
	t.Setenv("OATHMESH_PRIVATE_KEY", "dummy")

	cfg := LoadFromEnv()
	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected HTTPS validation error")
	}
	if !strings.Contains(err.Error(), "must use HTTPS") {
		t.Fatalf("expected HTTPS error, got: %v", err)
	}
}
