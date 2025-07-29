package otp_queue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"github.com/patrickmn/go-cache"
)

const (
	pendingOTPKeyPrefix = "otp:pending:" // otp:pending:<tenant>:<receiver>
	retryOTPKeyPrefix   = "otp:retry:"   // otp:retry:<tenant>:<receiver>
	retryTaskTTL        = 5 * time.Minute
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

	if task.RetryCount == 0 {
		if existing, found := q.cache.Get(key); found {
			if prevTask, ok := existing.(types.RetryTask); ok {
				task.RetryCount = prevTask.RetryCount + 1
			}
		} else {
			task.RetryCount = 1
		}
	}

	if task.ReadyAt.IsZero() {
		task.ReadyAt = time.Now().Add(delay)
	}

	// Save task
	q.cache.Set(key, task, retryTaskTTL)
	logger.GetLogger().Infof("[EnqueueRetry] Saving retry task for %s | Retry #%d | ReadyAt = %s",
		task.Receiver, task.RetryCount, task.ReadyAt.Format(time.RFC3339))
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

// ListReceivers returns all receiver IDs that have pending OTPs for a given tenant
func (q *memoryOTPQueue) ListReceivers(ctx context.Context, tenantName string) ([]string, error) {
	var receivers []string
	prefix := pendingOTPKeyPrefix + tenantName + ":"

	for k := range q.cache.Items() {
		if after, ok := strings.CutPrefix(k, prefix); ok {
			receivers = append(receivers, after)
		}
	}
	return receivers, nil
}
