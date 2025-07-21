package workers

import (
	"context"
	"time"

	"github.com/lifenetwork-ai/iam-service/constants"
	otp_queue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	"github.com/lifenetwork-ai/iam-service/internal/workers/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type otpDeliveryWorker struct {
	curierUseCase interfaces.CourierUseCase
	tenantUseCase interfaces.TenantUseCase
	queue         otp_queue.OTPQueueRepository
}

// NewOTPDeliveryWorker creates a new worker instance
func NewOTPDeliveryWorker(
	curierUseCase interfaces.CourierUseCase,
	tenantUseCase interfaces.TenantUseCase,
	queue otp_queue.OTPQueueRepository,
) types.Worker {
	return &otpDeliveryWorker{
		curierUseCase: curierUseCase,
		tenantUseCase: tenantUseCase,
		queue:         queue,
	}
}

// Name returns the worker name
func (w *otpDeliveryWorker) Name() string {
	return "otp-delivery-worker"
}

// Start periodically checks OTP queue and delivers OTPs
func (w *otpDeliveryWorker) Start(ctx context.Context, interval time.Duration) {
	logger.GetLogger().Infof("[%s] started with interval %s", w.Name(), interval.String())
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processPendingOTPs(ctx)
			w.retryFailedOTPs(ctx)
		case <-ctx.Done():
			logger.GetLogger().Info("OTPDeliveryWorker stopped")
			return
		}
	}
}

// processPendingOTPs checks OTPs in queue and delivers them
func (w *otpDeliveryWorker) processPendingOTPs(ctx context.Context) {
	// Get all tenants to process
	tenants, err := w.tenantUseCase.GetAll(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get tenants: %v", err)
		return
	}
	tenantList := make([]string, len(tenants))
	for i, tenant := range tenants {
		tenantList[i] = tenant.Name
	}

	for _, tenant := range tenantList {
		receivers, err := w.queue.ListReceivers(ctx, tenant)
		if err != nil {
			logger.GetLogger().Errorf("Failed to list receivers for tenant %s: %v", tenant, err)
			continue
		}

		for _, receiver := range receivers {
			// TODO: could determine channel here or pass a default (e.g., SMS)
			err := w.curierUseCase.DeliverOTP(ctx, tenant, receiver, constants.ChannelSMS)
			if err != nil {
				logger.GetLogger().Warnf("Failed to deliver OTP to %s: %v", receiver, err)
			}
		}
	}
}

// retryFailedOTPs retries failed OTP deliveries
func (w *otpDeliveryWorker) retryFailedOTPs(ctx context.Context) {
	err := w.curierUseCase.RetryFailedOTPs(ctx, time.Now())
	if err != nil {
		logger.GetLogger().Errorf("Failed to retry OTPs: %v", err)
	}
}
