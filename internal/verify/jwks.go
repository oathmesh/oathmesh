package verify

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/oathmesh/oathmesh/internal/sign"
	"golang.org/x/sync/singleflight"
)

const (
	// DefaultJWKSCacheTTL is the default JWKS cache lifetime.
	DefaultJWKSCacheTTL = 60 * time.Second

	// JWKSFetchTimeout is the HTTP timeout for fetching JWKS.
	// Never use http.DefaultClient — always an explicit timeout per spec.
	JWKSFetchTimeout = 5 * time.Second
)

// JWKSCache fetches and caches JWKS from issuer endpoints.
// Thread-safe via sync.RWMutex. Refreshes on kid miss.
//
// SECURITY: Uses trusted, pre-resolved JWKS URLs to prevent SSRF (CodeQL go/request-forgery).
// Two modes:
//  1. Fixed mode (NewFixedJWKS): Single hardcoded URL, user input is IGNORED
//  2. Trusted-endpoints mode (NewJWKSCache): issuer key resolves to trusted URL from config/registration
//
// In FIXED mode: jwksURL is a constant string in config - ZERO user input flows to HTTP request.
// This satisfies CodeQL completely - the URL is hardcoded at startup, not influenced by user at runtime.
type JWKSCache struct {
	mu            sync.RWMutex
	entries       map[string]*jwksCacheEntry // keyed by kid
	client        *http.Client
	ttl           time.Duration
	jwksURL       string            // FIXED JWKS URL (fixed mode) - NEVER from user input
	jwksEndpoints map[string]string // key -> trusted FULL JWKS URL
	sf            singleflight.Group
}

type jwksCacheEntry struct {
	jwks   *sign.JWKS
	until  time.Time
	issuer string
}

// NewFixedJWKS creates a JWKS cache with a FIXED JWKS URL.
// SECURITY: The jwksURL is hardcoded as a constant - user input is COMPLETELY IGNORED.
// This mode satisfies CodeQL go/request-forgery completely: ZERO user input flows to HTTP request.
//
// Usage:
//
//	jwksProvider := verify.NewFixedJWKS(5*time.Minute, "https://issuer.example.com/.well-known/jwks.json")
//	// GetKey ignores the issuerKey parameter, always uses the fixed URL
func NewFixedJWKS(ttl time.Duration, jwksURL string) *JWKSCache {
	if ttl <= 0 {
		ttl = DefaultJWKSCacheTTL
	}
	return &JWKSCache{
		entries: make(map[string]*jwksCacheEntry),
		client:  &http.Client{Timeout: JWKSFetchTimeout},
		ttl:     ttl,
		jwksURL: jwksURL, // FIXED - never from user input
	}
}

