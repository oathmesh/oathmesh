package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/oathmesh/oathmesh/internal/metrics"
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
	if err := rl.sync(); err != nil {
		log.Printf("ERROR: revocation list sync failed: %v", err)
		metrics.RevocationSyncErrors.Inc()
	}

	go rl.poll(pollInterval)
	return rl
}

func (rl *MemoryRevocationList) poll(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := rl.sync(); err != nil {
				log.Printf("ERROR: revocation list sync failed: %v", err)
				metrics.RevocationSyncErrors.Inc()
			}
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

// IsRevoked checks if a subject has been revoked.
// OathMesh revocation policy: if a subject is revoked, ALL tokens for that subject
// are invalid regardless of when they were issued (iat). Revocation invalidates
// all existing tokens and prevents any future tokens for that subject.
func (rl *MemoryRevocationList) IsRevoked(ctx context.Context, subject string) (bool, error) {
	rl.mu.RLock()
	_, exists := rl.revocations[subject]
	rl.mu.RUnlock()

	if exists {
		return true, nil
	}
	return false, nil
}

// Revoke manually adds a subject to the revocation list.
// Primarily used for testing or when bypassing the background sync.
func (rl *MemoryRevocationList) Revoke(subject string) {
	rl.mu.Lock()
	rl.revocations[subject] = time.Now()
	rl.mu.Unlock()
}

// Close stops the polling goroutine.
func (rl *MemoryRevocationList) Close() {
	close(rl.done)
}

var _ RevocationList = (*MemoryRevocationList)(nil)
