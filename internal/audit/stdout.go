package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/oathmesh/oathmesh/internal/core"
)

// StdoutAuditSink writes NDJSON audit events to stdout.
// This is the default sink for container and cloud-native deployments.
type StdoutAuditSink struct{}

// NewStdoutAuditSink creates a new stdout audit sink.
func NewStdoutAuditSink() *StdoutAuditSink {
	return &StdoutAuditSink{}
}

// Emit writes an audit event as a single-line JSON object to stdout.
// Never logs the full Oath Token string — jti + claim summary only.
func (s *StdoutAuditSink) Emit(ctx context.Context, event *core.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	// Write as NDJSON (one JSON object per line)
	_, err = fmt.Fprintln(os.Stdout, string(data))
	return err
}

// Compile-time check that StdoutAuditSink implements core.AuditSink.
var _ core.AuditSink = (*StdoutAuditSink)(nil)
