package audit

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/oathmesh/oathmesh/internal/core"
)

func testEvent() *core.AuditEvent {
	return &core.AuditEvent{
		Event:     "oathmesh.verify",
		Outcome:   "allow",
		Reason:    "policy:storefront-read",
		JTI:       "550e8400-e29b-41d4-a716-446655440000",
		Sub:       "agent://repo/acme/deploy-bot",
		Aud:       "https://inventory.internal",
		Act:       "inventory.write",
		Iss:       "https://issuer.oathmesh.tech",
		Env:       "prod",
		Timestamp: time.Date(2026, 4, 12, 14, 30, 0, 0, time.UTC),
		Source: &core.Source{
			Type:     "github_actions",
			Repo:     "acme/storefront",
			Workflow: "deploy.yml",
		},
	}
}

func TestStdoutSink_EmitsNDJSON(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	sink := NewStdoutAuditSink()
	err := sink.Emit(context.Background(), testEvent())
	if err != nil {
		t.Fatalf("emit error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Verify JSON structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}

	if parsed["event"] != "oathmesh.verify" {
		t.Errorf("expected event oathmesh.verify, got %v", parsed["event"])
	}
	if parsed["outcome"] != "allow" {
		t.Errorf("expected outcome allow, got %v", parsed["outcome"])
	}
	if parsed["jti"] != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected jti, got %v", parsed["jti"])
	}
}

func TestFileSink_AppendsNDJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "audit.ndjson")

	sink, err := NewFileAuditSink(path)
	if err != nil {
		t.Fatalf("create file sink: %v", err)
	}
	defer sink.Close()

	// Write two events
	err = sink.Emit(context.Background(), testEvent())
	if err != nil {
		t.Fatalf("emit error: %v", err)
	}

	event2 := testEvent()
	event2.Outcome = "deny"
	event2.Reason = "policy_denied"
	err = sink.Emit(context.Background(), event2)
	if err != nil {
		t.Fatalf("emit error: %v", err)
	}

	// Read back and verify NDJSON format (one JSON object per line)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		var parsed map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &parsed); err != nil {
			t.Errorf("line %d is not valid JSON: %v", lineCount, err)
		}
	}

	if lineCount != 2 {
		t.Errorf("expected 2 NDJSON lines, got %d", lineCount)
	}
}

func TestFileSink_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "audit-concurrent.ndjson")

	sink, err := NewFileAuditSink(path)
	if err != nil {
		t.Fatalf("create file sink: %v", err)
	}
	defer sink.Close()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if err := sink.Emit(context.Background(), testEvent()); err != nil {
				t.Errorf("emit error: %v", err)
			}
		}()
	}

	wg.Wait()

	// Read back and verify no corruption
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		var parsed map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &parsed); err != nil {
			t.Errorf("line %d is corrupted JSON: %v\ncontent: %s", lineCount, err, scanner.Text())
		}
	}

	if lineCount != goroutines {
		t.Errorf("expected %d NDJSON lines, got %d", goroutines, lineCount)
	}
}

func TestNoopSink_NoOutput(t *testing.T) {
	sink := NewNoopAuditSink()
	err := sink.Emit(context.Background(), testEvent())
	if err != nil {
		t.Errorf("noop sink should return nil, got: %v", err)
	}
}
