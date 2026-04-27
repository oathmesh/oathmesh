package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeygenFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test_private.pem")

	cmd := buildKeygenCmd()
	cmd.SetArgs([]string{"--out", outPath})
	
	if err := cmd.Execute(); err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat private key failed: %v", err)
	}

	if info.Size() == 0 {
		t.Errorf("private key file is empty")
	}
	
	pubInfo, err := os.Stat(outPath + ".pub")
	if err != nil {
		t.Fatalf("stat public key failed: %v", err)
	}
	if pubInfo.Size() == 0 {
		t.Errorf("public key file is empty")
	}
}
