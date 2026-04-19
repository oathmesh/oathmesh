package policy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oathmesh/oathmesh/internal/verify"
)

// ── Glob matching tests ─────────────────────────────────────────────────────

func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		// Exact match
		{"exact", "exact", true},
		{"exact", "other", false},

		// Single * matches everything
		{"*", "anything", true},
		{"*", "", true},

		// Suffix wildcard (prefix match) — critical for URI patterns
		{"agent://repo/acme/*", "agent://repo/acme/deploy-bot", true},
		{"agent://repo/acme/*", "agent://repo/acme/other-bot", true},
		{"agent://repo/acme/*", "agent://repo/other/deploy-bot", false},

		// Prefix wildcard (suffix match)
		{"*.yml", "deploy.yml", true},
		{"*.yml", "deploy.yaml", false},

		// Middle wildcard
		{"agent://*/deploy-bot", "agent://repo/deploy-bot", true},
		{"agent://*/deploy-bot", "agent://repo/acme/deploy-bot", true},

		// No match
		{"agent://repo/other/*", "agent://repo/acme/deploy-bot", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.value, func(t *testing.T) {
			got := globMatch(tt.pattern, tt.value)
			if got != tt.want {
				t.Errorf("globMatch(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
			}
		})
	}
}

// ── Policy evaluation tests ─────────────────────────────────────────────────

func testPolicy() *Policy {
	sub1 := "agent://repo/acme/*"
	act1 := "inventory.read"
	sub2 := "agent://repo/acme/deploy-bot"
	act2 := "inventory.write"
	srcType := "github_actions"
	srcRepo := "acme/storefront"
	srcWflow := "deploy.yml"

	return &Policy{
		Version:   1,
		Issuers:   []string{"https://issuer.oathmesh.tech"},
		Audiences: []string{"https://inventory.internal"},
		Rules: []Rule{
			{
				Name:  "storefront-read",
				Match: MatchCriteria{Sub: &sub1, Act: &act1},
				Allow: true,
			},
			{
				Name: "deploy-write",
				Match: MatchCriteria{
					Sub: &sub2,
					Act: &act2,
					Src: &SourceCriteria{
						Type:     &srcType,
						Repo:     &srcRepo,
						Workflow: &srcWflow,
					},
				},
				Allow: true,
			},
			{
				Name:  "default",
				Match: MatchCriteria{},
				Allow: false,
			},
		},
	}
}

func TestEvaluate_AllowMatch(t *testing.T) {
	pe := NewPolicyEngine(testPolicy())

	input := &verify.PolicyInput{
		Sub: "agent://repo/acme/deploy-bot",
		Act: "inventory.read",
	}

	decision := pe.Evaluate(input)
	if decision.Outcome != "allow" {
		t.Errorf("expected allow, got %s (rule: %s)", decision.Outcome, decision.RuleName)
	}
	if decision.RuleName != "storefront-read" {
		t.Errorf("expected rule storefront-read, got %s", decision.RuleName)
	}
}

func TestEvaluate_DenyNoMatch(t *testing.T) {
	pe := NewPolicyEngine(testPolicy())

	// Subject from a different org — should not match any allow rule
	input := &verify.PolicyInput{
		Sub: "agent://repo/other/deploy-bot",
		Act: "inventory.read",
	}

	decision := pe.Evaluate(input)
	if decision.Outcome != "deny" {
		t.Errorf("expected deny, got %s (rule: %s)", decision.Outcome, decision.RuleName)
	}
}

func TestEvaluate_DefaultDeny(t *testing.T) {
	// Empty policy (no rules except default deny)
	pe := NewPolicyEngine(&Policy{
		Version:   1,
		Issuers:   []string{"https://issuer.test"},
		Audiences: []string{"https://api.test"},
		Rules:     []Rule{{Name: "default", Allow: false}},
	})

	input := &verify.PolicyInput{
		Sub: "agent://repo/acme/anything",
		Act: "anything",
	}

	decision := pe.Evaluate(input)
	if decision.Outcome != "deny" {
		t.Errorf("expected deny for default-deny policy, got %s", decision.Outcome)
	}
	if decision.RuleName != "default" {
		t.Errorf("expected rule default, got %s", decision.RuleName)
	}
}

