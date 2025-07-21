package types

import (
	"context"
	"time"
)

// Worker is the interface that every background worker should implement
type Worker interface {
	Name() string
	Start(ctx context.Context, interval time.Duration)
}
