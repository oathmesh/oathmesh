package issuer

import (
	"encoding/json"
	"net/http"
)

func (s *Server) jwksHandler(w http.ResponseWriter, r *http.Request) {
	jwks, err := s.keySet.JWKS()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jwks)
}

type IssuerMetadata struct {
	Issuer                string   `json:"issuer"`
	JWKSURI               string   `json:"jwks_uri"`
	AlgValuesSupported    []string `json:"alg_values_supported"`
	MaxTTLSeconds         int      `json:"max_ttl_seconds"`
	TokenType             string   `json:"token_type"`
	BindingModesSupported []string `json:"binding_modes_supported"`
	Version               string   `json:"version"`
}

func (s *Server) discoveryHandler(w http.ResponseWriter, r *http.Request) {
	issuer := s.keySet.GetIssuer()

	metadata := IssuerMetadata{
		Issuer:                issuer,
		JWKSURI:               issuer + "/.well-known/jwks.json",
		AlgValuesSupported:    []string{"EdDSA", "ES256"},
		MaxTTLSeconds:         300,
		TokenType:             "om+jwt",
		BindingModesSupported: []string{"none", "request-hash"},
		Version:               "1.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metadata)
}