func TestEvaluate_NilPolicy(t *testing.T) {
	pe := NewPolicyEngine(nil)

	input := &verify.PolicyInput{Sub: "anything"}
	decision := pe.Evaluate(input)
	if decision.Outcome != "deny" {
		t.Errorf("expected deny for nil policy, got %s", decision.Outcome)
	}
}

func TestEvaluate_FirstMatchWins(t *testing.T) {
	sub := "agent://repo/acme/deploy-bot"
	act := "inventory.read"

	pe := NewPolicyEngine(&Policy{
		Version:   1,
		Issuers:   []string{"https://issuer.test"},
		Audiences: []string{"https://api.test"},
		Rules: []Rule{
			{Name: "early-deny", Match: MatchCriteria{Sub: &sub, Act: &act}, Allow: false},
			{Name: "late-allow", Match: MatchCriteria{Sub: &sub, Act: &act}, Allow: true},
			{Name: "default", Allow: false},
		},
	})

	input := &verify.PolicyInput{Sub: "agent://repo/acme/deploy-bot", Act: "inventory.read"}
	decision := pe.Evaluate(input)

	if decision.Outcome != "deny" {
		t.Errorf("expected early deny (first match wins), got %s", decision.Outcome)
	}
	if decision.RuleName != "early-deny" {
		t.Errorf("expected rule early-deny, got %s", decision.RuleName)
	}
}

func TestEvaluate_SourceCriteria(t *testing.T) {
	pe := NewPolicyEngine(testPolicy())

	// Correct source — should match deploy-write rule
	input := &verify.PolicyInput{
		Sub:      "agent://repo/acme/deploy-bot",
		Act:      "inventory.write",
		SrcType:  "github_actions",
		SrcRepo:  "acme/storefront",
		SrcWflow: "deploy.yml",
	}

	decision := pe.Evaluate(input)
	if decision.Outcome != "allow" {
		t.Errorf("expected allow with correct source, got %s (rule: %s)", decision.Outcome, decision.RuleName)
	}
	if decision.RuleName != "deploy-write" {
		t.Errorf("expected rule deploy-write, got %s", decision.RuleName)
	}

	// Wrong source repo — should fall through to default deny
	input.SrcRepo = "evil/repo"
	decision = pe.Evaluate(input)
	if decision.Outcome != "deny" {
		t.Errorf("expected deny with wrong source repo, got %s", decision.Outcome)
	}
}

func TestEvaluate_ScopeMatch(t *testing.T) {
	requiredScope := []string{"inventory.read", "inventory.write"}

	pe := NewPolicyEngine(&Policy{
		Version:   1,
		Issuers:   []string{"https://issuer.test"},
		Audiences: []string{"https://api.test"},
		Rules: []Rule{
			{Name: "scoped-rule", Match: MatchCriteria{Scope: &requiredScope}, Allow: true},
			{Name: "default", Allow: false},
		},
	})

	// Token has superset of required scope — should match (token ⊇ rule)
	input := &verify.PolicyInput{
		Scope: []string{"inventory.read", "inventory.write", "inventory.admin"},
	}
	decision := pe.Evaluate(input)
	if decision.Outcome != "allow" {
		t.Errorf("expected allow when token scope is superset, got %s", decision.Outcome)
	}

	// Token has subset of required scope — should NOT match
	input.Scope = []string{"inventory.read"}
	decision = pe.Evaluate(input)
	if decision.Outcome != "deny" {
		t.Errorf("expected deny when token scope is subset, got %s", decision.Outcome)
	}

	// Token has exact required scope — should match
	input.Scope = []string{"inventory.read", "inventory.write"}
	decision = pe.Evaluate(input)
	if decision.Outcome != "allow" {
		t.Errorf("expected allow when token scope is exact match, got %s", decision.Outcome)
	}
}

// ── Policy loading tests ────────────────────────────────────────────────────

func TestLoadPolicy_ValidJSON(t *testing.T) {
	policyJSON := `{
		"version": 1,
		"issuers": ["https://issuer.oathmesh.tech"],
		"audiences": ["https://inventory.internal"],
		"rules": [
			{
				"name": "allow-all",
				"match": {},
				"allow": true
			},
			{
				"name": "default",
				"match": {},
				"allow": false
			}
		]
	}`

	p, err := LoadPolicyFromJSON([]byte(policyJSON))
	if err != nil {
		t.Fatalf("expected valid policy, got error: %v", err)
	}

	if p.Version != 1 {
		t.Errorf("expected version 1, got %d", p.Version)
	}
	if len(p.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(p.Rules))
	}
}

