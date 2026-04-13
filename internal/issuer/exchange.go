package issuer

import (
	"context"
	"crypto"
	"crypto/rsa"
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

const (
	GitHubJWKSURL = "https://token.actions.githubusercontent.com/.well-known/jwks"
	GitHubIssuer  = "https://token.actions.githubusercontent.com"
)

type GitHubExchangeRequest struct {
	GitHubToken string `json:"github_token"`
}

type GitHubExchangeResponse struct {
	Token string `json:"token"`
}

type GitHubIDToken struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Aud       string `json:"aud"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Job       struct {
		Workflow string `json:"workflow_name,omitempty"`
		RunID    string `json:"run_id,omitempty"`
	} `json:"job,omitempty"`
	Repository struct {
		FullName string `json:"full_name,omitempty"`
		Name     string `json:"name,omitempty"`
		Owner    struct {
			Login string `json:"login,omitempty"`
		} `json:"owner,omitempty"`
	} `json:"repository,omitempty"`
	Workflow string `json:"workflow,omitempty"`
	SHA      string `json:"sha,omitempty"`
	Ref      string `json:"ref,omitempty"`
}

type GitHubJWK struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type GitHubJWKS struct {
	Keys []GitHubJWK `json:"keys"`
}

var (
	// githubJWKSMu synchronizes concurrent reads and writes to the cache payload.
	// We mandate explicit RLock usage for standard token resolutions preventing
	// data races under load, and escalate to full Lock access linearly during HTTP fetch.
	githubJWKSMu    sync.RWMutex
	
	// githubJWKS and githubJWKSCache strictly require githubJWKSMu locking before access.
	githubJWKS      *GitHubJWKS
	githubJWKSCache *cacheEntry
)

type cacheEntry struct {
	jwks  *GitHubJWKS
	until time.Time
}

func fetchGitHubJWKS(ctx context.Context) (*GitHubJWKS, error) {
	// Fast path: read-only cache check under RLock
	githubJWKSMu.RLock()
	if githubJWKS != nil && githubJWKSCache != nil && time.Now().Before(githubJWKSCache.until) {
		result := githubJWKS
		githubJWKSMu.RUnlock()
		return result, nil
	}
	githubJWKSMu.RUnlock()

	// Slow path: fetch under write lock
	githubJWKSMu.Lock()
	defer githubJWKSMu.Unlock()

	// Double-check after acquiring write lock
	if githubJWKS != nil && githubJWKSCache != nil && time.Now().Before(githubJWKSCache.until) {
		return githubJWKS, nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", GitHubJWKSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned %d", resp.StatusCode)
	}

	var jwks GitHubJWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("decode JWKS: %w", err)
	}

	githubJWKS = &jwks
	githubJWKSCache = &cacheEntry{
		jwks:  &jwks,
		until: time.Now().Add(5 * time.Minute),
	}

	return &jwks, nil
}

func verifyGitHubToken(ctx context.Context, token string) (*GitHubIDToken, error) {
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

	jwks, err := fetchGitHubJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}

	var jwk *GitHubJWK
	for _, k := range jwks.Keys {
		if k.Kid == header.Kid {
			jwk = &k
			break
		}
	}
	if jwk == nil {
		return nil, fmt.Errorf("key not found: %s", header.Kid)
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	var claims GitHubIDToken
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("parse claims: %w", err)
	}

	if claims.Issuer != GitHubIssuer {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("decode JWK N: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("decode JWK E: %w", err)
	}
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	rsaKey := &rsa.PublicKey{
		N: n,
		E: e,
	}

	signingInput := parts[0] + "." + parts[1]
	h := crypto.SHA256.New()
	h.Write([]byte(signingInput))
	digest := h.Sum(nil)

	if err := rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, digest, signature); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	return &claims, nil
}

func (s *Server) exchangeGitHubHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		s.writeError(w, "invalid_content_type", "Content-Type must be application/json", "")
		return
	}

	var req GitHubExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "invalid_request", "Failed to parse request body", err.Error())
		return
	}

	if req.GitHubToken == "" {
		s.writeError(w, "claim_missing:github_token", "github_token is required", "Provide github_token in request body")
		return
	}

	githubClaims, err := verifyGitHubToken(r.Context(), req.GitHubToken)
	if err != nil {
		s.writeError(w, "invalid_token", "Failed to verify GitHub OIDC token", err.Error())
		return
	}

	sub := fmt.Sprintf("job://github/%s/%s", githubClaims.Repository.FullName, githubClaims.Workflow)

	var src *sign.Source
	if githubClaims.Repository.FullName != "" {
		src = &sign.Source{
			Type:     "github_actions",
			Repo:     githubClaims.Repository.FullName,
			Workflow: githubClaims.Workflow,
			RunID:    githubClaims.Job.RunID,
			SHA:      githubClaims.SHA,
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
		s.writeError(w, "mint_failed", "Failed to sign token", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GitHubExchangeResponse{Token: token})
}
