package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

// Config holds the runtime configuration for the OathMesh issuer.
type Config struct {
	Issuer         string
	Port           int
	PrivateKey     string
	PrivateKeyFile string
	ConfigFile     string
	Env            string

	// TTL
	TTLDefault int
	TTLMax     int
	TTLWrite   int

	// Rate Limiting
	RateLimitRPM   int
	RateLimitBurst int

	// Audit
	AuditSink    string
	AuditFile    string
	AuditHMACKey string

	// JWKS
	JWKSCacheTTL int

	// Redis
	RedisURL string

	// Database
	DatabaseURL string

	// Logging
	LogLevel string

	// Gateway
	GatewayUpstream string
	GatewayAudience string
	GatewayIssuers  string
	GatewayPolicy   string
}

// LoadFromEnv populates a Config from environment variables.
// Values not set in the environment use defaults matching the canonical .env.example.
func LoadFromEnv() *Config {
	return &Config{
		Issuer:         getEnv("OATHMESH_ISSUER", "http://localhost:4000"),
		Port:           getEnvInt("OATHMESH_PORT", 4000),
		PrivateKey:     os.Getenv("OATHMESH_PRIVATE_KEY"),
		PrivateKeyFile: os.Getenv("OATHMESH_PRIVATE_KEY_FILE"),
		ConfigFile:     os.Getenv("OATHMESH_CONFIG_FILE"),
		Env:            getEnv("OATHMESH_ENV", "production"),

		TTLDefault: getEnvInt("OATHMESH_TTL_DEFAULT", 120),
		TTLMax:     getEnvInt("OATHMESH_TTL_MAX", 300),
		TTLWrite:   getEnvInt("OATHMESH_TTL_WRITE", 60),

		RateLimitRPM:   getEnvInt("OATHMESH_RATE_LIMIT_RPM", 100),
		RateLimitBurst: getEnvInt("OATHMESH_RATE_LIMIT_BURST", 20),

		AuditSink:    getEnv("OATHMESH_AUDIT_SINK", "stdout"),
		AuditFile:    os.Getenv("OATHMESH_AUDIT_FILE"),
		AuditHMACKey: os.Getenv("OATHMESH_AUDIT_HMAC_KEY"),

		JWKSCacheTTL: getEnvInt("OATHMESH_JWKS_CACHE_TTL", 60),

		RedisURL:    os.Getenv("REDIS_URL"),
		DatabaseURL: os.Getenv("DATABASE_URL"),

		LogLevel: getEnv("OATHMESH_LOG_LEVEL", "info"),

		GatewayUpstream: os.Getenv("OATHMESH_GATEWAY_UPSTREAM"),
		GatewayAudience: os.Getenv("OATHMESH_GATEWAY_AUDIENCE"),
		GatewayIssuers:  os.Getenv("OATHMESH_GATEWAY_ISSUERS"),
		GatewayPolicy:   os.Getenv("OATHMESH_GATEWAY_POLICY"),
	}
}

// Validate returns an error if required configuration is missing.
func (c *Config) Validate() error {
	if c.Issuer == "" {
		return fmt.Errorf("OATHMESH_ISSUER is required")
	}

	u, err := url.Parse(c.Issuer)
	if err != nil {
		return fmt.Errorf("invalid OATHMESH_ISSUER: %w", err)
	}
	if c.Env != "development" && u.Scheme != "https" {
		return fmt.Errorf("OATHMESH_ISSUER must use HTTPS in non-development environments (got %q). Set OATHMESH_ENV=development to suppress this check", c.Issuer)
	}

	if c.PrivateKey == "" && c.PrivateKeyFile == "" {
		return fmt.Errorf("OATHMESH_PRIVATE_KEY or OATHMESH_PRIVATE_KEY_FILE is required")
	}
	if c.TTLMax < 1 || c.TTLMax > 300 {
		return fmt.Errorf("OATHMESH_TTL_MAX must be between 1 and 300")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
