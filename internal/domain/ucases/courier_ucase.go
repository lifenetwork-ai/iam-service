package ucases

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	otpqueue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
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
		delay := 30 * time.Second
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

func (u *courierUseCase) RetryFailedOTPs(ctx context.Context, now time.Time) *domainerrors.DomainError {
	return nil
}
