package audit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"

	"github.com/oathmesh/oathmesh/internal/core"
)

// HMACChainAuditSink wraps an existing AuditSink and injects HMAC chaining
// directly into the core.AuditEvent before passing it to the underlying sink.
type HMACChainAuditSink struct {
	mu       sync.Mutex
	sink     core.AuditSink
	key      []byte
	seq      uint64
	lastHash string
}

// NewHMACChainAuditSink creates a new chained sink.
func NewHMACChainAuditSink(sink core.AuditSink, key []byte) *HMACChainAuditSink {
	return &HMACChainAuditSink{
		sink:     sink,
		key:      key,
		seq:      0,
		lastHash: "",
	}
}

// Emit secures the event and delegates it.
func (h *HMACChainAuditSink) Emit(ctx context.Context, event *core.AuditEvent) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. Populate chain state without HMAC
	event.Seq = h.seq
	event.PrevHash = h.lastHash
	event.HMAC = ""

	// 2. Marshal to compute hash
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// 3. Compute HMAC(key, json)
	mac := hmac.New(sha256.New, h.key)
	mac.Write(data)
	event.HMAC = hex.EncodeToString(mac.Sum(nil))

	// Re-marshal to calculate the next prevHash (now including the HMAC!)
	finalData, _ := json.Marshal(event)

	// Update chain state
	hash := sha256.Sum256(finalData)
	h.lastHash = hex.EncodeToString(hash[:])
	h.seq++

	// 4. Delegate to underlying sink
	return h.sink.Emit(ctx, event)
}

// Compile-time check
var _ core.AuditSink = (*HMACChainAuditSink)(nil)
