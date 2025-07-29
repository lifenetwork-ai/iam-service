package otp_queue_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue"
	"github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
)

// Create a new in-memory OTP queue instance for testing
func newTestQueue() types.OTPQueueRepository {
	return otp_queue.NewMemoryOTPQueueRepository(cache.New(5*time.Minute, 10*time.Minute))
}

// Test OTP enqueue → get → delete using table-driven tests
func TestMemoryOTPQueue_PendingOTP(t *testing.T) {
	tests := []struct {
		name  string
		input types.OTPQueueItem
		ttl   time.Duration
	}{
		{
			name: "basic enqueue and get",
			input: types.OTPQueueItem{
				ID:         "otp-001",
				TenantName: "test_tenant",
				Receiver:   "user@example.com",
				Message:    "Your OTP is 123456",
				CreatedAt:  time.Now(),
			},
			ttl: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newTestQueue()
			ctx := context.Background()

			// Enqueue
			err := q.Enqueue(ctx, tt.input, tt.ttl)
			require.NoError(t, err)

			// Get and verify
			got, err := q.Get(ctx, tt.input.TenantName, tt.input.Receiver)
			require.NoError(t, err)
			require.Equal(t, tt.input.ID, got.ID)
			require.Equal(t, tt.input.Receiver, got.Receiver)
			require.Equal(t, tt.input.Message, got.Message)
			require.Equal(t, tt.input.TenantName, got.TenantName)
			require.WithinDuration(t, tt.input.CreatedAt, got.CreatedAt, time.Second)

			// Delete
			err = q.Delete(ctx, tt.input.TenantName, tt.input.Receiver)
			require.NoError(t, err)

			// Ensure deletion
			got, err = q.Get(ctx, tt.input.TenantName, tt.input.Receiver)
			require.Error(t, err)
			require.Nil(t, got)
		})
	}
}

// Test enqueueing retry tasks and getting them when due
func TestMemoryOTPQueue_RetryTask(t *testing.T) {
	tests := []struct {
		name      string
		task      types.RetryTask
		retryWait time.Duration
	}{
		{
			name: "basic retry task with increasing count",
			task: types.RetryTask{
				TenantName: "test_tenant",
				Receiver:   "user@example.com",
				Channel:    "email",
				Message:    "Your OTP is 654321",
			},
			retryWait: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newTestQueue()
			ctx := context.Background()

			// Enqueue 1st retry
			err := q.EnqueueRetry(ctx, tt.task, tt.retryWait)
			require.NoError(t, err)

			// Ensure it's not ready yet
			tasks, err := q.GetDueRetryTasks(ctx, time.Now().Add(-1*time.Second))
			require.NoError(t, err)
			require.Len(t, tasks, 0)

			// Wait until due (add buffer to ensure it's ready)
			time.Sleep(tt.retryWait + 10*time.Millisecond)

			tasks, err = q.GetDueRetryTasks(ctx, time.Now())
			require.NoError(t, err)
			require.Len(t, tasks, 1)
			require.Equal(t, 1, tasks[0].RetryCount)

			// Enqueue again → retry count increases
			err = q.EnqueueRetry(ctx, tt.task, tt.retryWait)
			require.NoError(t, err)

			time.Sleep(tt.retryWait + 10*time.Millisecond)

			tasks, err = q.GetDueRetryTasks(ctx, time.Now())
			require.NoError(t, err)
			require.Len(t, tasks, 1)
			require.Equal(t, 2, tasks[0].RetryCount)

			// Delete retry task
			err = q.DeleteRetryTask(ctx, tt.task)
			require.NoError(t, err)

			tasks, err = q.GetDueRetryTasks(ctx, time.Now())
			require.NoError(t, err)
			require.Empty(t, tasks)
		})
	}
}

func TestMemoryOTPQueue_RetryTask_PersistsUpdatedRetryCount(t *testing.T) {
	q := newTestQueue()
	ctx := context.Background()

	task := types.RetryTask{
		TenantName: "test_tenant",
		Receiver:   "persist@example.com",
		Channel:    "email",
		Message:    "Test OTP",
		RetryCount: 1,
		ReadyAt:    time.Now().Add(200 * time.Millisecond),
	}

	// Step 1: Enqueue initial retry
	err := q.EnqueueRetry(ctx, task, 200*time.Millisecond)
	require.NoError(t, err)

	// Wait until ready
	time.Sleep(250 * time.Millisecond)

	// Step 2: Get task and confirm RetryCount = 1
	tasks, err := q.GetDueRetryTasks(ctx, time.Now())
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.Equal(t, 1, tasks[0].RetryCount)

	// Step 3: Increase RetryCount and re-enqueue
	tasks[0].RetryCount++
	newDelay := 1 * time.Second
	tasks[0].ReadyAt = time.Now().Add(newDelay)
	err = q.EnqueueRetry(ctx, tasks[0], newDelay)
	require.NoError(t, err)

	// Wait again until ready
	time.Sleep(newDelay + 250*time.Millisecond)

	// Step 4: Get task again and confirm RetryCount = 2
	tasks, err = q.GetDueRetryTasks(ctx, time.Now())
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.Equal(t, 2, tasks[0].RetryCount)
}

