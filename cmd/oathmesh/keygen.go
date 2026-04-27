package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/oathmesh/oathmesh/internal/sign"
)

func buildKeygenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate an Ed25519 key pair for OathMesh",
		Long: `Generate a new Ed25519 private and public key pair.

The private key is saved with strict permissions (0600) to the path specified by --out.
The public key is saved to <out>.pub.

SECURITY WARNING: Never commit private keys to version control.
For production deployments, it is highly recommended to use a Hardware Security Module (HSM)
or cloud Key Management Service (KMS) like AWS KMS, Google Cloud KMS, or Azure Key Vault
rather than local file-based keys.`,
		RunE: keygenRunE,
	}
	cmd.Flags().String("out", "private.pem", "Output path for the private key file")
	return cmd
}

func keygenRunE(cmd *cobra.Command, args []string) error {
	outPath, _ := cmd.Flags().GetString("out")
	pubPath := outPath + ".pub"

	privateKey, publicKey, err := sign.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Format private key as PKCS8 PEM
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("marshal private key: %w", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	// Save private key with strict 0600 permissions
	if err := os.WriteFile(outPath, privPEM, 0600); err != nil {
		return fmt.Errorf("write private key: %w", err)
	}

	// Format public key as PKIX PEM
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("marshal public key: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	// Save public key
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		return fmt.Errorf("write public key: %w", err)
	}

	if !quiet {
		fmt.Printf("✓ Generated new Ed25519 key pair\n")
		fmt.Printf("  Private key: %s (permissions 0600)\n", outPath)
		fmt.Printf("  Public key:  %s\n\n", pubPath)
		fmt.Println("SECURITY GUIDANCE:")
		fmt.Println("  - Never commit the private key to version control.")
		fmt.Println("  - Set OATHMESH_PRIVATE_KEY_PATH=" + outPath)
		fmt.Println("  - For production, use an HSM or Cloud KMS (e.g. AWS KMS).")
	}

	return nil
}
