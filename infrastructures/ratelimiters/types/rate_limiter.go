package types

import (
	"context"
	"time"
)

type RateLimiter interface {
	IsLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	RegisterAttempt(ctx context.Context, key string, window time.Duration) error
	ResetAttempts(ctx context.Context, key string) error
}
