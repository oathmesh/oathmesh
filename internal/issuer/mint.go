package issuer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/oathmesh/oathmesh/internal/sign"
)

// ValidSubjectSchemes mirrors the CLI validation for subject URI schemes.
// OathMesh subjects MUST use one of these standardized URI schemes.
var ValidSubjectSchemes = []string{
	"svc://",   // services and microservices
	"agent://", // AI agents and bots
	"job://",   // CI/CD jobs
	"tool://",  // MCP-adjacent tool clients
	"user://",  // human delegation context only
}

type MintRequest struct {
	Sub    string       `json:"sub"`
	Aud    string       `json:"aud"`
	Act    string       `json:"act"`
	TTL    int          `json:"ttl_hint,omitempty"`
	Scope  []string     `json:"scope,omitempty"`
	Reason string       `json:"reason,omitempty"`
	Env    string       `json:"env,omitempty"`
	RQH    string       `json:"rqh,omitempty"`
	Src    *sign.Source `json:"src,omitempty"`
}

type MintResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	TokenType string `json:"token_type"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Fix     string `json:"fix,omitempty"`
}

func (s *Server) mintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Rate limit using the resolved client IP (after chi middleware.RealIP).
	// chi's RealIP middleware sets r.RemoteAddr to the real client IP,
	// but it may still include a port. Strip the port for consistent keying.
	ip := clientIP(r)
	if s.rateLimiter != nil && !s.rateLimiter.Allow(ip) {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		s.writeError(w, "invalid_content_type", "Content-Type must be application/json", "set Content-Type: application/json header")
		return
	}

	// Enforce request body size limit to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	var req MintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "invalid_request", "Failed to parse request body", "check JSON syntax and ensure body is under 64KB")
		return
	}

	if req.Sub == "" {
		s.writeError(w, "claim_missing:sub", "sub claim is required", "Provide sub in request body")
		return
	}
	if req.Aud == "" {
		s.writeError(w, "claim_missing:aud", "aud claim is required", "Provide aud in request body")
		return
	}
	if req.Act == "" {
		s.writeError(w, "claim_missing:act", "act claim is required", "Provide act in request body")
		return
	}

	// Validate subject URI scheme (same rules as CLI)
	if err := validateSubjectURI(req.Sub); err != nil {
		s.writeError(w, "invalid_subject", err.Error(), "subject must use svc://, agent://, job://, tool://, or user:// scheme")
		return
	}

	signReq := sign.MintRequest{
		Sub:    req.Sub,
		Aud:    req.Aud,
		Act:    req.Act,
		TTL:    req.TTL,
		Scope:  req.Scope,
		Reason: req.Reason,
		Env:    req.Env,
		RQH:    req.RQH,
		Src:    req.Src,
	}

	token, err := s.keySet.SignToken(signReq)
	if err != nil {
		// Do NOT expose internal signing errors to caller
		s.logger.Error("mint failed", "error", err)
		s.writeError(w, "mint_failed", "Failed to sign token", "contact the issuer operator")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(MintResponse{
		Token:     token,
		ExpiresIn: computeExpiresIn(req, token),
		TokenType: "OathMesh",
	})
}

// validateSubjectURI checks that a subject string starts with a known URI scheme.
func validateSubjectURI(sub string) error {
	for _, scheme := range ValidSubjectSchemes {
		if strings.HasPrefix(sub, scheme) {
			return nil
		}
	}
	return fmt.Errorf(
		"invalid subject URI %q: must start with one of: svc://, agent://, job://, tool://, user://",
		sub,
	)
}

// clientIP extracts the client IP from the request, stripping the port if present.
// After chi's middleware.RealIP runs, r.RemoteAddr contains the resolved IP.
func clientIP(r *http.Request) string {
	ip := r.RemoteAddr
	// Strip port if present (e.g. "192.168.1.1:12345" → "192.168.1.1")
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		// Check if this is an IPv6 address in brackets
		if strings.Contains(ip, "]") {
			// IPv6: "[::1]:port" → "[::1]"
			if bracketIdx := strings.LastIndex(ip, "]"); bracketIdx < idx {
				ip = ip[:idx]
			}
		} else {
			ip = ip[:idx]
		}
	}
	return ip
}

func (s *Server) writeError(w http.ResponseWriter, code, message, fix string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error:   code,
		Message: message,
		Fix:     fix,
	})
}

func computeExpiresIn(req MintRequest, token string) int {
	if claims, err := sign.UnverifiedClaims(token); err == nil {
		now := time.Now().Unix()
		if claims.Exp > now {
			return int(claims.Exp - now)
		}
	}

	ttl := req.TTL
	if ttl <= 0 {
		ttl = sign.DefaultTTL
	}
	if ttl > sign.MaxTTL {
		ttl = sign.MaxTTL
	}
	return ttl
}
