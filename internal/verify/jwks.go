package verify

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/oathmesh/oathmesh/internal/sign"
)

const (
	// DefaultJWKSCacheTTL is the default JWKS cache lifetime.
	DefaultJWKSCacheTTL = 5 * time.Minute

	// JWKSFetchTimeout is the HTTP timeout for fetching JWKS.
	// Never use http.DefaultClient — always an explicit timeout per spec.
	JWKSFetchTimeout = 5 * time.Second
)

// JWKSCache fetches and caches JWKS from issuer endpoints.
// Thread-safe via sync.RWMutex. Refreshes on kid miss.
type JWKSCache struct {
	mu      sync.RWMutex
	entries map[string]*jwksCacheEntry // keyed by issuer URL
	client  *http.Client
	ttl     time.Duration
}

type jwksCacheEntry struct {
	jwks   *sign.JWKS
	until  time.Time
	issuer string
}

// NewJWKSCache creates a new JWKS cache with the given TTL.
// Uses a dedicated http.Client with 5-second timeout — never http.DefaultClient.
func NewJWKSCache(ttl time.Duration) *JWKSCache {
	if ttl <= 0 {
		ttl = DefaultJWKSCacheTTL
	}
	return &JWKSCache{
		entries: make(map[string]*jwksCacheEntry),
		client:  &http.Client{Timeout: JWKSFetchTimeout},
		ttl:     ttl,
	}
}

// GetKey returns the public key for the given issuer URL and kid.
// Algorithm (alg) is also returned for algorithm confusion checking.
//
// Lookup order:
//  1. Check cache — if valid entry exists and kid is found, return key
//  2. If kid not found in cache (rotation), fetch fresh JWKS once
//  3. If kid still not found after refresh — reject with issuer_untrusted
func (c *JWKSCache) GetKey(issuerURL string, kid string) (ed25519.PublicKey, string, error) {
	// Try cache first
	c.mu.RLock()
	entry, exists := c.entries[issuerURL]
	c.mu.RUnlock()

	if exists && time.Now().Before(entry.until) {
		key, alg, err := findKeyInJWKS(entry.jwks, kid)
		if err == nil {
			return key, alg, nil
		}
		// kid not found in cache — fall through to refresh
	}

	// Fetch fresh JWKS (cache miss or kid miss)
	return c.fetchAndCache(issuerURL, kid)
}

func (c *JWKSCache) fetchAndCache(issuerURL string, kid string) (ed25519.PublicKey, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists := c.entries[issuerURL]; exists && time.Now().Before(entry.until) {
		key, alg, err := findKeyInJWKS(entry.jwks, kid)
		if err == nil {
			return key, alg, nil
		}
	}

	// Validate issuer URL before making HTTP request (prevent SSRF)
	parsed, err := url.Parse(issuerURL)
	if err != nil {
		return nil, "", fmt.Errorf("invalid issuer URL: %w", err)
	}
	if !parsed.IsAbs() {
		return nil, "", fmt.Errorf("issuer URL must be absolute: %s", issuerURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, "", fmt.Errorf("issuer URL must use http or https: %s", issuerURL)
	}

	jwksURL := issuerURL + "/.well-known/jwks.json"
	resp, err := c.client.Get(jwksURL)
	if err != nil {
		return nil, "", fmt.Errorf("fetch JWKS from %s: %w", jwksURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("JWKS endpoint %s returned %d", jwksURL, resp.StatusCode)
	}

	var jwks sign.JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, "", fmt.Errorf("decode JWKS from %s: %w", jwksURL, err)
	}

	c.entries[issuerURL] = &jwksCacheEntry{
		jwks:   &jwks,
		until:  time.Now().Add(c.ttl),
		issuer: issuerURL,
	}

	key, alg, err := findKeyInJWKS(&jwks, kid)
	if err != nil {
		return nil, "", err
	}
	return key, alg, nil
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
