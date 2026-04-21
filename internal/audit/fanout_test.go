package audit

import (
	"context"
	"fmt"
	"testing"

	"github.com/oathmesh/oathmesh/internal/core"
)

// recordingSink captures emitted events for assertion.
type recordingSink struct {
	events []*core.AuditEvent
}

func (r *recordingSink) Emit(_ context.Context, event *core.AuditEvent) error {
	r.events = append(r.events, event)
	return nil
}

// errorSink always returns an error.
type errorSink struct {
	err error
}

func (e *errorSink) Emit(_ context.Context, _ *core.AuditEvent) error {
	return e.err
}

func TestFanOutAuditSink_DispatchesToAllSinks(t *testing.T) {
	sink1 := &recordingSink{}
	sink2 := &recordingSink{}
	sink3 := &recordingSink{}

	fanout := NewFanOutAuditSink(sink1, sink2, sink3)

	event := &core.AuditEvent{
		Event:   "verify",
		Outcome: "allow",
	}

	err := fanout.Emit(context.Background(), event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	for i, sink := range []*recordingSink{sink1, sink2, sink3} {
		if len(sink.events) != 1 {
			t.Errorf("sink %d: expected 1 event, got %d", i, len(sink.events))
		}
		if sink.events[0].Event != "verify" {
			t.Errorf("sink %d: expected event 'verify', got %q", i, sink.events[0].Event)
		}
	}
}

func TestFanOutAuditSink_ReturnsFirstError_InvokesAll(t *testing.T) {
	sink1 := &recordingSink{}
	errSink := &errorSink{err: fmt.Errorf("disk full")}
	sink3 := &recordingSink{}

	fanout := NewFanOutAuditSink(sink1, errSink, sink3)

	event := &core.AuditEvent{
		Event:   "verify",
		Outcome: "deny",
	}

	err := fanout.Emit(context.Background(), event)
	if err == nil {
		t.Fatal("expected error from failing sink")
	}
	if err.Error() != "disk full" {
		t.Fatalf("expected 'disk full', got: %v", err)
	}

	// All sinks should still have been invoked
	if len(sink1.events) != 1 {
		t.Error("sink1 was not invoked")
	}
	if len(sink3.events) != 1 {
		t.Error("sink3 was not invoked despite earlier sink error")
	}
}

func TestFanOutAuditSink_PanicsOnZeroSinks(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic with zero sinks")
		}
	}()
	NewFanOutAuditSink()
}
