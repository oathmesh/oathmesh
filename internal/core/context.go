package core

import (
	"time"
)

type VerifiedCallerContext struct {
	Principal Principal
	Action    string
	Scope     []string
	Reason    string
	Source    *Source
	TokenID   string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Env       string
}

type Principal struct {
	Issuer  string
	Subject string
}

type Source struct {
	Type     string
	Repo     string
	Workflow string
	RunID    string
	SHA      string
}
