package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LoadPolicyFromFile loads a policy from a file path.
// Supports JSON files directly, and Pkl files if the pkl binary is available.
//
// Loading order:
//  1. If file ends in .json: parse directly as JSON
//  2. If file ends in .pkl: shell out to `pkl eval --format json <file>` if pkl is in PATH
//  3. Otherwise: try JSON parsing first
func LoadPolicyFromFile(path string) (*Policy, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".json":
		return loadFromJSON(path)
	case ".pkl":
		return loadFromPkl(path)
	default:
		// Try JSON first
		p, err := loadFromJSON(path)
		if err == nil {
			return p, nil
		}
		return nil, fmt.Errorf("unsupported policy file format %q: %w", ext, err)
	}
}

// LoadPolicyFromJSON loads a policy from a JSON byte slice.
func LoadPolicyFromJSON(data []byte) (*Policy, error) {
	var p Policy
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse policy JSON: %w", err)
	}

	if err := ValidatePolicy(&p); err != nil {
		return nil, fmt.Errorf("validate policy: %w", err)
	}

	return &p, nil
}

func loadFromJSON(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy file: %w", err)
	}
	return LoadPolicyFromJSON(data)
}

func loadFromPkl(path string) (*Policy, error) {
	// Check if pkl binary is available
	pklPath, err := exec.LookPath("pkl")
	if err != nil {
		return nil, fmt.Errorf("pkl binary not found in PATH — install pkl or convert policy to JSON with: pkl eval --format json %s > policy.json", path)
	}

	// Shell out to pkl eval --format json
	cmd := exec.Command(pklPath, "eval", "--format", "json", path)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("pkl eval failed: %s\n%s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("pkl eval failed: %w", err)
	}

	return LoadPolicyFromJSON(output)
}
