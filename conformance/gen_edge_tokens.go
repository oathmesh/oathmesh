package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/oathmesh/oathmesh/internal/sign"
)

func main() {
	keyBytes, err := os.ReadFile("test-key.pem")
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(keyBytes)
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	privateKey := privKey.(ed25519.PrivateKey)

	now := time.Now().Unix()

	// 29s expired (PASS)
	t29 := mintWithExp(privateKey, now-29)
	// 30s expired (PASS)
	t30 := mintWithExp(privateKey, now-30)
	// 31s expired (FAIL)
	t31 := mintWithExp(privateKey, now-31)

	fmt.Printf("%s\n%s\n%s\n", t29, t30, t31)
}

func mintWithExp(priv ed25519.PrivateKey, exp int64) string {
	claims := sign.Claims{
		Iss: "http://localhost:4000",
		Sub: "agent://test/svc",
		Aud: "https://inventory.internal",
		Act: "read",
		Iat: exp - 120,
		Exp: exp,
		JTI: uuid.New().String(),
	}
	header := sign.Header{Typ: sign.TypeHeader, Alg: sign.AlgEdDSA, Kid: "issuer-key-2026-04"}
	t, _ := sign.BuildJWS(header, claims, priv)
	return t
}
