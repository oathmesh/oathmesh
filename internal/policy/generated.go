package policy

// This file is a placeholder for pkl-gen-go output.
// Run `make pkl-gen` to regenerate from policy/schema.pkl
// after installing pkl-gen-go.

// Policy represents the root policy document
type Policy struct {
	Version   int      `pkl:"version"`
	Issuers   []string `pkl:"issuers"`
	Audiences []string `pkl:"audiences"`
	Rules     []Rule   `pkl:"rules"`
}

// Rule represents a single allow/deny rule
type Rule struct {
	Name  string        `pkl:"name"`
	Match MatchCriteria `pkl:"match"`
	Allow bool          `pkl:"allow"`
}

// MatchCriteria for a policy rule
type MatchCriteria struct {
	Sub   *string         `pkl:"sub"`
	Act   *string         `pkl:"act"`
	Scope *[]string       `pkl:"scope"`
	Env   *string         `pkl:"env"`
	Src   *SourceCriteria `pkl:"src"`
}

// SourceCriteria provenance match criteria
type SourceCriteria struct {
	Type     *string `pkl:"type"`
	Repo     *string `pkl:"repo"`
	Workflow *string `pkl:"workflow"`
}
