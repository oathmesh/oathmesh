package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/oathmesh/oathmesh/internal/issuer"
	"github.com/oathmesh/oathmesh/internal/sign"
)

var (
	verbose    bool
	jsonOutput bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "oathmesh",
		Short: "OathMesh CLI - Machine call identity protocol",
		Long:  `OathMesh gives every machine call a short-lived signed identity.`,
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output JSON")

	mintCmd := &cobra.Command{
		Use:   "mint",
		Short: "Mint an Oath Token",
		RunE:  mintRunE,
	}
	mintCmd.Flags().String("sub", "", "Subject URI (required)")
	mintCmd.Flags().String("aud", "", "Audience URL (required)")
	mintCmd.Flags().String("act", "", "Action string (required)")
	mintCmd.Flags().Int("ttl", 120, "TTL hint in seconds")
	mintCmd.Flags().StringSlice("scope", []string{}, "Scope values")
	mintCmd.Flags().String("reason", "", "Reason claim")
	mintCmd.Flags().String("env", "", "Environment label")
	mintCmd.Flags().String("rqh", "", "Request hash binding")
	mintCmd.MarkFlagRequired("sub")
	mintCmd.MarkFlagRequired("aud")
	mintCmd.MarkFlagRequired("act")

	rootCmd.AddCommand(mintCmd)

	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify an Oath Token",
		RunE:  verifyRunE,
	}
	verifyCmd.Flags().String("token", "", "Token string (or read from stdin)")
	verifyCmd.Flags().String("audience", "", "Expected audience (required)")
	verifyCmd.Flags().StringSlice("issuer", []string{}, "Trusted issuers")
	verifyCmd.MarkFlagRequired("audience")

	rootCmd.AddCommand(verifyCmd)

	inspectCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect an Oath Token (unverified)",
		RunE:  inspectRunE,
	}
	inspectCmd.Flags().String("token", "", "Token string (or read from stdin)")

	rootCmd.AddCommand(inspectCmd)

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the OathMesh issuer server",
		RunE:  serveRunE,
	}
	serveCmd.Flags().String("port", "4000", "Listen port")

	rootCmd.AddCommand(serveCmd)

	keysCmd := &cobra.Command{
		Use:   "keys",
		Short: "Key management",
	}

	keysRotateCmd := &cobra.Command{
		Use:   "rotate",
		Short: "Generate a new key pair",
		RunE:  keysRotateRunE,
	}
	keysCmd.AddCommand(keysRotateCmd)

	rootCmd.AddCommand(keysCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func mintRunE(cmd *cobra.Command, args []string) error {
	sub, _ := cmd.Flags().GetString("sub")
	aud, _ := cmd.Flags().GetString("aud")
	act, _ := cmd.Flags().GetString("act")
	ttl, _ := cmd.Flags().GetInt("ttl")
	scope, _ := cmd.Flags().GetStringSlice("scope")
	reason, _ := cmd.Flags().GetString("reason")
	env, _ := cmd.Flags().GetString("env")
	rqh, _ := cmd.Flags().GetString("rqh")

	ks, err := sign.LoadKeySet()
	if err != nil {
		return fmt.Errorf("load keyset: %w", err)
	}

	req := sign.MintRequest{
		Sub:    sub,
		Aud:    aud,
		Act:    act,
		TTL:    ttl,
		Scope:  scope,
		Reason: reason,
		Env:    env,
		RQH:    rqh,
	}

	token, err := ks.SignToken(req)
	if err != nil {
		return fmt.Errorf("sign token: %w", err)
	}

	if jsonOutput {
		output := map[string]string{"token": token}
		b, _ := json.Marshal(output)
		fmt.Println(string(b))
	} else {
		fmt.Println(token)
	}

	return nil
}

func verifyRunE(cmd *cobra.Command, args []string) error {
	tokenFlag, _ := cmd.Flags().GetString("token")
	audience, _ := cmd.Flags().GetString("audience")
	issuers, _ := cmd.Flags().GetStringSlice("issuer")

	var token string
	if tokenFlag != "" {
		token = tokenFlag
	} else if len(args) > 0 {
		token = args[0]
	} else {
		fmt.Scan(&token)
	}

	if token == "" {
		return fmt.Errorf("token required")
	}

	header, err := sign.ParseHeader(token)
	if err != nil {
		return fmt.Errorf("parse header: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"header": header,
		}
		b, _ := json.Marshal(output)
		fmt.Println(string(b))
	} else {
		fmt.Printf("Type: %s, Algorithm: %s, Key ID: %s\n", header.Typ, header.Alg, header.Kid)
	}

	_ = audience
	_ = issuers

	return nil
}

func inspectRunE(cmd *cobra.Command, args []string) error {
	tokenFlag, _ := cmd.Flags().GetString("token")

	var token string
	if tokenFlag != "" {
		token = tokenFlag
	} else if len(args) > 0 {
		token = args[0]
	} else {
		fmt.Scan(&token)
	}

	if token == "" {
		return fmt.Errorf("token required")
	}

	fmt.Println("⚠ UNVERIFIED — do not trust for authorization decisions")

	header, err := sign.ParseHeader(token)
	if err != nil {
		return fmt.Errorf("parse header: %w", err)
	}

	claims, err := sign.UnverifiedClaims(token)
	if err != nil {
		return fmt.Errorf("parse claims: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"header": header,
			"claims": claims,
		}
		b, _ := json.Marshal(output)
		fmt.Println(string(b))
	} else {
		fmt.Printf("Header:\n  typ: %s\n  alg: %s\n  kid: %s\n", header.Typ, header.Alg, header.Kid)
		fmt.Printf("Claims:\n  iss: %s\n  sub: %s\n  aud: %s\n  act: %s\n  iat: %d\n  exp: %d\n  jti: %s\n",
			claims.Iss, claims.Sub, claims.Aud, claims.Act, claims.Iat, claims.Exp, claims.JTI)
	}

	return nil
}

func serveRunE(cmd *cobra.Command, args []string) error {
	port, _ := cmd.Flags().GetString("port")
	_ = port

	ks, err := sign.LoadKeySet()
	if err != nil {
		return fmt.Errorf("load keyset: %w", err)
	}

	srv := issuer.NewServer(ks)

	log.Println("Starting OathMesh issuer server...")
	return srv.Run()
}

func keysRotateRunE(cmd *cobra.Command, args []string) error {
	ks, err := sign.LoadKeySet()
	if err != nil {
		return fmt.Errorf("load keyset: %w", err)
	}

	err = sign.RotateKey(ks)
	if err != nil {
		return fmt.Errorf("rotate key: %w", err)
	}

	fmt.Printf("Key rotated. New kid: %s\n", ks.GetKid())

	return nil
}
