package issuer

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/oathmesh/oathmesh/internal/sign"
)

// GitLab OIDC Discovery and JWKS URLs
const gitlabIssuer = "https://gitlab.com"
const gitlabJWKSURL = "https://gitlab.com/oauth/discovery/keys"

// GitLabExchangeRequest represents the body for exchanging a GitLab OIDC token.
type GitLabExchangeRequest struct {
	GitLabToken string `json:"gitlab_token"`
}

// GitLabExchangeResponse is the payload returned on successful exchange.
type GitLabExchangeResponse struct {
	Token string `json:"token"`
}

// GitLabJWK and GitLabJWKS reflect the structure of GitLab's JWKS.
type GitLabJWKS struct {
	Keys []GitLabJWK `json:"keys"`
}
type GitLabJWK struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// GitLabIDToken represents the claims from a GitLab CI OIDC token.
type GitLabIDToken struct {
	Iss         string `json:"iss"`
	Sub         string `json:"sub"`
	Aud         string `json:"aud"`
	Exp         int64  `json:"exp"`
	Iat         int64  `json:"iat"`
	ProjectPath string `json:"project_path"`
	PipelineID  string `json:"pipeline_id"`
	JobID       string `json:"job_id"`
	Ref         string `json:"ref"`
}

var gitlabJWKSCache struct {
	jwks  *GitLabJWKS
	until time.Time
	mu    sync.RWMutex
}

func fetchGitLabJWKS(ctx context.Context) (*GitLabJWKS, error) {
	gitlabJWKSCache.mu.RLock()
	if gitlabJWKSCache.jwks != nil && time.Now().Before(gitlabJWKSCache.until) {
		defer gitlabJWKSCache.mu.RUnlock()
		return gitlabJWKSCache.jwks, nil
	}
	gitlabJWKSCache.mu.RUnlock()

	gitlabJWKSCache.mu.Lock()
	defer gitlabJWKSCache.mu.Unlock()

	if gitlabJWKSCache.jwks != nil && time.Now().Before(gitlabJWKSCache.until) {
		return gitlabJWKSCache.jwks, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gitlabJWKSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create JWKS request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned %d", resp.StatusCode)
	}

	var jwks GitLabJWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("decode JWKS: %w", err)
	}

	gitlabJWKSCache.jwks = &jwks
	gitlabJWKSCache.until = time.Now().Add(5 * time.Minute)

	return &jwks, nil
}

func verifyGitLabToken(ctx context.Context, token string) (*GitLabIDToken, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format: expected 3 parts, got %d", len(parts))
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}

	var header struct {
		Kid string `json:"kid"`
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported algorithm: %s", header.Alg)
	}

	jwks, err := fetchGitLabJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}

	var jwk *GitLabJWK
	for _, k := range jwks.Keys {
		if k.Kid == header.Kid {
			jwk = &k
			break
		}
	}
	if jwk == nil {
		return nil, fmt.Errorf("key %s not found in JWKS", header.Kid)
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("decode N: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("decode E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = (e << 8) | int(b)
	}

	pubKey := &rsa.PublicKey{N: n, E: e}

	msg := []byte(parts[0] + "." + parts[1])
	hashed := sha256.Sum256(msg)

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signature); err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	var claims GitLabIDToken
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("parse payload: %w", err)
	}

	if claims.Iss != gitlabIssuer {
		return nil, fmt.Errorf("untrusted issuer: %s", claims.Iss)
	}
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

func (s *Server) exchangeGitLabHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		s.writeError(w, "invalid_content_type", "Content-Type must be application/json", "set Content-Type: application/json header")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	var req GitLabExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "invalid_request", "Failed to parse request body", "check JSON syntax and ensure body is under 64KB")
		return
	}

	if req.GitLabToken == "" {
		s.writeError(w, "claim_missing:gitlab_token", "gitlab_token is required", "Provide gitlab_token in request body")
		return
	}

	gitlabClaims, err := verifyGitLabToken(r.Context(), req.GitLabToken)
	if err != nil {
		s.logger.Error("gitlab token verification failed", "error", err)
		s.writeError(w, "invalid_token", "Failed to verify GitLab OIDC token", "ensure the GitLab OIDC token is valid and not expired")
		return
	}

	sub := fmt.Sprintf("job://gitlab/%s", gitlabClaims.ProjectPath)

	var src *sign.Source
	if gitlabClaims.ProjectPath != "" {
		src = &sign.Source{
			Type:     "gitlab_ci",
			Repo:     gitlabClaims.ProjectPath,
			Workflow: gitlabClaims.Ref,
			RunID:    gitlabClaims.PipelineID,
		}
	}

	mintReq := sign.MintRequest{
		Sub: sub,
		Aud: s.keySet.GetIssuer(),
		Act: "oathmesh.issue",
		Src: src,
	}

	token, err := s.keySet.SignToken(mintReq)
	if err != nil {
		s.logger.Error("exchange mint failed", "error", err)
		s.writeError(w, "mint_failed", "Failed to sign token", "contact the issuer operator")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(GitLabExchangeResponse{Token: token})
}
