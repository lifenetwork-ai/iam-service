package otp_queue_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue"
	"github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisOTPQueueTestSuite struct {
	suite.Suite
	ctx            context.Context
	queue          types.OTPQueueRepository
	redisClient    *redis.Client
	redisContainer testcontainers.Container
}

func TestRedisOTPQueueTestSuite(t *testing.T) {
	suite.Run(t, new(RedisOTPQueueTestSuite))
}

func (s *RedisOTPQueueTestSuite) SetupSuite() {
	s.ctx = context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(s.T(), err)
	s.redisContainer = container

	port, _ := container.MappedPort(s.ctx, "6379")
	host, _ := container.Host(s.ctx)

	s.redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
		DB:   0,
	})
	s.queue = otp_queue.NewRedisOTPQueueRepository(s.redisClient)
}

func (s *RedisOTPQueueTestSuite) TearDownTest() {
	_ = s.redisClient.FlushDB(s.ctx).Err()
}

func (s *RedisOTPQueueTestSuite) TearDownSuite() {
	_ = s.redisContainer.Terminate(s.ctx)
}

// ------------------------- Tests ----------------------------

func (s *RedisOTPQueueTestSuite) Test_EnqueueGetDelete() {
	item := types.OTPQueueItem{
		ID:         "otp-001",
		TenantName: "tenantX",
		Receiver:   "alice@example.com",
		Message:    "Your OTP is 123456",
		CreatedAt:  time.Now(),
	}

	err := s.queue.Enqueue(s.ctx, item, 2*time.Second)
	require.NoError(s.T(), err)

	got, err := s.queue.Get(s.ctx, item.TenantName, item.Receiver)
	require.NoError(s.T(), err)
	require.Equal(s.T(), item.ID, got.ID)
	require.Equal(s.T(), item.Message, got.Message)

	err = s.queue.Delete(s.ctx, item.TenantName, item.Receiver)
	require.NoError(s.T(), err)

	got, err = s.queue.Get(s.ctx, item.TenantName, item.Receiver)
	require.Error(s.T(), err)
	require.Nil(s.T(), got)
}

func (s *RedisOTPQueueTestSuite) Test_EnqueueRetryTask() {
	task := types.RetryTask{
		TenantName: "tenantY",
		Receiver:   "bob@example.com",
		Channel:    "email",
		Message:    "retry OTP",
	}

	err := s.queue.EnqueueRetry(s.ctx, task, 1*time.Second)
	require.NoError(s.T(), err)

	time.Sleep(1100 * time.Millisecond)

	tasks, err := s.queue.GetDueRetryTasks(s.ctx, time.Now())
	require.NoError(s.T(), err)
	require.Len(s.T(), tasks, 1)
	require.Equal(s.T(), 1, tasks[0].RetryCount)

	// Enqueue again → should increment retry count
	err = s.queue.EnqueueRetry(s.ctx, task, 1*time.Second)
	require.NoError(s.T(), err)

	time.Sleep(1100 * time.Millisecond)

	tasks, err = s.queue.GetDueRetryTasks(s.ctx, time.Now())
	require.NoError(s.T(), err)
	require.Len(s.T(), tasks, 1)
	require.Equal(s.T(), 2, tasks[0].RetryCount)

	// Delete it
	err = s.queue.DeleteRetryTask(s.ctx, task)
	require.NoError(s.T(), err)

	tasks, err = s.queue.GetDueRetryTasks(s.ctx, time.Now())
	require.NoError(s.T(), err)
	require.Empty(s.T(), tasks)
}

func (s *RedisOTPQueueTestSuite) Test_ListReceivers() {
	items := []types.OTPQueueItem{
		{ID: "1", TenantName: "tenantZ", Receiver: "a@example.com", Message: "otpA", CreatedAt: time.Now()},
		{ID: "2", TenantName: "tenantZ", Receiver: "b@example.com", Message: "otpB", CreatedAt: time.Now()},
		{ID: "3", TenantName: "otherTenant", Receiver: "c@example.com", Message: "otpC", CreatedAt: time.Now()},
	}

	for _, item := range items {
		err := s.queue.Enqueue(s.ctx, item, 5*time.Minute)
		require.NoError(s.T(), err)
	}

	receivers, err := s.queue.ListReceivers(s.ctx, "tenantZ")
	require.NoError(s.T(), err)
	require.ElementsMatch(s.T(), []string{"a@example.com", "b@example.com"}, receivers)

	receivers, err = s.queue.ListReceivers(s.ctx, "otherTenant")
	require.NoError(s.T(), err)
	require.ElementsMatch(s.T(), []string{"c@example.com"}, receivers)
}
