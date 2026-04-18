package sign

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	TypeHeader = "om+jwt"
	AlgEdDSA   = "EdDSA"
	KidFormat  = "issuer-key-2006-01"
	DefaultTTL = 120
	MaxTTL     = 300
)

type Token struct {
	Header    Header
	Claims    Claims
	Signature []byte
}

type Header struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type Claims struct {
	Iss         string   `json:"iss"`
	Sub         string   `json:"sub"`
	Aud         string   `json:"aud"`
	Act         string   `json:"act"`
	Iat         int64    `json:"iat"`
	Nbf         int64    `json:"nbf,omitempty"`
	Exp         int64    `json:"exp"`
	JTI         string   `json:"jti"`
	Scope       []string `json:"scope,omitempty"`
	Reason      string   `json:"reason,omitempty"`
	Src         *Source  `json:"src,omitempty"`
	DelegatedBy string   `json:"delegated_by,omitempty"`
	Env         string   `json:"env,omitempty"`
	Tenant      string   `json:"tenant,omitempty"`
	RQH         string   `json:"rqh,omitempty"`
}

type Source struct {
	Type     string `json:"type"`
	Repo     string `json:"repo"`
	Workflow string `json:"workflow"`
	RunID    string `json:"run_id,omitempty"`
	SHA      string `json:"sha,omitempty"`
}

type MintRequest struct {
	Sub    string   `json:"sub"`
	Aud    string   `json:"aud"`
	Act    string   `json:"act"`
	TTL    int      `json:"ttl_hint,omitempty"`
	Nbf    int      `json:"nbf_hint,omitempty"`
	Scope  []string `json:"scope,omitempty"`
	Reason string   `json:"reason,omitempty"`
	Env    string   `json:"env,omitempty"`
	Tenant string   `json:"tenant,omitempty"`
	RQH    string   `json:"rqh,omitempty"`
	Src    *Source  `json:"src,omitempty"`
}

func BuildJWS(header Header, claims Claims, privateKey ed25519.PrivateKey) (string, error) {
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64

	signature := ed25519.Sign(privateKey, []byte(signingInput))

	token := signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)

	return token, nil
}

func SignToken(req MintRequest, issuer string, privateKey ed25519.PrivateKey, kid string) (string, error) {
	now := time.Now().Unix()

	ttl := req.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	if ttl > MaxTTL {
		ttl = MaxTTL
	}

	nbf := now
	if req.Nbf > 0 {
		nbf = now + int64(req.Nbf)
	}

	claims := Claims{
		Iss:    issuer,
		Sub:    req.Sub,
		Aud:    req.Aud,
		Act:    req.Act,
		Iat:    now,
		Nbf:    nbf,
		Exp:    now + int64(ttl),
		JTI:    uuid.New().String(),
		Scope:  req.Scope,
		Reason: req.Reason,
		Env:    req.Env,
		Tenant: req.Tenant,
		RQH:    req.RQH,
		Src:    req.Src,
	}

	header := Header{
		Typ: TypeHeader,
		Alg: AlgEdDSA,
		Kid: kid,
	}

	return BuildJWS(header, claims, privateKey)
}

func GenerateKeyPair() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}
	return privateKey, publicKey, nil
}

func ParsePrivateKey(pemData []byte) (ed25519.PrivateKey, error) {
	block, err := x509.ParsePKCS8PrivateKey(pemData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
	}

	key, ok := block.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not Ed25519")
	}

	return key, nil
}

func ParsePrivateKeyPEM(pemString string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	return ParsePrivateKey(block.Bytes)
}

func MarshalPrivateKeyToPEM(key ed25519.PrivateKey) ([]byte, error) {
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	return pem.EncodeToMemory(block), nil
}

type PublicKey = ed25519.PublicKey

func PublicKeyFromPrivate(privateKey ed25519.PrivateKey) []byte {
	return privateKey.Public().(ed25519.PublicKey)
}

func MarshalJWK(publicKey ed25519.PublicKey, kid string) (map[string]interface{}, error) {
	pkBytes := publicKey

	x := make([]byte, base64.RawURLEncoding.EncodedLen(len(pkBytes)))
	base64.RawURLEncoding.Encode(x, pkBytes)

	jwk := map[string]interface{}{
		"kty": "OKP",
		"alg": "EdDSA",
		"kid": kid,
		"use": "sig",
		"crv": "Ed25519",
		"x":   string(x),
	}

	return jwk, nil
}

func SignJWSCompact(header, claims []byte, privateKey ed25519.PrivateKey) (string, error) {
	headerB64 := base64.RawURLEncoding.EncodeToString(header)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claims)

	signingInput := headerB64 + "." + claimsB64

	signature := ed25519.Sign(privateKey, []byte(signingInput))

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func ParseJWSCompact(token string) ([]byte, []byte, []byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, nil, fmt.Errorf("invalid JWS format: expected 3 parts, got %d", len(parts))
	}

	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("decode header: %w", err)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("decode payload: %w", err)
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("decode signature: %w", err)
	}

	return header, payload, signature, nil
}

func VerifyJWSCompact(token string, publicKey ed25519.PublicKey) ([]byte, []byte, error) {
	header, payload, signature, err := ParseJWSCompact(token)
	if err != nil {
		return nil, nil, err
	}

	signingInput := strings.Join([]string{
		base64.RawURLEncoding.EncodeToString(header),
		base64.RawURLEncoding.EncodeToString(payload),
	}, ".")

	if !ed25519.Verify(publicKey, []byte(signingInput), signature) {
		return nil, nil, fmt.Errorf("signature verification failed")
	}

	return header, payload, nil
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Crv string `json:"crv"`
	X   string `json:"x"`
}

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

func (j *JWKS) ToJSON() ([]byte, error) {
	return json.Marshal(j)
}

func ParseJWKS(data []byte) (*JWKS, error) {
	var jwks JWKS
	err := json.Unmarshal(data, &jwks)
	return &jwks, err
}

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

type HeaderPayload struct {
	Header Header
	Claims Claims
}

func ParseHeader(token string) (Header, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Header{}, fmt.Errorf("invalid token format")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Header{}, fmt.Errorf("decode header: %w", err)
	}

	var header Header
	err = json.Unmarshal(headerJSON, &header)
	if err != nil {
		return Header{}, fmt.Errorf("parse header: %w", err)
	}

	return header, nil
}

func UnverifiedClaims(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode claims: %w", err)
	}

	var claims Claims
	err = json.Unmarshal(claimsJSON, &claims)
	if err != nil {
		return nil, fmt.Errorf("parse claims: %w", err)
	}

	return &claims, nil
}

func MarshalJSON(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

var _ crypto.Signer = (*ed25519PrivateKeyWrapper)(nil)

type ed25519PrivateKeyWrapper struct {
	key ed25519.PrivateKey
}

func (w *ed25519PrivateKeyWrapper) Public() crypto.PublicKey {
	return w.key.Public()
}

func (w *ed25519PrivateKeyWrapper) Sign(_ io.Reader, msg []byte, _ crypto.SignerOpts) ([]byte, error) {
	return ed25519.Sign(w.key, msg), nil
}
