package sign

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/google/uuid"
)

// KMSSigner heavily restricts access mapping exclusively to AWS KMS for
// payload signing preventing physical node extraction.
type KMSSigner struct {
	issuer string
	keyID  string
	client *kms.Client
	kid    string
	pubKey ed25519.PublicKey
}

// NewKMSSigner dynamically instantiates an AWS KMS Signer verifying the key integrity locally.
func NewKMSSigner(ctx context.Context, keyID string, issuer string) (*KMSSigner, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws load default config: %w", err)
	}

	client := kms.NewFromConfig(cfg)

	out, err := client.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: aws.String(keyID),
	})
	if err != nil {
		return nil, fmt.Errorf("aws kms get public key: %w", err)
	}

	if out.KeySpec != types.KeySpecEccNistEdwards25519 {
		return nil, fmt.Errorf("aws kms key %s is not ECC_ED25519, got %s", keyID, string(out.KeySpec))
	}

	parsedKey, err := x509.ParsePKIXPublicKey(out.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse spki string from kms public key: %w", err)
	}

	edPubKey, ok := parsedKey.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("parsed key is not ed25519 public key")
	}

	var kid string
	if len(keyID) > 8 {
		kid = "kms-" + keyID[len(keyID)-8:]
	} else {
		kid = "kms-" + keyID
	}

	return &KMSSigner{
		issuer: issuer,
		keyID:  keyID,
		client: client,
		kid:    kid,
		pubKey: edPubKey,
	}, nil
}

func (s *KMSSigner) GetIssuer() string {
	return s.issuer
}

func (s *KMSSigner) JWKS() (*JWKS, error) {
	keys := map[string]ed25519.PublicKey{
		s.kid: s.pubKey,
	}
	return BuildJWKS(keys)
}

func (s *KMSSigner) SignToken(req MintRequest) (string, error) {
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
		Iss:    s.issuer,
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
		Kid: s.kid,
	}

	// Use the shared JWS assembly path with KMS as the signing primitive.
	return BuildJWSWithSignFunc(header, claims, func(signingInput []byte) ([]byte, error) {
		out, err := s.client.Sign(context.TODO(), &kms.SignInput{
			KeyId:            aws.String(s.keyID),
			Message:          signingInput,
			MessageType:      types.MessageTypeRaw,
			SigningAlgorithm: types.SigningAlgorithmSpecEd25519Sha512,
		})
		if err != nil {
			return nil, fmt.Errorf("aws kms sign: %w", err)
		}
		return out.Signature, nil
	})
}
