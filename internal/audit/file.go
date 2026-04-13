package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/oathmesh/oathmesh/internal/core"
)

// FileAuditSink writes NDJSON audit events to a file.
// Append-only, mutex-protected for concurrent writes.
type FileAuditSink struct {
	mu   sync.Mutex
	file *os.File
	path string
}

// NewFileAuditSink creates a new file audit sink at the given path.
// Opens the file in append-only mode, creating it if it doesn't exist.
func NewFileAuditSink(path string) (*FileAuditSink, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}

	return &FileAuditSink{
		file: file,
		path: path,
	}, nil
}

// Emit writes an audit event as a single-line JSON object to the file.
// Thread-safe via mutex. Never logs the full Oath Token string.
func (s *FileAuditSink) Emit(ctx context.Context, event *core.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}

// Close closes the underlying file handle.
func (s *FileAuditSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Close()
}

// Compile-time check that FileAuditSink implements core.AuditSink.
var _ core.AuditSink = (*FileAuditSink)(nil)
