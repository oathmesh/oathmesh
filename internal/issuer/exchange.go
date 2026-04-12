package issuer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/oathmesh/oathmesh/internal/sign"
)

const (
	GitHubJWKSURL = "https://token.actions.githubusercontent.com/.well-known/jwks"
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

type JWK struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
	X   string `json:"x,omitempty"`
	Crv string `json:"crv,omitempty"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
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

	parts := strings.Split(req.GitHubToken, ".")
	if len(parts) != 3 {
		s.writeError(w, "invalid_token", "Invalid GitHub OIDC token format", "Token must have 3 parts")
		return
	}

	githubClaims, err := parseGitHubToken(req.GitHubToken)
	if err != nil {
		s.writeError(w, "invalid_token", "Failed to parse GitHub OIDC token", err.Error())
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

func parseGitHubToken(token string) (*GitHubIDToken, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := decodeBase64URL(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	var claims GitHubIDToken
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal claims: %w", err)
	}

	if claims.Issuer != "https://token.actions.githubusercontent.com" {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	return &claims, nil
}

func decodeBase64URL(s string) ([]byte, error) {
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	result := make([]byte, len(s))
	n, err := base64URLDecode(result, []byte(s))
	if err != nil {
		return nil, err
	}
	return result[:n], nil
}

func base64URLDecode(dst, src []byte) (int, error) {
	for i := 0; i < len(src); i++ {
		switch {
		case src[i] >= 'A' && src[i] <= 'Z':
			dst[i] = src[i] - 'A'
		case src[i] >= 'a' && src[i] <= 'z':
			dst[i] = src[i] - 'a' + 26
		case src[i] >= '0' && src[i] <= '9':
			dst[i] = src[i] - '0' + 52
		case src[i] == '-':
			dst[i] = 62
		case src[i] == '_':
			dst[i] = 63
		default:
			return 0, fmt.Errorf("invalid base64url character: %c", src[i])
		}
	}
	return len(src), nil
}

func init() {
	_ = time.Time{}
}
