package audit

import (
	"context"

	"github.com/oathmesh/oathmesh/internal/core"
)

// FanOutAuditSink implements core.AuditSink by dispatching each event to
// multiple downstream sinks. This enables composing audit logging, metrics,
// and tracing without introducing a separate observer interface.
//
// Error semantics: the first error encountered is returned, but all sinks
// are always invoked regardless of earlier failures — no short-circuit.
type FanOutAuditSink struct {
	sinks []core.AuditSink
}

// NewFanOutAuditSink creates a fan-out sink that dispatches to all provided sinks.
// Panics if no sinks are provided — a fan-out with zero sinks is a configuration error.
func NewFanOutAuditSink(sinks ...core.AuditSink) *FanOutAuditSink {
	if len(sinks) == 0 {
		panic("oathmesh: FanOutAuditSink requires at least one sink")
	}
	return &FanOutAuditSink{sinks: sinks}
}

// Emit dispatches the event to all sinks. Returns the first error encountered,
// but always invokes every sink.
func (f *FanOutAuditSink) Emit(ctx context.Context, event *core.AuditEvent) error {
	var firstErr error
	for _, s := range f.sinks {
		if err := s.Emit(ctx, event); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Compile-time check that FanOutAuditSink implements core.AuditSink.
var _ core.AuditSink = (*FanOutAuditSink)(nil)
