package policy

import (
	"encoding/json"
	"fmt"
	"os"
)

// ValidateFile validates a policy file (JSON format) for structural correctness.
// Returns nil if the policy is valid, or an error describing the problem.
//
// In production with Pkl installed, use `pkl eval <file>` for full schema validation.
// This Go wrapper handles JSON policy files used in environments without the Pkl runtime.
func ValidateFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read policy file: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse policy file: %w", err)
	}

	// Check required top-level fields
	if _, ok := raw["version"]; !ok {
		return fmt.Errorf("policy missing required field: version")
	}
	if _, ok := raw["issuers"]; !ok {
		return fmt.Errorf("policy missing required field: issuers")
	}
	if _, ok := raw["audiences"]; !ok {
		return fmt.Errorf("policy missing required field: audiences")
	}
	rules, ok := raw["rules"]
	if !ok {
		return fmt.Errorf("policy missing required field: rules")
	}

	// Check rules is an array
	rulesArr, ok := rules.([]interface{})
	if !ok {
		return fmt.Errorf("policy field 'rules' must be an array")
	}

	if len(rulesArr) == 0 {
		return fmt.Errorf("policy must have at least one rule")
	}

	// Check last rule is default deny
	lastRule, ok := rulesArr[len(rulesArr)-1].(map[string]interface{})
	if !ok {
		return fmt.Errorf("last rule must be an object")
	}
	if name, _ := lastRule["name"].(string); name != "default" {
		return fmt.Errorf("last rule must have name 'default', got '%s'", name)
	}
	if allow, _ := lastRule["allow"].(bool); allow {
		return fmt.Errorf("last rule (default) must have allow: false")
	}

	return nil
}
