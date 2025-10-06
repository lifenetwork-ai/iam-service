package ucases

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	cachingtypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	otpqueue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	smscommon "github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	services "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/services"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type courierUseCase struct {
	channelCache cachingtypes.CacheRepository
	queue        otpqueue.OTPQueueRepository
	defaultTTL   time.Duration
	smsProvider  services.SMSProvider
}

func NewCourierUseCase(
	queue otpqueue.OTPQueueRepository,
	smsProvider services.SMSProvider,
	channelCache cachingtypes.CacheRepository,
) interfaces.CourierUseCase {
	return &courierUseCase{
		queue:        queue,
		defaultTTL:   constants.DefaultChallengeDuration,
		smsProvider:  smsProvider,
		channelCache: channelCache,
	}
}

// ChooseChannel chooses the channel to send OTP to the receiver
func (u *courierUseCase) ChooseChannel(ctx context.Context, tenantName, receiver, channel string) *domainerrors.DomainError {
	key := &cachingtypes.Keyer{
		Raw: fmt.Sprintf("channel:%s:%s", tenantName, receiver),
	}

	// Get supported channels from tenant
	supportedChannels := u.GetAvailableChannels(ctx, tenantName, receiver)
	if !slices.Contains(supportedChannels, channel) {
		return domainerrors.NewValidationError("MSG_CHANNEL_NOT_SUPPORTED", "Channel not supported", []any{
			map[string]string{"channel": channel, "supported_channels": strings.Join(supportedChannels, ", ")},
		})
	}

	err := u.channelCache.SaveItem(key, channel, u.defaultTTL)
	if err != nil {
		return domainerrors.NewInternalError("MSG_CHOOSE_CHANNEL_FAILED", "Failed to choose channel to send OTP").WithCause(err)
	}

	return nil
}

func (u *courierUseCase) GetChannel(ctx context.Context, tenantName, receiver string) (types.ChooseChannelResponse, *domainerrors.DomainError) {
	key := &cachingtypes.Keyer{
		Raw: fmt.Sprintf("channel:%s:%s", tenantName, receiver),
	}

	var response string

	err := u.channelCache.RetrieveItem(key, &response)
	if err != nil {
		// fallback to webhooks channel if cache miss
		if errors.Is(err, cachingtypes.ErrCacheMiss) {
			return types.ChooseChannelResponse{
				Channel: constants.DefaultSMSChannel,
			}, nil
		}
		return types.ChooseChannelResponse{}, domainerrors.NewInternalError("MSG_GET_CHANNEL_FAILED", "Failed to get channel from cache").WithCause(err)
	}

	return types.ChooseChannelResponse{
		Channel: response,
	}, nil
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

	otp := smscommon.ExtractOTPFromMessage(body)
	if otp == "" {
		return domainerrors.NewValidationError(
			"MSG_INVALID_OTP",
			"Cannot extract OTP from body",
			[]any{"OTP must be 6 digits"},
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
		Message:    otp,
		TenantName: tenantName,
		CreatedAt:  time.Now(),
	}

	if err := u.queue.Enqueue(ctx, item, u.defaultTTL); err != nil {
		return domainerrors.NewInternalError("MSG_QUEUE_ENQUEUE_FAILED", "Failed to enqueue OTP").WithCause(err)
	}

	return nil
}

// https://geneticavietnam.slack.com/archives/C09DUSTLF1V/p1757998590863439?thread_ts=1757998451.351149&cid=C09DUSTLF1V
func (u *courierUseCase) GetAvailableChannels(ctx context.Context, tenantName, receiver string) []string {
	tn := strings.TrimSpace(strings.ToLower(tenantName))

	switch {
	case strings.EqualFold(tn, strings.ToLower(constants.TenantGenetica)):
		return []string{constants.ChannelSMS, constants.ChannelZalo, constants.ChannelWhatsApp}

	case strings.EqualFold(tn, strings.ToLower(constants.TenantLifeAI)):
		return []string{constants.ChannelSMS, constants.ChannelWhatsApp}

	default:
		return []string{constants.ChannelSMS, constants.ChannelWhatsApp, constants.ChannelZalo}
	}
}

func (u *courierUseCase) DeliverOTP(ctx context.Context, tenantName, receiver string) *domainerrors.DomainError {
	channel, usecaseErr := u.GetChannel(ctx, tenantName, receiver)
	if usecaseErr != nil {
		return usecaseErr
	}

	if channel.Channel == "" {
		return domainerrors.NewInternalError("MSG_CHANNEL_NOT_FOUND", "Channel not found").WithDetails([]any{
			map[string]string{"channel": channel.Channel},
		})
	}

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
	if err := u.smsProvider.SendOTP(ctx, tenantName, receiver, channel.Channel, item.Message, u.defaultTTL); err != nil {
		logger.GetLogger().Errorf("Failed to deliver OTP to %s via %s: %v", receiver, channel.Channel, err)
		// Prepare retry task
		retryTask := otpqueue.RetryTask{
			Receiver:   receiver,
			Message:    item.Message,
			Channel:    channel.Channel,
			TenantName: tenantName,
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

			// Try sending
			err := u.smsProvider.SendOTP(ctx, currentTask.TenantName, currentTask.Receiver, currentTask.Channel, currentTask.Message, u.defaultTTL)
			if err != nil {
				if currentTask.RetryCount < constants.MaxOTPRetryCount {
					// Retry again - do NOT delete
					_ = u.queue.EnqueueRetry(ctx, currentTask)
					return nil // skip deletion
				}

				// Exceeded max -> log + delete
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
