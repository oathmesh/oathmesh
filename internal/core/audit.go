package core

import (
	"context"
	"time"
)

type AuditEvent struct {
	Event     string    `json:"event"`
	Outcome   string    `json:"outcome"`
	Reason    string    `json:"reason"`
	JTI       string    `json:"jti"`
	Sub       string    `json:"sub"`
	Aud       string    `json:"aud"`
	Act       string    `json:"act"`
	Iss       string    `json:"iss"`
	Env       string    `json:"env,omitempty"`
	Source    *Source   `json:"src,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

type AuditSink interface {
	Emit(ctx context.Context, event *AuditEvent) error
}
