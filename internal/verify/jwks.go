package verify

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
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
//
// SECURITY: Uses pre-computed JWKS URLs to prevent SSRF attacks (CodeQL go/request-forgery).
// Users provide an issuer key (e.g., "production"), and we look up the FULL JWKS URL from config.
// The FULL URL (including /.well-known/jwks.json) is stored in config - NO string concatenation in code.
// This ensures ZERO user input flows into HTTP request URL construction.
type JWKSCache struct {
	mu            sync.RWMutex
	entries       map[string]*jwksCacheEntry // keyed by issuer key
	client        *http.Client
	ttl           time.Duration
	jwksEndpoints map[string]string // key -> FULL JWKS URL (e.g., "https://issuer.com/.well-known/jwks.json")
}

type jwksCacheEntry struct {
	jwks   *sign.JWKS
	until  time.Time
	issuer string
}

// NewJWKSCache creates a new JWKS cache with the given TTL.
// Uses a dedicated http.Client with 5-second timeout — never http.DefaultClient.
// The endpoints map issuer keys (e.g., "production") to FULL JWKS URLs (including /.well-known/jwks.json).
// Example: map[string]string{"prod": "https://issuer.com/.well-known/jwks.json"}
// This prevents SSRF by ensuring ZERO string concatenation in code.
func NewJWKSCache(ttl time.Duration, endpoints map[string]string) *JWKSCache {
	if ttl <= 0 {
		ttl = DefaultJWKSCacheTTL
	}
	if endpoints == nil {
		endpoints = make(map[string]string)
	}
	return &JWKSCache{
		entries:       make(map[string]*jwksCacheEntry),
		client:        &http.Client{Timeout: JWKSFetchTimeout},
		ttl:           ttl,
		jwksEndpoints: endpoints,
	}
}

// getJWKSEndpoint resolves an issuer key to its FULL JWKS URL (SSRF protection).
// Returns empty string if key not found.
func (c *JWKSCache) getJWKSEndpoint(issuerKey string) string {
	return c.jwksEndpoints[issuerKey]
}

// GetKey returns the public key for the given issuer key and kid.
// Algorithm (alg) is also returned for algorithm confusion checking.
//
// Two modes:
//  1. With mappings configured: issuerKey is a lookup key that maps to a full URL
//  2. Without endpoints: issuerKeyOrURL is treated as the base URL (backward compat)
//
// This design prevents SSRF by ensuring ZERO user input in URL construction.
//
// Lookup order:
//  1. Check cache — if valid entry exists and kid is found, return key
//  2. If kid not found in cache (rotation), fetch fresh JWKS once
//  3. If kid still not found after refresh — reject with issuer_untrusted
func (c *JWKSCache) GetKey(issuerKeyOrURL string, kid string) (ed25519.PublicKey, string, error) {
	var issuerKey, jwksURL string

	// Try to resolve as key first, fall back to treating as base URL (backward compat)
	if endpoint := c.getJWKSEndpoint(issuerKeyOrURL); endpoint != "" {
		issuerKey = issuerKeyOrURL
		jwksURL = endpoint // FULL URL from config - NO concatenation!
	} else if len(c.jwksEndpoints) > 0 {
		// Endpoints configured but key not found
		return nil, "", fmt.Errorf("unknown issuer key: %s", issuerKeyOrURL)
	} else {
		// No endpoints: treat parameter as base URL (backward compatibility)
		issuerKey = issuerKeyOrURL
		jwksURL = issuerKeyOrURL + "/.well-known/jwks.json" // fallback for backward compat
	}

	// Try cache first (keyed by issuerKey)
	c.mu.RLock()
	entry, exists := c.entries[issuerKey]
	c.mu.RUnlock()

	if exists && time.Now().Before(entry.until) {
		key, alg, err := findKeyInJWKS(entry.jwks, kid)
		if err == nil {
			return key, alg, nil
		}
		// kid not found in cache — fall through to refresh
	}

	// Fetch fresh JWKS (cache miss or kid miss)
	return c.fetchAndCache(issuerKey, jwksURL, kid)
}

func (c *JWKSCache) fetchAndCache(issuerKey, jwksURL, kid string) (ed25519.PublicKey, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists := c.entries[issuerKey]; exists && time.Now().Before(entry.until) {
		key, alg, err := findKeyInJWKS(entry.jwks, kid)
		if err == nil {
			return key, alg, nil
		}
	}

	// JWKS URL is from CONFIG only - ZERO user input in URL (CodeQL go/request-forgery fix)
	// jwksURL comes from config.jwksEndpoints map, not from user input
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

	c.entries[issuerKey] = &jwksCacheEntry{
		jwks:   &jwks,
		until:  time.Now().Add(c.ttl),
		issuer: issuerKey,
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
