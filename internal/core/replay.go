package core

import (
	"context"
	"time"
)

type ReplayCache interface {
	Check(ctx context.Context, jti string, ttl time.Duration) (bool, error)
}