// Test listing receivers with pending OTPs for a given tenant
func TestMemoryOTPQueue_ListReceivers(t *testing.T) {
	tests := []struct {
		name           string
		items          []types.OTPQueueItem
		tenant         string
		expectedEmails []string
	}{
		{
			name: "tenant1 with two receivers",
			items: []types.OTPQueueItem{
				{ID: "1", TenantName: "tenant1", Receiver: "a@example.com", Message: "OTP 111", CreatedAt: time.Now()},
				{ID: "2", TenantName: "tenant1", Receiver: "b@example.com", Message: "OTP 222", CreatedAt: time.Now()},
				{ID: "3", TenantName: "tenant2", Receiver: "c@example.com", Message: "OTP 333", CreatedAt: time.Now()},
			},
			tenant:         "tenant1",
			expectedEmails: []string{"a@example.com", "b@example.com"},
		},
		{
			name: "tenant2 with one receiver",
			items: []types.OTPQueueItem{
				{ID: "4", TenantName: "tenant2", Receiver: "d@example.com", Message: "OTP 444", CreatedAt: time.Now()},
			},
			tenant:         "tenant2",
			expectedEmails: []string{"d@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newTestQueue()
			ctx := context.Background()

			for _, item := range tt.items {
				err := q.Enqueue(ctx, item, 5*time.Minute)
				require.NoError(t, err)
			}

			receivers, err := q.ListReceivers(ctx, tt.tenant)
			require.NoError(t, err)
			require.ElementsMatch(t, receivers, tt.expectedEmails)
		})
	}
}

func TestMemoryOTPQueue_PerformanceUnderLoad(t *testing.T) {
	const totalOps = 25000
	const tenant = "perfTenant"
	ctx := context.Background()
	q := newTestQueue()

	start := time.Now()

	// Enqueue test
	for i := range totalOps {
		item := types.OTPQueueItem{
			ID:         fmt.Sprintf("id-%d", i),
			TenantName: tenant,
			Receiver:   fmt.Sprintf("user%d@example.com", i),
			Message:    "otp",
			CreatedAt:  time.Now(),
		}
		err := q.Enqueue(ctx, item, 2*time.Minute)
		require.NoError(t, err)
	}
	enqueueDuration := time.Since(start)
	t.Logf("[PERF] Enqueued %d items in %v (avg: %v/op)", totalOps, enqueueDuration, enqueueDuration/time.Duration(totalOps))

	// Get test
	start = time.Now()
	for i := range totalOps {
		receiver := fmt.Sprintf("user%d@example.com", i)
		_, err := q.Get(ctx, tenant, receiver)
		require.NoError(t, err)
	}
	getDuration := time.Since(start)
	t.Logf("[PERF] Retrieved %d items in %v (avg: %v/op)", totalOps, getDuration, getDuration/time.Duration(getDuration.Nanoseconds()/int64(totalOps)))
}

func TestMemoryOTPQueue_Performance_RetryTasks(t *testing.T) {
	const totalTasks = 25000
	const tenant = "retryPerfTenant"
	ctx := context.Background()
	q := newTestQueue()

	// Enqueue retry tasks with slight delay
	start := time.Now()
	for i := range totalTasks {
		task := types.RetryTask{
			TenantName: tenant,
			Receiver:   fmt.Sprintf("user%d@example.com", i),
			Channel:    "sms",
			Message:    "retry-otp",
			RetryCount: 1,
			ReadyAt:    time.Now().Add(500 * time.Millisecond), // due soon
		}
		err := q.EnqueueRetry(ctx, task, 500*time.Millisecond)
		require.NoError(t, err)
	}
	enqueueDur := time.Since(start)
	t.Logf("[PERF] Enqueued %d retry tasks in %v (avg: %v/task)", totalTasks, enqueueDur, enqueueDur/time.Duration(totalTasks))

	// Wait for tasks to be due
	time.Sleep(600 * time.Millisecond)

	// Fetch due retry tasks
	start = time.Now()
	tasks, err := q.GetDueRetryTasks(ctx, time.Now())
	require.NoError(t, err)
	require.Len(t, tasks, totalTasks)

	fetchDur := time.Since(start)
	t.Logf("[PERF] Retrieved %d retry tasks in %v (avg: %v/task)", totalTasks, fetchDur, fetchDur/time.Duration(totalTasks))
}
