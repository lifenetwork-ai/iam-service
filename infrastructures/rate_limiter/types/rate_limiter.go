package types

import (
	"time"
)

type RateLimiter interface {
	IsLimited(key string, limit int, window time.Duration) (bool, error)
	RegisterAttempt(key string, window time.Duration) error
	ResetAttempts(key string) error
}
