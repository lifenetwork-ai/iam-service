package otp_queue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/patrickmn/go-cache"
)

const (
	pendingOTPKeyPrefix = "otp:pending:" // otp:pending:<tenant>:<receiver>
	retryOTPKeyPrefix   = "otp:retry:"   // otp:retry:<tenant>:<receiver>
)

type memoryOTPQueue struct {
	cache *cache.Cache
}

func NewMemoryOTPQueueRepository(c *cache.Cache) types.OTPQueueRepository {
	return &memoryOTPQueue{
		cache: c,
	}
}

// Helper
func pendingOTPKey(tenantName, receiver string) string {
	return fmt.Sprintf("%s%s:%s", pendingOTPKeyPrefix, tenantName, receiver)
}

func retryOTPKey(tenantName, receiver string) string {
	return fmt.Sprintf("%s%s:%s", retryOTPKeyPrefix, tenantName, receiver)
}

// Enqueue OTP
func (q *memoryOTPQueue) Enqueue(ctx context.Context, item types.OTPQueueItem, ttl time.Duration) error {
	key := pendingOTPKey(item.TenantName, item.Receiver)
	q.cache.Set(key, item, ttl)
	return nil
}

func (q *memoryOTPQueue) Get(ctx context.Context, tenantName, receiver string) (*types.OTPQueueItem, error) {
	key := pendingOTPKey(tenantName, receiver)
	val, found := q.cache.Get(key)
	if !found {
		return nil, fmt.Errorf("OTP not found for %s", receiver)
	}
	item := val.(types.OTPQueueItem)
	return &item, nil
}

func (q *memoryOTPQueue) Delete(ctx context.Context, tenantName, receiver string) error {
	key := pendingOTPKey(tenantName, receiver)
	q.cache.Delete(key)
	return nil
}

// Retry Tasks
func (q *memoryOTPQueue) EnqueueRetry(ctx context.Context, task types.RetryTask, delay time.Duration) error {
	key := retryOTPKey(task.TenantName, task.Receiver)
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
	key := retryOTPKey(task.TenantName, task.Receiver)
	q.cache.Delete(key)
	return nil
}
