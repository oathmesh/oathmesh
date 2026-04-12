package issuer

import (
	"encoding/json"
	"net/http"

	"github.com/oathmesh/oathmesh/internal/sign"
)

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
	Token string `json:"token"`
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

	if r.Header.Get("Content-Type") != "application/json" {
		s.writeError(w, "invalid_content_type", "Content-Type must be application/json", "")
		return
	}

	var req MintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "invalid_request", "Failed to parse request body", err.Error())
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
		s.writeError(w, "mint_failed", "Failed to sign token", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MintResponse{Token: token})
}

func (s *Server) writeError(w http.ResponseWriter, code, message, fix string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   code,
		Message: message,
		Fix:     fix,
	})
}
