package audit

import (
	"context"

	"github.com/MustafaMahmoudAtta111/oathmesh/internal/core"
)

// NoopAuditSink discards all audit events.
//
// ⚠ TESTING ONLY — never use as default in production.
// Audit events are a first-class requirement of the OathMesh protocol.
// Disabling them in production violates the security model.
type NoopAuditSink struct{}

// NewNoopAuditSink creates a new noop audit sink.
// ⚠ TESTING ONLY — never use as default in production.
func NewNoopAuditSink() *NoopAuditSink {
	return &NoopAuditSink{}
}

// Emit discards the event and returns nil.
func (s *NoopAuditSink) Emit(ctx context.Context, event *core.AuditEvent) error {
	return nil
}

// Compile-time check that NoopAuditSink implements core.AuditSink.
var _ core.AuditSink = (*NoopAuditSink)(nil)
