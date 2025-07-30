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

	// Always delete the OTP from queue — success or not
	defer func() {
		if delErr := u.queue.Delete(ctx, tenantName, receiver); delErr != nil {
			logger.GetLogger().Warnf("Failed to delete OTP from queue after attempt: %v", delErr)
		}
	}()

	// Attempt to send OTP
	if err := sendViaProvider(ctx, channel, receiver, item.Message); err != nil {
		// Prepare retry task
		retryTask := otpqueue.RetryTask{
			Receiver:   receiver,
			Message:    item.Message,
			Channel:    channel,
			TenantName: tenantName,
			RetryCount: 1,
			// ReadyAt will be computed inside EnqueueRetry
		}
		if err := u.queue.EnqueueRetry(ctx, retryTask); err != nil {
			return domainerrors.NewInternalError("MSG_RETRY_ENQUEUE_FAILED", "Failed to enqueue retry task").WithCause(err)
		}
		return domainerrors.NewInternalError("MSG_DELIVER_FAILED", "Failed to deliver OTP. Will retry later").WithCause(err)
	}

	return nil // delivery success
}

func (u *courierUseCase) RetryFailedOTPs(ctx context.Context, now time.Time) (int, *domainerrors.DomainError) {
	tasks, err := u.queue.GetDueRetryTasks(ctx, now)
	if err != nil {
		return 0, domainerrors.NewInternalError("MSG_GET_RETRY_TASKS_FAILED", "Failed to fetch retry tasks").WithCause(err)
	}

	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, constants.MaxConcurrency)

	for _, task := range tasks {
		currentTask := task

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			// Send OTP using retry task content directly
			logger.GetLogger().Infof("Retrying OTP to %s | Retry #%d", currentTask.Receiver, currentTask.RetryCount)
			if err := sendViaProvider(ctx, currentTask.Channel, currentTask.Receiver, currentTask.Message); err != nil {
				if currentTask.RetryCount < constants.MaxOTPRetryCount {
					// Retry again - do NOT delete
					_ = u.queue.EnqueueRetry(ctx, currentTask)
					return nil // skip deletion
				}

				// Exceeded max → log + delete
				logger.GetLogger().Warnf("Exceeded max retry count for %s | Retry #%d. Discarding.", currentTask.Receiver, currentTask.RetryCount)
			} else {
				// Delivered successfully
				logger.GetLogger().Infof("OTP delivered successfully to %s", currentTask.Receiver)
			}

			// Only delete if success OR exceeded retry
			if err := u.queue.DeleteRetryTask(ctx, currentTask); err != nil {
				logger.GetLogger().Warnf("Failed to delete retry task: %v", err)
			}

			// Clean up original OTP if exists (optional)
			_ = u.queue.Delete(ctx, currentTask.TenantName, currentTask.Receiver)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		logger.GetLogger().Errorf("Error retrying OTPs: %v", err)
	}

	return len(tasks), nil
}
