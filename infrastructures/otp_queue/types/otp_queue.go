package types

import (
	"context"
	"time"
)

type OTPQueueItem struct {
	ID        string    `json:"id"`
	Receiver  string    `json:"receiver"`
	Message   string    `json:"message"`
	TenantID  string    `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
}

type RetryTask struct {
	Receiver   string    `json:"receiver"`
	Message    string    `json:"message"`
	Channel    string    `json:"channel"`
	TenantID   string    `json:"tenant_id"`
	RetryCount int       `json:"retry_count"`
	ReadyAt    time.Time `json:"ready_at"` // used for memory impl
}

type OTPQueueRepository interface {
	Enqueue(ctx context.Context, item OTPQueueItem, ttl time.Duration) error
	Get(ctx context.Context, receiver string) (*OTPQueueItem, error)
	Delete(ctx context.Context, receiver string) error

	EnqueueRetry(ctx context.Context, task RetryTask, delay time.Duration) error
	GetDueRetryTasks(ctx context.Context, now time.Time) ([]RetryTask, error)
	DeleteRetryTask(ctx context.Context, task RetryTask) error
}
