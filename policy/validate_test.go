package policy

import (
	"testing"
)

func TestValidateFile_Valid(t *testing.T) {
	err := ValidateFile("../policy/production.json")
	if err != nil {
		t.Fatalf("expected valid policy, got: %v", err)
	}
}

func TestValidateFile_MissingFile(t *testing.T) {
	err := ValidateFile("nonexistent.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
