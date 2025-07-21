package otpqueue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lifenetwork-ai/iam-service/infrastructures/otpqueue/types"
	"github.com/patrickmn/go-cache"
)

var (
	pendingOTPKeyPrefix = "otp:pending:"
	retryOTPKeyPrefix   = "otp:retry:"
)

type memoryOTPQueue struct {
	cache *cache.Cache
}

func NewMemoryOTPQueueRepository(c *cache.Cache) types.OTPQueueRepository {
	return &memoryOTPQueue{cache: c}
}

func (q *memoryOTPQueue) Enqueue(ctx context.Context, item types.OTPQueueItem, ttl time.Duration) error {
	key := pendingOTPKeyPrefix + item.Receiver
	q.cache.Set(key, item, ttl)
	return nil
}

func (q *memoryOTPQueue) Get(ctx context.Context, receiver string) (*types.OTPQueueItem, error) {
	key := pendingOTPKeyPrefix + receiver
	val, found := q.cache.Get(key)
	if !found {
		return nil, fmt.Errorf("not found")
	}
	item := val.(types.OTPQueueItem)
	return &item, nil
}

func (q *memoryOTPQueue) Delete(ctx context.Context, receiver string) error {
	key := pendingOTPKeyPrefix + receiver
	q.cache.Delete(key)
	return nil
}

func (q *memoryOTPQueue) EnqueueRetry(ctx context.Context, task types.RetryTask, delay time.Duration) error {
	key := retryOTPKeyPrefix + task.Receiver
	task.ReadyAt = time.Now().Add(delay)
	q.cache.Set(key, task, delay+1*time.Minute)
	return nil
}

func (q *memoryOTPQueue) GetDueRetryTasks(ctx context.Context, now time.Time) ([]types.RetryTask, error) {
	var tasks []types.RetryTask
	for k, v := range q.cache.Items() {
		if strings.HasPrefix(k, retryOTPKeyPrefix) {
			task := v.Object.(types.RetryTask)
			if now.After(task.ReadyAt) {
				tasks = append(tasks, task)
			}
		}
	}
	return tasks, nil
}

func (q *memoryOTPQueue) DeleteRetryTask(ctx context.Context, task types.RetryTask) error {
	key := retryOTPKeyPrefix + task.Receiver
	q.cache.Delete(key)
	return nil
}
