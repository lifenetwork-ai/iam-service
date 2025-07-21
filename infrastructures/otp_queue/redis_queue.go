package otp_queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/redis/go-redis/v9"
)

type redisOTPQueue struct {
	client *redis.Client
}

func NewRedisOTPQueueRepository(client *redis.Client) types.OTPQueueRepository {
	return &redisOTPQueue{
		client: client,
	}
}

// Enqueue OTP
func (r *redisOTPQueue) Enqueue(ctx context.Context, item types.OTPQueueItem, ttl time.Duration) error {
	key := pendingOTPKey(item.TenantName, item.Receiver)

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP item: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisOTPQueue) Get(ctx context.Context, tenantName, receiver string) (*types.OTPQueueItem, error) {
	key := pendingOTPKey(tenantName, receiver)

	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("OTP not found for %s", receiver)
	} else if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	var item types.OTPQueueItem
	if err := json.Unmarshal([]byte(data), &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP item: %w", err)
	}

	return &item, nil
}

func (r *redisOTPQueue) Delete(ctx context.Context, tenantName, receiver string) error {
	key := pendingOTPKey(tenantName, receiver)
	return r.client.Del(ctx, key).Err()
}

func (r *redisOTPQueue) EnqueueRetry(ctx context.Context, task types.RetryTask, delay time.Duration) error {
	key := retryOTPKey(task.TenantName, task.Receiver)
	// Get existing tasks to check for duplicates
	existingTasks, err := r.client.ZRange(ctx, key, 0, -1).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to fetch retry tasks: %w", err)
	}

	for _, raw := range existingTasks {
		var existing types.RetryTask
		if err := json.Unmarshal([]byte(raw), &existing); err == nil {
			if existing.Receiver == task.Receiver && existing.TenantName == task.TenantName {
				task.RetryCount = existing.RetryCount + 1
				// Remove the old task
				_ = r.client.ZRem(ctx, key, raw).Err()
				break
			}
		}
	}

	if task.RetryCount == 0 {
		task.RetryCount = 1
	}

	// Calculate the score based on the ready time
	score := float64(time.Now().Add(delay).Unix())
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal retry task: %w", err)
	}

	return r.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}

// Retry Tasks
func (r *redisOTPQueue) GetDueRetryTasks(ctx context.Context, now time.Time) ([]types.RetryTask, error) {
	// Lấy tất cả key retry (toàn bộ tenants)
	keys, err := r.client.Keys(ctx, retryOTPKeyPrefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list retry keys: %w", err)
	}

	var tasks []types.RetryTask

	for _, key := range keys {
		entries, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: "0",
			Max: fmt.Sprintf("%d", now.Unix()),
		}).Result()

		if err != nil && err != redis.Nil {
			continue
		}

		for _, raw := range entries {
			var task types.RetryTask
			if err := json.Unmarshal([]byte(raw), &task); err == nil {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

func (r *redisOTPQueue) DeleteRetryTask(ctx context.Context, task types.RetryTask) error {
	key := retryOTPKey(task.TenantName, task.Receiver)

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal retry task: %w", err)
	}

	return r.client.ZRem(ctx, key, data).Err()
}

// ListReceivers returns all receiver IDs that have pending OTPs for a given tenant
func (r *redisOTPQueue) ListReceivers(ctx context.Context, tenantName string) ([]string, error) {
	var receivers []string
	prefix := pendingOTPKeyPrefix + tenantName + ":"

	iter := r.client.Scan(ctx, 0, prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// key should be: otp:pending:<tenant>:<receiver>
		parts := strings.SplitN(key, ":", 4)
		if len(parts) == 4 && parts[0] == "otp" && parts[1] == "pending" && parts[2] == tenantName {
			receivers = append(receivers, parts[3])
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan redis keys: %w", err)
	}

	return receivers, nil
}
