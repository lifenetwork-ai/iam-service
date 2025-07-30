package otp_queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
	"github.com/redis/go-redis/v9"
)

const (
	retryZSetKey    = "otp:retry_tasks"
	retryTaskMapKey = "otp:retry_map" // This is a Redis Hash: field = taskKey, value = json
)

// Helper to generate a unique task key (used as ZSET member and HASH field)
func makeTaskKey(task types.RetryTask) string {
	return fmt.Sprintf("%s:%s:%s", task.TenantName, task.Receiver, task.Channel)
}

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

// EnqueueRetry adds a retry task to the queue
func (r *redisOTPQueue) EnqueueRetry(ctx context.Context, task types.RetryTask) error {
	taskKey := makeTaskKey(task)

	// Try to get existing task to increase RetryCount
	raw, err := r.client.HGet(ctx, retryTaskMapKey, taskKey).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get existing retry task: %w", err)
	}

	if err == nil {
		var existing types.RetryTask
		if err := json.Unmarshal([]byte(raw), &existing); err == nil {
			task.RetryCount = existing.RetryCount + 1
		}
	}

	if task.RetryCount == 0 {
		task.RetryCount = 1
	}

	// Compute backoff delay based on RetryCount
	delay := utils.ComputeBackoffDuration(task.RetryCount)

	// Set ReadyAt to current time + backoff delay
	task.ReadyAt = time.Now().Add(delay)

	// Marshal task
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal retry task: %w", err)
	}

	logger.GetLogger().Infof(
		"[EnqueueRetry] Saving retry task for %s | Retry #%d | Delay = %s | ReadyAt = %s",
		task.Receiver,
		task.RetryCount,
		delay,
		task.ReadyAt.Format(time.RFC3339),
	)

	// Store in sorted set (by ReadyAt timestamp) and hash map
	score := float64(task.ReadyAt.Unix())
	pipe := r.client.TxPipeline()
	pipe.ZAdd(ctx, retryZSetKey, redis.Z{
		Score:  score,
		Member: taskKey,
	})
	pipe.HSet(ctx, retryTaskMapKey, taskKey, data)

	_, err = pipe.Exec(ctx)
	return err
}

// GetDueRetryTasks retrieves all retry tasks that are due
func (r *redisOTPQueue) GetDueRetryTasks(ctx context.Context, now time.Time) ([]types.RetryTask, error) {
	// Get due task keys from ZSET
	taskKeys, err := r.client.ZRangeByScore(ctx, retryZSetKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%d", now.Unix()),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get retry task keys: %w", err)
	}
	if len(taskKeys) == 0 {
		return nil, nil
	}

	// Fetch JSON payloads from hash
	taskJSONs, err := r.client.HMGet(ctx, retryTaskMapKey, taskKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch retry tasks from hash: %w", err)
	}

	var tasks []types.RetryTask
	for _, raw := range taskJSONs {
		if rawStr, ok := raw.(string); ok {
			var t types.RetryTask
			if err := json.Unmarshal([]byte(rawStr), &t); err == nil {
				tasks = append(tasks, t)
			}
		}
	}

	return tasks, nil
}

func (r *redisOTPQueue) DeleteRetryTask(ctx context.Context, task types.RetryTask) error {
	taskKey := makeTaskKey(task)

	pipe := r.client.TxPipeline()
	pipe.ZRem(ctx, retryZSetKey, taskKey)
	pipe.HDel(ctx, retryTaskMapKey, taskKey)
	_, err := pipe.Exec(ctx)
	return err
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
