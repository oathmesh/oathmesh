package main

import (
	"strings"
	"testing"
)

func TestServeRunE_ValidatesStartupConfig(t *testing.T) {
	t.Setenv("OATHMESH_ENV", "production")
	t.Setenv("OATHMESH_ISSUER", "http://issuer.example.com")
	t.Setenv("OATHMESH_KMS_KEY_ID", "test-kms-key")
	t.Setenv("OATHMESH_PRIVATE_KEY", "")
	t.Setenv("OATHMESH_PRIVATE_KEY_B64", "")
	t.Setenv("OATHMESH_PRIVATE_KEY_FILE", "")

	cmd := buildServeCmd()
	err := serveRunE(cmd, nil)
	if err == nil {
		t.Fatalf("expected startup validation error")
	}
	if !strings.Contains(err.Error(), "invalid startup config") || !strings.Contains(err.Error(), "must use HTTPS") {
		t.Fatalf("expected HTTPS validation failure, got: %v", err)
	}
}

func TestServeRunE_GatewayRequiresPolicyInNonDevelopment(t *testing.T) {
	t.Setenv("OATHMESH_ENV", "production")
	t.Setenv("OATHMESH_ISSUER", "https://issuer.example.com")
	t.Setenv("OATHMESH_KMS_KEY_ID", "test-kms-key")
	t.Setenv("OATHMESH_PRIVATE_KEY", "")
	t.Setenv("OATHMESH_PRIVATE_KEY_B64", "")
	t.Setenv("OATHMESH_PRIVATE_KEY_FILE", "")
	t.Setenv("OATHMESH_GATEWAY_POLICY", "")

	cmd := buildServeCmd()
	if err := cmd.Flags().Set("gateway", "true"); err != nil {
		t.Fatalf("set gateway flag: %v", err)
	}

	err := serveRunE(cmd, nil)
	if err == nil {
		t.Fatalf("expected missing gateway policy error")
	}
	if !strings.Contains(err.Error(), "OATHMESH_GATEWAY_POLICY") {
		t.Fatalf("expected gateway policy validation failure, got: %v", err)
	}
}

func TestValidateGatewayPolicyRequirement(t *testing.T) {
	tests := []struct {
		name          string
		gateway       bool
		env           string
		policyFile    string
		wantErrSubstr string
	}{
		{
			name:    "gateway disabled allows empty policy",
			gateway: false,
			env:     "production",
		},
		{
			name:    "development allows empty policy",
			gateway: true,
			env:     "development",
		},
		{
			name:          "non-development requires policy",
			gateway:       true,
			env:           "production",
			policyFile:    "",
			wantErrSubstr: "OATHMESH_GATEWAY_POLICY",
		},
		{
			name:       "non-development accepts configured policy",
			gateway:    true,
			env:        "production",
			policyFile: "policy/example.pkl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGatewayPolicyRequirement(tt.gateway, tt.env, tt.policyFile)
			if tt.wantErrSubstr == "" {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErrSubstr)
			}
			if !strings.Contains(err.Error(), tt.wantErrSubstr) {
				t.Fatalf("expected error containing %q, got: %v", tt.wantErrSubstr, err)
			}
		})
	}
}
