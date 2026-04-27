package main

import (
	"testing"
	"time"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/oathmesh/oathmesh/internal/core"
	"google.golang.org/grpc/codes"
)

// A mock verify function would be ideal, but for now we'll test the allow/deny response structures.

func TestEnvoyExtAuthzServer_Deny(t *testing.T) {
	s := &envoyExtAuthzServer{}

	resp := s.deny(codes.Unauthenticated, "missing token")

	if resp.Status.Code != int32(codes.Unauthenticated) {
		t.Errorf("expected Unauthenticated code, got %v", resp.Status.Code)
	}

	deniedResp := resp.HttpResponse.(*authv3.CheckResponse_DeniedResponse).DeniedResponse
	if deniedResp.Status.Code != typev3.StatusCode_Unauthorized {
		t.Errorf("expected 401 Unauthorized HTTP status, got %v", deniedResp.Status.Code)
	}
	if deniedResp.Body != "missing token" {
		t.Errorf("expected body 'missing token', got '%s'", deniedResp.Body)
	}
}

func TestEnvoyExtAuthzServer_DenyForbidden(t *testing.T) {
	s := &envoyExtAuthzServer{}

	resp := s.deny(codes.PermissionDenied, "policy denied")

	if resp.Status.Code != int32(codes.PermissionDenied) {
		t.Errorf("expected PermissionDenied code, got %v", resp.Status.Code)
	}

	deniedResp := resp.HttpResponse.(*authv3.CheckResponse_DeniedResponse).DeniedResponse
	if deniedResp.Status.Code != typev3.StatusCode_Forbidden {
		t.Errorf("expected 403 Forbidden HTTP status, got %v", deniedResp.Status.Code)
	}
}

func TestEnvoyExtAuthzServer_Allow(t *testing.T) {
	s := &envoyExtAuthzServer{}

	vcc := &core.VerifiedCallerContext{
		Principal: core.Principal{
			Subject: "svc://test",
			Issuer:  "https://issuer.local",
		},
		Action:    "read",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		TokenID:   "test-jti",
	}

	resp := s.allow(vcc)

	if resp.Status.Code != int32(codes.OK) {
		t.Errorf("expected OK code, got %v", resp.Status.Code)
	}

	okResp := resp.HttpResponse.(*authv3.CheckResponse_OkResponse).OkResponse
	headers := okResp.Headers

	foundSub := false
	for _, h := range headers {
		if h.Header.Key == "x-oathmesh-subject" && h.Header.Value == "svc://test" {
			foundSub = true
		}
	}

	if !foundSub {
		t.Error("expected x-oathmesh-subject header to be injected")
	}
}
