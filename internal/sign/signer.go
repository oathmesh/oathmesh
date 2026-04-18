package sign

// Signer formally bounds what the HTTP issuer and handlers require 
// for minting cryptographic tokens.
type Signer interface {
	GetIssuer() string
	JWKS() (*JWKS, error)
	SignToken(req MintRequest) (string, error)
}
