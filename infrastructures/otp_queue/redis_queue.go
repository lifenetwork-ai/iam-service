package otp_queue

import (
	"context"
	"encoding/json"
	"fmt"
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

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal retry task: %w", err)
	}

	score := float64(time.Now().Add(delay).Unix())
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
