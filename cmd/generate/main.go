package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Println("Error generating:", err)
		os.Exit(1)
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		fmt.Println("Error marshalling:", err)
		os.Exit(1)
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	pemBytes := pem.EncodeToMemory(block)
	fmt.Println(string(pemBytes))

	os.WriteFile("private.pem", pemBytes, 0600)
	fmt.Println("Written to private.pem")
}
