package sign

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	EnvPrivateKey     = "OATHMESH_PRIVATE_KEY"
	EnvPrivateKeyB64  = "OATHMESH_PRIVATE_KEY_B64"
	EnvPrivateKeyFile = "OATHMESH_PRIVATE_KEY_FILE"
	EnvIssuer         = "OATHMESH_ISSUER"
	DefaultIssuer     = "https://issuer.oathmesh.tech"
	EnvJWKS_TTL       = "OATHMESH_JWKS_CACHE_TTL"
	DefaultJWKS_TTL   = 300
)

type KeySet struct {
	Current     ed25519.PrivateKey
	CurrentKid  string
	Previous    ed25519.PrivateKey
	PreviousKid string
	Issuer      string
	PublicKeys  map[string]ed25519.PublicKey
}

func LoadKeySet() (*KeySet, error) {
	privateKey, kid, err := loadPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}

	issuer := getEnv(EnvIssuer, DefaultIssuer)

	ks := &KeySet{
		Current:    privateKey,
		CurrentKid: kid,
		Issuer:     issuer,
		PublicKeys: make(map[string]ed25519.PublicKey),
	}

	ks.PublicKeys[kid] = PublicKeyFromPrivate(privateKey)

	return ks, nil
}

func loadPrivateKey() (ed25519.PrivateKey, string, error) {
	var pemData = os.Getenv(EnvPrivateKey)
	
	if b64Data := os.Getenv(EnvPrivateKeyB64); b64Data != "" {
		decoded, err := base64.StdEncoding.DecodeString(b64Data)
		if err != nil {
			return nil, "", fmt.Errorf("decode OATHMESH_PRIVATE_KEY_B64: %w", err)
		}
		pemData = string(decoded)
	}

	if pemData == "" {
		keyFile := os.Getenv(EnvPrivateKeyFile)
		if keyFile != "" {
			data, err := os.ReadFile(keyFile)
			if err != nil {
				return nil, "", fmt.Errorf("read key file: %w", err)
			}
			pemData = string(data)
		}
	}

	if pemData == "" {
		return nil, "", fmt.Errorf("no private key found: set %s or %s", EnvPrivateKey, EnvPrivateKeyFile)
	}

	privateKey, err := ParsePrivateKeyPEM(pemData)
	if err != nil {
		return nil, "", fmt.Errorf("parse private key: %w", err)
	}

	kid := generateKid()

	return privateKey, kid, nil
}

func generateKid() string {
	now := time.Now()
	randBytes := make([]byte, 2)
	_, _ = rand.Read(randBytes)
	randHex := fmt.Sprintf("%02x%02x", randBytes[0], randBytes[1])
	return fmt.Sprintf("issuer-key-%04d-%02d-%s", now.Year(), now.Month(), randHex)
}

func (ks *KeySet) GetKid() string {
	return ks.CurrentKid
}

func (ks *KeySet) GetIssuer() string {
	return ks.Issuer
}

func (ks *KeySet) GetPublicKey() ed25519.PublicKey {
	return PublicKeyFromPrivate(ks.Current)
}

func (ks *KeySet) GetAllPublicKeys() map[string]ed25519.PublicKey {
	return ks.PublicKeys
}

func (ks *KeySet) JWKS() (*JWKS, error) {
	return BuildJWKS(ks.PublicKeys)
}

func (ks *KeySet) SignToken(req MintRequest) (string, error) {
	return SignToken(req, ks.Issuer, ks.Current, ks.CurrentKid)
}

func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func getEnvInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return i
}

func ParsePrivateKeyFromFile(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return ParsePrivateKey(data)
}

func GenerateAndSaveKeyPair(path string) (ed25519.PrivateKey, error) {
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	pemBytes, err := MarshalPrivateKeyToPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}

	err = os.WriteFile(path, pemBytes, 0600)
	if err != nil {
		return nil, fmt.Errorf("write key file: %w", err)
	}

	return privateKey, nil
}

func RotateKey(ks *KeySet) error {
	newPrivateKey, newPublicKey, err := GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("generate new key: %w", err)
	}

	newKid := generateKid()

	ks.Previous = ks.Current
	ks.PreviousKid = ks.CurrentKid

	ks.Current = newPrivateKey
	ks.CurrentKid = newKid

	ks.PublicKeys = make(map[string]ed25519.PublicKey)
	ks.PublicKeys[newKid] = newPublicKey
	if ks.PreviousKid != "" {
		ks.PublicKeys[ks.PreviousKid] = PublicKeyFromPrivate(ks.Previous)
	}

	// Persist to disk if a key file is configured
	keyFile := os.Getenv(EnvPrivateKeyFile)
	if keyFile != "" {
		pemBytes, err := MarshalPrivateKeyToPEM(newPrivateKey)
		if err == nil {
			_ = os.WriteFile(keyFile, pemBytes, 0600)
		}
	}

	return nil
}

func (ks *KeySet) GetJWKS() int {
	return getEnvInt(EnvJWKS_TTL, DefaultJWKS_TTL)
}
