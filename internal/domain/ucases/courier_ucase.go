package ucases

import (
	"context"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	otpqueue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
)

type courierUseCase struct {
	queue      otpqueue.OTPQueueRepository
	defaultTTL time.Duration
}

func NewCourierUseCase(
	queue otpqueue.OTPQueueRepository,
) interfaces.CourierUseCase {
	return &courierUseCase{
		queue:      queue,
		defaultTTL: constants.DefaultChallengeDuration,
	}
}

func (u *courierUseCase) ReceiveOTP(ctx context.Context, receiver, body string) *domainerrors.DomainError {
	tenantName := extractTenantNameFromBody(body)
	if tenantName == "" {
		return domainerrors.NewValidationError(
			"MSG_INVALID_TENANT",
			"Cannot extract tenant from body",
			[]any{"Tenant name must be life_ai or genetica"},
		)
	}

	if tenantName != constants.TenantLifeAI && tenantName != constants.TenantGenetica {
		return domainerrors.NewValidationError(
			"MSG_INVALID_TENANT",
			"Invalid tenant name",
			[]any{"Tenant name must be life_ai or genetica"},
		)
	}

	item := otpqueue.OTPQueueItem{
		ID:         uuid.New().String(),
		Receiver:   receiver,
		Message:    body,
		TenantName: tenantName,
		CreatedAt:  time.Now(),
	}

	if err := u.queue.Enqueue(ctx, item, u.defaultTTL); err != nil {
		return domainerrors.NewInternalError("MSG_QUEUE_ENQUEUE_FAILED", "Failed to enqueue OTP").WithCause(err)
	}

	return nil
}

func (u *courierUseCase) GetAvailableChannels(ctx context.Context, tenantName, receiver string) []string {
	var channels []string

	// Always supported SMS and WhatsApp channels
	channels = append(channels, constants.ChannelSMS, constants.ChannelWhatsApp)

	// If the receiver is a Vietnamese number and the tenant supports Zalo, add Zalo
	if strings.HasPrefix(receiver, "+84") && strings.ToLower(tenantName) == constants.TenantGenetica {
		channels = append(channels, constants.ChannelZalo)
	}

	return channels
}

func (u *courierUseCase) DeliverOTP(ctx context.Context, tenantName, receiver, channel string) *domainerrors.DomainError {
	// Get OTP from queue
	item, err := u.queue.Get(ctx, tenantName, receiver)
	if err != nil {
		return domainerrors.NewInternalError("MSG_GET_OTP_FAILED", "Failed to get OTP from queue").WithCause(err)
	}

	// Send OTP via the corresponding provider
	if err := sendViaProvider(ctx, channel, receiver, item.Message); err != nil {
		delay := utils.ComputeBackoffDuration(1)
		retryTask := otpqueue.RetryTask{
			Receiver:   receiver,
			Message:    item.Message,
			Channel:    channel,
			TenantName: tenantName,
			RetryCount: 1,
			ReadyAt:    time.Now().Add(delay),
		}
		if err := u.queue.EnqueueRetry(ctx, retryTask, delay); err != nil {
			return domainerrors.NewInternalError("MSG_RETRY_ENQUEUE_FAILED", "Failed to enqueue retry task").WithCause(err)
		}
		return domainerrors.NewInternalError("MSG_DELIVER_FAILED", "Failed to deliver OTP. Will retry later").WithCause(err)
	}

	// Send success => delete OTP from queue
	if err := u.queue.Delete(ctx, tenantName, receiver); err != nil {
		return domainerrors.NewInternalError("MSG_DELETE_OTP_FAILED", "Failed to delete OTP after successful delivery").WithCause(err)
	}

	return nil
}

func (u *courierUseCase) RetryFailedOTPs(ctx context.Context, now time.Time) (int, *domainerrors.DomainError) {
	tasks, err := u.queue.GetDueRetryTasks(ctx, now)
	if err != nil {
		return 0, domainerrors.NewInternalError("MSG_GET_RETRY_TASKS_FAILED", "Failed to fetch retry tasks").WithCause(err)
	}

	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, constants.MaxConcurrency)

	for _, task := range tasks {
		currentTask := task // capture task is necessary for goroutine safety

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			// Check OTP still valid before retrying
			_, getErr := u.queue.Get(ctx, currentTask.TenantName, currentTask.Receiver)
			if getErr != nil {
				// OTP expired → skip and clean up retry task
				_ = u.queue.DeleteRetryTask(ctx, currentTask)
				return nil
			}

			// Try sending
			err := sendViaProvider(ctx, currentTask.Channel, currentTask.Receiver, currentTask.Message)
			if err != nil {
				if currentTask.RetryCount < constants.MaxOTPRetryCount {
					backoffDelay := utils.ComputeBackoffDuration(currentTask.RetryCount + 1)
					lastTime := currentTask.ReadyAt
					if lastTime.Before(time.Now()) {
						lastTime = time.Now()
					}
					currentTask.ReadyAt = lastTime.Add(backoffDelay)
					_ = u.queue.EnqueueRetry(ctx, currentTask, backoffDelay)
				}
				_ = u.queue.DeleteRetryTask(ctx, currentTask)
				return nil
			}

			// Success: cleanup OTP and retry task
			_ = u.queue.DeleteRetryTask(ctx, currentTask)
			_ = u.queue.Delete(ctx, currentTask.TenantName, currentTask.Receiver)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		logger.GetLogger().Errorf("Error retrying OTPs: %v", err)
	}

	return len(tasks), nil
}
