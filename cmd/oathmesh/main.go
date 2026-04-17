package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/oathmesh/oathmesh/internal/audit"
	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/gateway"
	"github.com/oathmesh/oathmesh/internal/issuer"
	"github.com/oathmesh/oathmesh/internal/metrics"
	"github.com/oathmesh/oathmesh/internal/policy"
	"github.com/oathmesh/oathmesh/internal/sign"
	"github.com/oathmesh/oathmesh/internal/verify"
)

var (
	verbose    bool
	jsonOutput bool
	quiet      bool
)

// ValidSubjectSchemes defines the allowed URI schemes for --sub.
// OathMesh subjects MUST use one of these standardized schemes.
var ValidSubjectSchemes = []string{
	"svc://",   // services and microservices
	"agent://", // AI agents and bots
	"job://",   // CI/CD jobs
	"tool://",  // MCP-adjacent tool clients
	"user://",  // human delegation context only
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "oathmesh",
		Short: "OathMesh CLI — Machine call identity protocol",
		Long: `OathMesh gives every machine call a short-lived signed identity.

OathMesh authenticates the caller. The receiver authorizes the request.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging via slog")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Machine-readable JSON output for all commands")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress informational output (errors to stderr only)")

	rootCmd.AddCommand(buildMintCmd())
	rootCmd.AddCommand(buildVerifyCmd())
	rootCmd.AddCommand(buildInspectCmd())
	rootCmd.AddCommand(buildServeCmd())
	rootCmd.AddCommand(buildKeysCmd())
	rootCmd.AddCommand(buildPolicyCmd())

	if err := rootCmd.Execute(); err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
}

// ── oathmesh mint ───────────────────────────────────────────────────────────

func buildMintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "Mint an Oath Token",
		Long: `Mint a signed Oath Token for machine-to-machine authentication.

The token is output on stdout (pipeable). Use --inspect to decode and
pretty-print the minted token with an UNVERIFIED warning.

Examples:
  oathmesh mint --sub "agent://repo/acme/deploy-bot" --aud "https://api.internal" --act "inventory.write"
  oathmesh mint --sub "job://github/acme/deploy" --aud "https://api.internal" --act "deploy" --ttl 60
  oathmesh mint --sub "agent://repo/acme/bot" --aud "https://api.internal" --act "read" | oathmesh verify --audience "https://api.internal"

Exit codes:
  0 = success
  1 = signing failure
  2 = config error (missing key, invalid flags)`,
		RunE: mintRunE,
	}
	cmd.Flags().String("sub", "", "Subject URI (required; must use svc://, agent://, job://, tool://, or user:// scheme)")
	cmd.Flags().String("aud", "", "Audience URL (required)")
	cmd.Flags().String("act", "", "Action string (required)")
	cmd.Flags().Int("ttl", 120, "TTL hint in seconds (clamped server-side to max 300)")
	cmd.Flags().StringSlice("scope", []string{}, "Scope values (repeatable)")
	cmd.Flags().String("reason", "", "Reason claim")
	cmd.Flags().String("env", "", "Environment label")
	cmd.Flags().String("rqh", "", "Request hash binding (sha256:... format)")
	cmd.Flags().Bool("inspect", false, "Decode and pretty-print the minted token (with UNVERIFIED warning)")
	_ = cmd.MarkFlagRequired("sub")
	_ = cmd.MarkFlagRequired("aud")
	_ = cmd.MarkFlagRequired("act")
	return cmd
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
	inspect, _ := cmd.Flags().GetBool("inspect")

	// Validate subject URI scheme
	if err := validateSubjectURI(sub); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	ks, err := sign.LoadKeySet()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load keyset: %v\n", err)
		os.Exit(2)
		return nil
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
	metrics.TokensMintedTotal.Add(1)

	if jsonOutput {
		output := map[string]string{"token": token}
		b, _ := json.Marshal(output)
		fmt.Println(string(b))
	} else {
		fmt.Println(token)
	}

	// If --inspect: also decode and print (after outputting the raw token)
	if inspect {
		fmt.Fprintln(os.Stderr)
		printInspection(token)
	}

	return nil
}

// ── oathmesh verify ─────────────────────────────────────────────────────────

func buildVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [token]",
		Short: "Verify an Oath Token",
		Long: `Verify an Oath Token using the full 14-step verification pipeline.

The token can be provided as a positional argument, via --token flag,
or read from stdin (pipeable from oathmesh mint).

Output: VerifiedCallerContext as JSON on stdout (with --json) or
human-readable summary.

Examples:
  oathmesh verify <token> --audience "https://api.internal" --issuer "https://issuer.oathmesh.dev"
  echo <token> | oathmesh verify --audience "https://api.internal"
  oathmesh mint --sub "..." --aud "..." --act "..." | oathmesh verify --audience "..." --local-keys

Exit codes:
  0 = valid token
  1 = auth failure (signature, expiry, policy, replay, etc.)
  2 = config error (missing flags, bad keyset)`,
		RunE: verifyRunE,
	}
	cmd.Flags().String("token", "", "Token string (or provide as positional arg, or read from stdin)")
	cmd.Flags().String("audience", "", "Receiver audience URL (required)")
	cmd.Flags().StringSlice("issuer", []string{}, "Trusted issuer URLs (repeatable)")
	cmd.Flags().Bool("local-keys", false, "Use local keyset instead of fetching JWKS from issuer URL (dev only)")
	cmd.MarkFlagRequired("audience")
	return cmd
}

func verifyRunE(cmd *cobra.Command, args []string) error {
	tokenFlag, _ := cmd.Flags().GetString("token")
	audience, _ := cmd.Flags().GetString("audience")
	issuers, _ := cmd.Flags().GetStringSlice("issuer")
	localKeys, _ := cmd.Flags().GetBool("local-keys")

	token := readToken(tokenFlag, args)
	if token == "" {
		fmt.Fprintln(os.Stderr, "error: token required (provide as arg, --token flag, or pipe to stdin)")
		os.Exit(2)
	}

	// Build JWKS provider: default fetches from issuer URL,
	// --local-keys uses the local keyset (dev/testing only).
	var jwksProvider verify.JWKSProvider
	if localKeys {
		ks, err := sign.LoadKeySet()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: load local keyset: %v\n", err)
			os.Exit(2)
		}
		jwksProvider = verify.NewStaticJWKSProvider(ks.GetAllPublicKeys())
		// If no issuers specified, trust the local issuer
		if len(issuers) == 0 {
			issuers = []string{ks.GetIssuer()}
		}
	} else {
		// Default: fetch JWKS from the iss claim URL
		jwksProvider = verify.NewJWKSCache(verify.DefaultJWKSCacheTTL, nil)
		// If no issuers specified, extract from token (unverified) for JWKS fetch,
		// but the verify pipeline will still check against trusted issuers.
		if len(issuers) == 0 {
			claims, err := sign.UnverifiedClaims(token)
			if err == nil && claims.Iss != "" {
				issuers = []string{claims.Iss}
			}
		}
	}

	cfg := &verify.VerifierConfig{
		Audience:       audience,
		TrustedIssuers: issuers,
		JWKSProvider:   jwksProvider,
		ReplayCache:    verify.NewMemoryReplayCache(),
	}

	ctx := cmd.Context()
	vcc, err := verify.Verify(ctx, token, cfg)
	if err != nil {
		if jsonOutput {
			b, _ := json.Marshal(err)
			fmt.Fprintln(os.Stderr, string(b))
		} else if !quiet {
			fmt.Fprintf(os.Stderr, "verification failed: %v\n", err)
			if !localKeys && (strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host")) {
				fmt.Fprintln(os.Stderr, "Hint: if you are testing locally without a running issuer, use the --local-keys flag.")
			}
		}
		os.Exit(1)
	}

	if jsonOutput {
		b, _ := json.MarshalIndent(vcc, "", "  ")
		fmt.Println(string(b))
	} else if !quiet {
		fmt.Println("✓ Token verified successfully")
		fmt.Printf("  Issuer:  %s\n", vcc.Principal.Issuer)
		fmt.Printf("  Subject: %s\n", vcc.Principal.Subject)
		fmt.Printf("  Action:  %s\n", vcc.Action)
		fmt.Printf("  TokenID: %s\n", vcc.TokenID)
		fmt.Printf("  Expires: %s\n", vcc.ExpiresAt.Format(time.RFC3339))
		if len(vcc.Scope) > 0 {
			fmt.Printf("  Scope:   %s\n", strings.Join(vcc.Scope, ", "))
		}
		if vcc.Env != "" {
			fmt.Printf("  Env:     %s\n", vcc.Env)
		}
		if vcc.Source != nil {
			fmt.Printf("  Source:  %s/%s/%s\n", vcc.Source.Type, vcc.Source.Repo, vcc.Source.Workflow)
		}
	}

	return nil
}

// ── oathmesh inspect ────────────────────────────────────────────────────────

func buildInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect [token]",
		Short: "Inspect an Oath Token (unverified)",
		Long: `Decode and display the header and claims of an Oath Token WITHOUT verification.

⚠  The output is UNVERIFIED — do not trust for authorization decisions.

Shows: header fields, all claims, and an expiry countdown.

Examples:
  oathmesh inspect <token>
  oathmesh mint --sub "..." --aud "..." --act "..." | oathmesh inspect
  oathmesh inspect --json <token> | jq '.claims.sub'

Exit codes:
  0 = successfully decoded
  1 = parse error`,
		RunE: inspectRunE,
	}
	cmd.Flags().String("token", "", "Token string (or provide as positional arg, or read from stdin)")
	return cmd
}

func inspectRunE(cmd *cobra.Command, args []string) error {
	tokenFlag, _ := cmd.Flags().GetString("token")
	token := readToken(tokenFlag, args)
	if token == "" {
		return fmt.Errorf("token required (provide as arg, --token flag, or pipe to stdin)")
	}

	// Always print the UNVERIFIED warning — spec requirement
	fmt.Fprintln(os.Stderr, "⚠ UNVERIFIED — do not trust for authorization decisions")

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
			"warning": "UNVERIFIED — do not trust for authorization decisions",
			"header":  header,
			"claims":  claims,
		}
		b, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(b))
	} else {
		printInspection(token)
	}

	return nil
}

// ── oathmesh serve ──────────────────────────────────────────────────────────

func buildServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the OathMesh issuer server",
		Long: `Start the OathMesh issuer HTTP server.

The server listens on the configured port (default: 4000) and serves:
  POST /v1/token             — mint endpoint
  POST /v1/exchange/github   — GitHub OIDC exchange
  GET  /.well-known/jwks.json          — public keys
  GET  /.well-known/oathmesh-issuer    — discovery
  GET  /healthz              — liveness check

Examples:
  oathmesh serve
  oathmesh serve --port 8080
  oathmesh serve --config issuer.pkl

Exit codes:
  0 = clean shutdown
  1 = startup error`,
		RunE: serveRunE,
	}
	cmd.Flags().String("port", "4000", "Listen port")
	cmd.Flags().String("config", "", "Pkl config file path")
	cmd.Flags().Bool("gateway", false, "Enable reverse proxy / gateway mode")
	return cmd
}

func serveRunE(cmd *cobra.Command, args []string) error {
	port, _ := cmd.Flags().GetString("port")
	gatewayEnabled, _ := cmd.Flags().GetBool("gateway")

	if verbose {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	ks, err := sign.LoadKeySet()
	if err != nil {
		return fmt.Errorf("load keyset: %w", err)
	}

	srv := issuer.NewServer(ks)

	if gatewayEnabled {
		upstream := os.Getenv("OATHMESH_GATEWAY_UPSTREAM")
		audience := os.Getenv("OATHMESH_GATEWAY_AUDIENCE")
		issuersStr := os.Getenv("OATHMESH_GATEWAY_ISSUERS")
		policyFile := os.Getenv("OATHMESH_GATEWAY_POLICY")

		if upstream == "" || audience == "" || issuersStr == "" {
			return fmt.Errorf("gateway mode requires OATHMESH_GATEWAY_UPSTREAM, OATHMESH_GATEWAY_AUDIENCE, OATHMESH_GATEWAY_ISSUERS env vars")
		}

		issuers := strings.Split(issuersStr, ",")
		for i, v := range issuers {
			issuers[i] = strings.TrimSpace(v)
		}

		var evaluator verify.PolicyEvaluator
		if policyFile != "" {
			env := os.Getenv("OATHMESH_ENV")
			if env == "development" {
				pe, err := policy.NewWatchedPolicyEngine(policyFile, slog.Default())
				if err != nil {
					return fmt.Errorf("failed to init policy engine: %w", err)
				}
				evaluator = pe
			} else {
				p, err := policy.LoadPolicyFromFile(policyFile)
				if err != nil {
					return fmt.Errorf("failed to load policy engine: %w", err)
				}
				evaluator = policy.NewPolicyEngine(p)
			}
		}

		var gatewayAuditSink core.AuditSink = audit.NewStdoutAuditSink()
		if hmacKey := os.Getenv("OATHMESH_AUDIT_HMAC_KEY"); hmacKey != "" {
			gatewayAuditSink = audit.NewHMACChainAuditSink(gatewayAuditSink, []byte(hmacKey))
		}

		vCfg := &verify.VerifierConfig{
			Audience:        audience,
			TrustedIssuers:  issuers,
			JWKSProvider:    verify.NewJWKSCache(verify.DefaultJWKSCacheTTL, nil),
			ReplayCache:     verify.NewMemoryReplayCache(),
			PolicyEvaluator: evaluator,
			AuditSink:       gatewayAuditSink,
		}

		gwCfg := gateway.Config{
			UpstreamURL:  upstream,
			VerifyConfig: vCfg,
		}

		gwHandler, err := gateway.NewProxy(gwCfg)
		if err != nil {
			return fmt.Errorf("failed to initialize gateway proxy: %w", err)
		}

		srv.SetGateway(gwHandler)
	}

	if !quiet {
		log.Printf("OathMesh issuer starting on port %s\n", port)
		log.Printf("  JWKS:      http://localhost:%s/.well-known/jwks.json\n", port)
		log.Printf("  Discovery: http://localhost:%s/.well-known/oathmesh-issuer\n", port)
		log.Printf("  Mint:      http://localhost:%s/v1/token\n", port)
		log.Printf("  Health:    http://localhost:%s/healthz\n", port)
		if gatewayEnabled {
			log.Printf("  Gateway:   active (catch-all route /*)\n")
		}
	}

	return srv.Run()
}

// ── oathmesh keys ───────────────────────────────────────────────────────────

func buildKeysCmd() *cobra.Command {
	keysCmd := &cobra.Command{
		Use:   "keys",
		Short: "Key management commands",
	}

	rotateCmd := &cobra.Command{
		Use:   "rotate",
		Short: "Generate a new Ed25519 key pair and publish in JWKS",
		Long: `Generate a new Ed25519 key pair for the issuer.

The new key is published alongside the current key in JWKS during
the overlap period (default: 24 hours).

Examples:
  oathmesh keys rotate
  oathmesh keys rotate --json

Exit codes:
  0 = rotation successful
  1 = rotation error
  2 = keyset load error`,
		RunE: keysRotateRunE,
	}

	keysCmd.AddCommand(rotateCmd)
	return keysCmd
}

func keysRotateRunE(cmd *cobra.Command, args []string) error {
	ks, err := sign.LoadKeySet()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load keyset: %v\n", err)
		os.Exit(2)
		return nil
	}

	err = sign.RotateKey(ks)
	if err != nil {
		return fmt.Errorf("rotate key: %w", err)
	}

	if jsonOutput {
		output := map[string]string{"kid": ks.GetKid(), "status": "rotated"}
		b, _ := json.Marshal(output)
		fmt.Println(string(b))
	} else if !quiet {
		fmt.Printf("✓ Key rotated successfully. New kid: %s\n", ks.GetKid())
	}

	return nil
}

// ── oathmesh policy ─────────────────────────────────────────────────────────

func buildPolicyCmd() *cobra.Command {
	policyCmd := &cobra.Command{
		Use:   "policy",
		Short: "Policy management commands",
	}

	validateCmd := &cobra.Command{
		Use:   "validate <file>",
		Short: "Validate a policy file against the OathMesh schema",
		Long: `Validate a .pkl or .json policy file against the OathMesh policy schema.

Checks:
  - version == 1
  - at least one issuer
  - at least one audience
  - at least one rule
  - last rule is { name: "default", allow: false }

Examples:
  oathmesh policy validate policy/production.pkl
  oathmesh policy validate policy.json
  oathmesh policy validate --json policy.json

Exit codes:
  0 = valid policy
  1 = invalid policy (schema errors reported)`,
		Args: cobra.ExactArgs(1),
		RunE: policyValidateRunE,
	}

	policyCmd.AddCommand(validateCmd)
	return policyCmd
}

func policyValidateRunE(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	p, err := policy.LoadPolicyFromFile(filePath)
	if err != nil {
		if jsonOutput {
			output := map[string]interface{}{
				"valid":  false,
				"file":   filePath,
				"errors": []string{err.Error()},
			}
			b, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Fprintf(os.Stderr, "✗ Policy validation failed: %v\n", err)
		}
		os.Exit(1)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"valid":     true,
			"file":      filePath,
			"version":   p.Version,
			"issuers":   len(p.Issuers),
			"audiences": len(p.Audiences),
			"rules":     len(p.Rules),
		}
		b, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(b))
	} else if !quiet {
		fmt.Printf("✓ Policy valid: %s\n", filePath)
		fmt.Printf("  Version:   %d\n", p.Version)
		fmt.Printf("  Issuers:   %d\n", len(p.Issuers))
		fmt.Printf("  Audiences: %d\n", len(p.Audiences))
		fmt.Printf("  Rules:     %d\n", len(p.Rules))
		for _, r := range p.Rules {
			action := "deny"
			if r.Allow {
				action = "allow"
			}
			fmt.Printf("    → %s (%s)\n", r.Name, action)
		}
	}

	return nil
}

// ── Helpers ─────────────────────────────────────────────────────────────────

// readToken reads a token from flag, positional arg, or stdin.
func readToken(tokenFlag string, args []string) string {
	if tokenFlag != "" {
		return strings.TrimSpace(tokenFlag)
	}
	if len(args) > 0 {
		return strings.TrimSpace(args[0])
	}

	// Read from stdin (supports piping)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return strings.TrimSpace(scanner.Text())
		}
	}

	return ""
}

// validateSubjectURI checks that a subject string starts with a known URI scheme.
func validateSubjectURI(sub string) error {
	for _, scheme := range ValidSubjectSchemes {
		if strings.HasPrefix(sub, scheme) {
			return nil
		}
	}
	return fmt.Errorf(
		"invalid subject URI %q: must start with one of: svc://, agent://, job://, tool://, user://",
		sub,
	)
}

// printInspection decodes and pretty-prints a token with expiry countdown.
func printInspection(token string) {
	header, err := sign.ParseHeader(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  parse header failed: %v\n", err)
		return
	}

	claims, err := sign.UnverifiedClaims(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  parse claims failed: %v\n", err)
		return
	}

	fmt.Fprintln(os.Stderr, "⚠ UNVERIFIED — do not trust for authorization decisions")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Header:\n")
	fmt.Fprintf(os.Stderr, "  typ: %s\n", header.Typ)
	fmt.Fprintf(os.Stderr, "  alg: %s\n", header.Alg)
	fmt.Fprintf(os.Stderr, "  kid: %s\n", header.Kid)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Claims:\n")
	fmt.Fprintf(os.Stderr, "  iss: %s\n", claims.Iss)
	fmt.Fprintf(os.Stderr, "  sub: %s\n", claims.Sub)
	fmt.Fprintf(os.Stderr, "  aud: %s\n", claims.Aud)
	fmt.Fprintf(os.Stderr, "  act: %s\n", claims.Act)
	fmt.Fprintf(os.Stderr, "  iat: %d (%s)\n", claims.Iat, time.Unix(claims.Iat, 0).Format(time.RFC3339))
	fmt.Fprintf(os.Stderr, "  exp: %d (%s)\n", claims.Exp, time.Unix(claims.Exp, 0).Format(time.RFC3339))
	fmt.Fprintf(os.Stderr, "  jti: %s\n", claims.JTI)

	if len(claims.Scope) > 0 {
		fmt.Fprintf(os.Stderr, "  scope: %s\n", strings.Join(claims.Scope, ", "))
	}
	if claims.Reason != "" {
		fmt.Fprintf(os.Stderr, "  reason: %s\n", claims.Reason)
	}
	if claims.Env != "" {
		fmt.Fprintf(os.Stderr, "  env: %s\n", claims.Env)
	}
	if claims.RQH != "" {
		fmt.Fprintf(os.Stderr, "  rqh: %s\n", claims.RQH)
	}
	if claims.Src != nil {
		fmt.Fprintf(os.Stderr, "  src:\n")
		fmt.Fprintf(os.Stderr, "    type:     %s\n", claims.Src.Type)
		fmt.Fprintf(os.Stderr, "    repo:     %s\n", claims.Src.Repo)
		fmt.Fprintf(os.Stderr, "    workflow: %s\n", claims.Src.Workflow)
		if claims.Src.RunID != "" {
			fmt.Fprintf(os.Stderr, "    run_id:   %s\n", claims.Src.RunID)
		}
		if claims.Src.SHA != "" {
			fmt.Fprintf(os.Stderr, "    sha:      %s\n", claims.Src.SHA)
		}
	}

	// Expiry countdown
	fmt.Fprintln(os.Stderr)
	expTime := time.Unix(claims.Exp, 0)
	remaining := time.Until(expTime)
	if remaining > 0 {
		fmt.Fprintf(os.Stderr, "  ⏱ Expires in: %s\n", remaining.Round(time.Second))
	} else {
		fmt.Fprintf(os.Stderr, "  ✗ EXPIRED %s ago\n", (-remaining).Round(time.Second))
	}
}

// init sets up verbose logging if --verbose is specified.
func init() {
	// Default to info level; --verbose switches to debug in serveRunE
	_ = context.Background // ensure import
}
