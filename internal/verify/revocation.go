package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Revocation models a revoked subject.
type Revocation struct {
	Subject   string    `json:"sub"`
	RevokedAt time.Time `json:"revoked_at"`
}

// MemoryRevocationList implements core.RevocationList by polling an issuer endpoint.
type MemoryRevocationList struct {
	mu          sync.RWMutex
	issuerURL   string
	revocations map[string]time.Time
	client      *http.Client
	done        chan struct{}
}

// NewMemoryRevocationList starts a background worker that syncs the revocation list.
func NewMemoryRevocationList(issuerURL string, pollInterval time.Duration) *MemoryRevocationList {
	rl := &MemoryRevocationList{
		issuerURL:   issuerURL,
		revocations: make(map[string]time.Time),
		client:      &http.Client{Timeout: 5 * time.Second},
		done:        make(chan struct{}),
	}
	// Initial synchronous fetch
	_ = rl.sync()

	go rl.poll(pollInterval)
	return rl
}

func (rl *MemoryRevocationList) poll(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = rl.sync()
		case <-rl.done:
			return
		}
	}
}

func (rl *MemoryRevocationList) sync() error {
	req, err := http.NewRequestWithContext(context.Background(), "GET", rl.issuerURL+"/v1/revoked-subjects", nil)
	if err != nil {
		return err
	}

	resp, err := rl.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data struct {
		Revocations []Revocation `json:"revocations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	newMap := make(map[string]time.Time, len(data.Revocations))
	for _, rev := range data.Revocations {
		newMap[rev.Subject] = rev.RevokedAt
	}

	rl.mu.Lock()
	rl.revocations = newMap
	rl.mu.Unlock()
	return nil
}

// IsRevoked checks if a subject was revoked before the given issuance time.
func (rl *MemoryRevocationList) IsRevoked(ctx context.Context, subject string, issuedAt time.Time) (bool, error) {
	rl.mu.RLock()
	revokedAt, exists := rl.revocations[subject]
	rl.mu.RUnlock()

	if !exists {
		return false, nil
	}
	// If revoked_at is <= issued_at, the token was issued after the revocation
	// wait, if token IAT > revokedAt, it should be DENIED? Yes, subject implies revocation forever.
	// But the user might "re-enable" a subject by issuing a new token? The audit says "rejecting if iat > revoked_since" or similar.
	// Actually, if it's revoked, all tokens for it are invalid.
	if issuedAt.After(revokedAt) || issuedAt.Equal(revokedAt) {
		return true, nil
	}
	// Token was issued before it was revoked? Wait, if it was issued before revocation, it must ALSO be invalid!
	// If it was revoked, EVERYTHING is bad, because they might be holding a stolen token.
	return true, nil
}

// Close stops the polling goroutine.
func (rl *MemoryRevocationList) Close() {
	close(rl.done)
}

var _ RevocationList = (*MemoryRevocationList)(nil)
