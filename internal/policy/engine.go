package policy

import (
	"fmt"
	"strings"
	"sync"

	"github.com/oathmesh/oathmesh/internal/core"
	"github.com/oathmesh/oathmesh/internal/verify"
)



// PolicyEvaluator evaluates OathMesh policy rules against verified token claims.
// Thread-safe for concurrent use — policy is protected by sync.RWMutex for hot-reload.
type PolicyEngine struct {
	mu     sync.RWMutex
	policy *Policy
}

// NewPolicyEngine creates a policy engine from a loaded Policy.
func NewPolicyEngine(p *Policy) *PolicyEngine {
	return &PolicyEngine{
		policy: p,
	}
}

// Evaluate checks the given claims against the loaded policy rules.
// Rules are evaluated in order; first match wins. No match = deny (default deny enforced).
//
// OathMesh authenticates the caller. The receiver authorizes the request.
func (pe *PolicyEngine) Evaluate(claims *verify.PolicyInput) *verify.PolicyDecision {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	if pe.policy == nil || len(pe.policy.Rules) == 0 {
		return &verify.PolicyDecision{Outcome: "deny", RuleName: "default"}
	}

	for _, rule := range pe.policy.Rules {
		if matchesRule(&rule, claims) {
			outcome := "deny"
			if rule.Allow {
				outcome = "allow"
			}
			return &verify.PolicyDecision{Outcome: outcome, RuleName: rule.Name}
		}
	}

	// Default deny — no rule matched. This should not normally be reached
	// because valid policies always have a default deny rule as the last entry,
	// but we enforce it in the evaluator as a safety net.
	return &verify.PolicyDecision{Outcome: "deny", RuleName: "default"}
}

// SwapPolicy atomically replaces the loaded policy (used by hot-reload).
func (pe *PolicyEngine) SwapPolicy(p *Policy) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.policy = p
}

// GetPolicy returns the current loaded policy (used for inspection).
func (pe *PolicyEngine) GetPolicy() *Policy {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.policy
}

// PolicyInput holds the claims needed for policy evaluation.
// This is separate from sign.Claims to avoid a circular dependency.
type PolicyInput struct {
	Iss      string
	Sub      string
	Aud      string
	Act      string
	Scope    []string
	Env      string
	SrcType  string
	SrcRepo  string
	SrcWflow string
}

// matchesRule checks if a set of claims matches a single rule's criteria.
// All specified fields must match (AND logic). Unspecified fields are wildcards.
func matchesRule(rule *Rule, input *verify.PolicyInput) bool {
	m := &rule.Match

	// Sub: URI glob match
	if m.Sub != nil && !globMatch(*m.Sub, input.Sub) {
		return false
	}

	// Act: exact match (or glob if pattern contains *)
	if m.Act != nil && !globMatch(*m.Act, input.Act) {
		return false
	}

	// Env: exact match
	if m.Env != nil && *m.Env != input.Env {
		return false
	}

	// Tenant: exact match
	if m.Tenant != nil && *m.Tenant != input.Tenant {
		return false
	}

	// Scope: all listed scope values must be present in token (token ⊇ rule)
	if m.Scope != nil {
		for _, required := range *m.Scope {
			if !containsString(input.Scope, required) {
				return false
			}
		}
	}

	// Source criteria
	if m.Src != nil {
		if m.Src.Type != nil && !globMatch(*m.Src.Type, input.SrcType) {
			return false
		}
		if m.Src.Repo != nil && !globMatch(*m.Src.Repo, input.SrcRepo) {
			return false
		}
		if m.Src.Workflow != nil && !globMatch(*m.Src.Workflow, input.SrcWflow) {
			return false
		}
	}

	return true
}

// globMatch performs flat glob matching on the entire string.
// Unlike path.Match and filepath.Match, this does NOT treat "/" as a special
// path separator boundary. This is critical for URI patterns like
// "agent://repo/acme/*" which must match "agent://repo/acme/deploy-bot".
//
// Supported patterns:
//   - "*"       → matches everything
//   - "foo*"    → prefix match (strings.HasPrefix)
//   - "*foo"    → suffix match (strings.HasSuffix)
//   - "fo*bar"  → prefix + suffix match with wildcard in middle
//   - "exact"   → exact string match (no wildcard)
//
// Only single "*" wildcard is supported. This is intentionally simple.
func globMatch(pattern, value string) bool {
	// No wildcard: exact match
	if !strings.Contains(pattern, "*") {
		return pattern == value
	}

	// Single "*" matches everything
	if pattern == "*" {
		return true
	}

	// Split on "*" — we support exactly one wildcard
	parts := strings.SplitN(pattern, "*", 2)
	prefix := parts[0]
	suffix := parts[1]

	if len(value) < len(prefix)+len(suffix) {
		return false
	}

	return strings.HasPrefix(value, prefix) && strings.HasSuffix(value, suffix)
}

// containsString checks if a slice contains a specific string.
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// PolicyInputFromContext builds a verify.PolicyInput from a VerifiedCallerContext.
// Used to bridge between the verify package and policy evaluation.
func PolicyInputFromContext(vcc *core.VerifiedCallerContext) *verify.PolicyInput {
	input := &verify.PolicyInput{
		Iss:   vcc.Principal.Issuer,
		Sub:   vcc.Principal.Subject,
		Act:   vcc.Action,
		Scope: vcc.Scope,
		Env:   vcc.Env,
	}
	if vcc.Source != nil {
		input.SrcType = vcc.Source.Type
		input.SrcRepo = vcc.Source.Repo
		input.SrcWflow = vcc.Source.Workflow
	}
	return input
}

// Compile-time check that PolicyEngine implements verify.PolicyEvaluator.
var _ verify.PolicyEvaluator = (*PolicyEngine)(nil)

// ValidatePolicy checks that a policy conforms to OathMesh requirements.
func ValidatePolicy(p *Policy) error {
	if p.Version != 1 {
		return fmt.Errorf("unsupported policy version: %d (expected 1)", p.Version)
	}
	if len(p.Issuers) == 0 {
		return fmt.Errorf("policy must have at least one issuer")
	}
	if len(p.Audiences) == 0 {
		return fmt.Errorf("policy must have at least one audience")
	}
	if len(p.Rules) == 0 {
		return fmt.Errorf("policy must have at least one rule")
	}
	lastRule := p.Rules[len(p.Rules)-1]
	if lastRule.Name != "default" || lastRule.Allow {
		return fmt.Errorf("last rule must be { name: \"default\", allow: false } (default deny)")
	}
	return nil
}