func TestLoadPolicy_InvalidVersion(t *testing.T) {
	policyJSON := `{
		"version": 2,
		"issuers": ["https://issuer.test"],
		"audiences": ["https://api.test"],
		"rules": [{"name": "default", "allow": false}]
	}`

	_, err := LoadPolicyFromJSON([]byte(policyJSON))
	if err == nil {
		t.Fatal("expected error for invalid version, got nil")
	}
}

func TestLoadPolicy_MissingDefaultDeny(t *testing.T) {
	policyJSON := `{
		"version": 1,
		"issuers": ["https://issuer.test"],
		"audiences": ["https://api.test"],
		"rules": [{"name": "allow-all", "allow": true}]
	}`

	_, err := LoadPolicyFromJSON([]byte(policyJSON))
	if err == nil {
		t.Fatal("expected error for missing default deny, got nil")
	}
}

func TestLoadPolicy_EmptyIssuers(t *testing.T) {
	policyJSON := `{
		"version": 1,
		"issuers": [],
		"audiences": ["https://api.test"],
		"rules": [{"name": "default", "allow": false}]
	}`

	_, err := LoadPolicyFromJSON([]byte(policyJSON))
	if err == nil {
		t.Fatal("expected error for empty issuers, got nil")
	}
}

func TestLoadPolicy_FromFile(t *testing.T) {
	p := testPolicy()
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal policy: %v", err)
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-policy.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write temp policy: %v", err)
	}

	loaded, err := LoadPolicyFromFile(path)
	if err != nil {
		t.Fatalf("load policy from file: %v", err)
	}

	if len(loaded.Rules) != len(p.Rules) {
		t.Errorf("expected %d rules, got %d", len(p.Rules), len(loaded.Rules))
	}
}

// ── Hot-reload test ─────────────────────────────────────────────────────────

func TestWatchedPolicyEngine_HotReload(t *testing.T) {
	// Create initial policy
	initialPolicy := testPolicy()
	data, _ := json.Marshal(initialPolicy)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "policy.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write initial policy: %v", err)
	}

	wpe, err := NewWatchedPolicyEngine(path, nil)
	if err != nil {
		t.Fatalf("create watched engine: %v", err)
	}
	defer wpe.Close()

	// Verify initial policy works
	input := &verify.PolicyInput{Sub: "agent://repo/acme/deploy-bot", Act: "inventory.read"}
	decision := wpe.Evaluate(input)
	if decision.Outcome != "allow" {
		t.Fatalf("initial policy should allow, got %s", decision.Outcome)
	}

	// Write new policy that denies everything
	denyPolicy := &Policy{
		Version:   1,
		Issuers:   []string{"https://issuer.test"},
		Audiences: []string{"https://api.test"},
		Rules:     []Rule{{Name: "default", Allow: false}},
	}
	denyData, _ := json.Marshal(denyPolicy)
	if err := os.WriteFile(path, denyData, 0600); err != nil {
		t.Fatalf("write deny policy: %v", err)
	}

	// Wait for reload (fsnotify is async)
	time.Sleep(200 * time.Millisecond)

	// Verify new policy is applied
	decision = wpe.Evaluate(input)
	if decision.Outcome != "deny" {
		t.Errorf("after hot-reload, expected deny, got %s (rule: %s)", decision.Outcome, decision.RuleName)
	}
}

// ── ValidatePolicy tests ────────────────────────────────────────────────────

func TestValidatePolicy_Valid(t *testing.T) {
	err := ValidatePolicy(testPolicy())
	if err != nil {
		t.Errorf("expected valid policy, got: %v", err)
	}
}

func TestValidatePolicy_InvalidLastRule(t *testing.T) {
	p := testPolicy()
	p.Rules[len(p.Rules)-1].Allow = true // break the default deny rule
	err := ValidatePolicy(p)
	if err == nil {
		t.Error("expected error for last rule with allow=true")
	}
}