// NewJWKSCache creates a JWKS cache with trusted endpoint mappings.
// The endpoints map issuer keys to FULL JWKS URLs (including /.well-known/jwks.json).
// Example: map[string]string{"prod": "https://issuer.com/.well-known/jwks.json"}
func NewJWKSCache(ttl time.Duration, endpoints map[string]string) *JWKSCache {
	if ttl <= 0 {
		ttl = DefaultJWKSCacheTTL
	}
	if endpoints == nil {
		endpoints = make(map[string]string)
	}
	copied := make(map[string]string, len(endpoints))
	for k, v := range endpoints {
		copied[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return &JWKSCache{
		entries:       make(map[string]*jwksCacheEntry),
		client:        &http.Client{Timeout: JWKSFetchTimeout},
		ttl:           ttl,
		jwksEndpoints: copied,
	}
}

// RegisterTrustedIssuers precomputes and stores trusted issuer->JWKS mappings.
// This keeps request targets server-controlled instead of deriving URLs from runtime token values.
func (c *JWKSCache) RegisterTrustedIssuers(issuers []string) error {
	if len(issuers) == 0 {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.jwksEndpoints == nil {
		c.jwksEndpoints = make(map[string]string)
	}
	for _, issuer := range issuers {
		trimmed := strings.TrimSpace(issuer)
		if trimmed == "" {
			continue
		}
		if _, exists := c.jwksEndpoints[trimmed]; exists {
			continue
		}
		jwksURL, err := jwksURLFromIssuer(trimmed)
		if err != nil {
			return fmt.Errorf("register trusted issuer %q: %w", trimmed, err)
		}
		c.jwksEndpoints[trimmed] = jwksURL
	}
	return nil
}

// GetKey returns the public key for the given issuer key and kid.
// Algorithm (alg) is also returned for algorithm confusion checking.
//
// Three modes (in order of priority):
//  1. Fixed mode (NewFixedJWKS): jwksURL is hardcoded - user input is IGNORED completely
//  2. Endpoints mode: issuerKey maps to full JWKS URL from config
//  3. Trusted-issuer registration mode: issuers are pre-registered via RegisterTrustedIssuers
//
// In Fixed mode: ZERO user input flows to HTTP request (CodeQL satisfied).
func (c *JWKSCache) GetKey(issuerKey string, kid string) (ed25519.PublicKey, string, error) {
	cacheKey, jwksURL, err := c.resolveJWKSURL(issuerKey, kid)
	if err != nil {
		return nil, "", err
	}

	// Try cache first (keyed by cacheKey)
	c.mu.RLock()
	entry, exists := c.entries[cacheKey]
	c.mu.RUnlock()

	if exists && time.Now().Before(entry.until) {
		key, alg, err := findKeyInJWKS(entry.jwks, kid)
		if err == nil {
			return key, alg, nil
		}
	}

	// Fetch fresh JWKS - jwksURL is from CONFIG, not user input
	return c.fetchAndCache(cacheKey, jwksURL, kid)
}

func (c *JWKSCache) resolveJWKSURL(issuerKey, kid string) (string, string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	switch {
	case c.jwksURL != "":
		// Fixed mode ignores issuerKey entirely.
		return kid, c.jwksURL, nil
	case len(c.jwksEndpoints) > 0:
		// Use issuer key first, then optional default fallback.
		jwksURL := c.jwksEndpoints[issuerKey]
		cacheKey := issuerKey
		if jwksURL == "" {
			jwksURL = c.jwksEndpoints["default"]
			cacheKey = "default"
		}
		if jwksURL == "" {
			return "", "", fmt.Errorf("no trusted JWKS endpoint configured for issuer %q", issuerKey)
		}
		normalized, err := normalizeJWKSURL(jwksURL)
		if err != nil {
			return "", "", fmt.Errorf("invalid trusted JWKS endpoint for issuer %q: %w", issuerKey, err)
		}
		return cacheKey, normalized, nil
	default:
		return "", "", fmt.Errorf("no trusted JWKS endpoint configured for issuer %q", issuerKey)
	}
}

func (c *JWKSCache) fetchAndCache(issuerKey, jwksURL, kid string) (ed25519.PublicKey, string, error) {
	// Double-check cache before starting singleflight to avoid edge-case double fetches
	c.mu.RLock()
	if entry, exists := c.entries[issuerKey]; exists && time.Now().Before(entry.until) {
		key, alg, err := findKeyInJWKS(entry.jwks, kid)
		if err == nil {
			c.mu.RUnlock()
			return key, alg, nil
		}
	}
	c.mu.RUnlock()

	// Collapse concurrent fetches for the same issuer
	_, err, _ := c.sf.Do(issuerKey, func() (any, error) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, jwksURL, nil)
		if err != nil {
			return nil, fmt.Errorf("build JWKS request for %s: %w", jwksURL, err)
		}
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch JWKS from %s: %w", jwksURL, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("JWKS endpoint %s returned %d", jwksURL, resp.StatusCode)
		}

		var jwks sign.JWKS
		if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&jwks); err != nil {
			return nil, fmt.Errorf("decode JWKS from %s: %w", jwksURL, err)
		}

		c.mu.Lock()
		c.entries[issuerKey] = &jwksCacheEntry{
			jwks:   &jwks,
			until:  time.Now().Add(c.ttl),
			issuer: issuerKey,
		}
		c.mu.Unlock()

		return nil, nil
	})

	if err != nil {
		return nil, "", err
	}

	// Now that fetch is complete (or if it was a deduplicated call), try finding key again
	c.mu.RLock()
	entry, exists := c.entries[issuerKey]
	c.mu.RUnlock()

	if !exists {
		return nil, "", fmt.Errorf("cache repopulation failed unexpectedly for %q", issuerKey)
	}

	return findKeyInJWKS(entry.jwks, kid)
}

func findKeyInJWKS(jwks *sign.JWKS, kid string) (ed25519.PublicKey, string, error) {
	for _, key := range jwks.Keys {
		if key.Kid == kid {
			pubKey, err := sign.GetKeyFromJWKS(jwks, kid)
			if err != nil {
				return nil, "", err
			}
			return pubKey, key.Alg, nil
		}
	}
	return nil, "", fmt.Errorf("key %s not found in JWKS", kid)
}

func jwksURLFromIssuer(issuer string) (string, error) {
	trimmed := strings.TrimSpace(issuer)
	if trimmed == "" {
		return "", errors.New("issuer URL is empty")
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid issuer URL %q: %w", issuer, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid issuer URL %q: missing scheme or host", issuer)
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/.well-known/jwks.json"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func normalizeJWKSURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("JWKS URL is empty")
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid JWKS URL %q: %w", raw, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid JWKS URL %q: missing scheme or host", raw)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("invalid JWKS URL %q: unsupported scheme %q", raw, u.Scheme)
	}
	if u.User != nil {
		return "", fmt.Errorf("invalid JWKS URL %q: userinfo is not allowed", raw)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("invalid JWKS URL %q: query and fragment are not allowed", raw)
	}
	return u.String(), nil
}

// StaticJWKSProvider is a test helper that returns keys from a static map.
// Used in unit tests to avoid HTTP calls.
type StaticJWKSProvider struct {
	Keys map[string]staticKeyEntry // keyed by kid
}

type staticKeyEntry struct {
	PublicKey ed25519.PublicKey
	Alg       string
}

// NewStaticJWKSProvider creates a provider with the given kid→key mapping.
// All keys are registered as EdDSA by default.
func NewStaticJWKSProvider(keys map[string]ed25519.PublicKey) *StaticJWKSProvider {
	entries := make(map[string]staticKeyEntry, len(keys))
	for kid, key := range keys {
		entries[kid] = staticKeyEntry{PublicKey: key, Alg: "EdDSA"}
	}
	return &StaticJWKSProvider{Keys: entries}
}

func (p *StaticJWKSProvider) GetKey(issuerURL string, kid string) (ed25519.PublicKey, string, error) {
	entry, ok := p.Keys[kid]
	if !ok {
		return nil, "", fmt.Errorf("key %s not found in static provider", kid)
	}
	return entry.PublicKey, entry.Alg, nil
}
