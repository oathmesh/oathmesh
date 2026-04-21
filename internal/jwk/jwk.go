// Package jwk provides JWK (JSON Web Key) and JWKS (JSON Web Key Set)
// types and operations for the OathMesh protocol.
//
// This package has zero external dependencies beyond the Go standard library,
// following the OathMesh core isolation principle.
package jwk

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// JWKS represents a JSON Web Key Set containing one or more keys.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a single JSON Web Key (RFC 7517).
// OathMesh only supports OKP keys with Ed25519 curves.
type JWK struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Crv string `json:"crv"`
	X   string `json:"x"`
}

// BuildJWKS constructs a JWKS from a map of kid→public key.
// All keys are encoded as OKP/Ed25519/EdDSA.
func BuildJWKS(keys map[string]ed25519.PublicKey) (*JWKS, error) {
	jwks := JWKS{
		Keys: make([]JWK, 0, len(keys)),
	}

	for kid, pubKey := range keys {
		x := make([]byte, base64.RawURLEncoding.EncodedLen(len(pubKey)))
		base64.RawURLEncoding.Encode(x, pubKey)

		jwks.Keys = append(jwks.Keys, JWK{
			Kty: "OKP",
			Alg: "EdDSA",
			Kid: kid,
			Use: "sig",
			Crv: "Ed25519",
			X:   string(x),
		})
	}

	return &jwks, nil
}

// ToJSON marshals the JWKS to JSON bytes.
func (j *JWKS) ToJSON() ([]byte, error) {
	return json.Marshal(j)
}

// ParseJWKS parses a JWKS from JSON bytes.
func ParseJWKS(data []byte) (*JWKS, error) {
	var jwks JWKS
	err := json.Unmarshal(data, &jwks)
	return &jwks, err
}

// GetKeyFromJWKS extracts an Ed25519 public key from a JWKS by key ID.
// Returns an error if the key is not found or is not an Ed25519 key.
func GetKeyFromJWKS(jwks *JWKS, kid string) (ed25519.PublicKey, error) {
	for _, key := range jwks.Keys {
		if key.Kid == kid {
			if key.Kty != "OKP" || key.Crv != "Ed25519" {
				return nil, fmt.Errorf("key is not Ed25519")
			}

			x, err := base64.RawURLEncoding.DecodeString(key.X)
			if err != nil {
				return nil, fmt.Errorf("decode x: %w", err)
			}

			pubKey := make(ed25519.PublicKey, len(x))
			copy(pubKey, x)
			return pubKey, nil
		}
	}

	return nil, fmt.Errorf("key not found: %s", kid)
}
