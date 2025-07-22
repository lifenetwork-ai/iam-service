package workers

import (
	"context"
	"time"

	otp_queue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	"github.com/lifenetwork-ai/iam-service/internal/workers/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type otpRetryWorker struct {
	curierUseCase interfaces.CourierUseCase
	queue         otp_queue.OTPQueueRepository
}

// NewOTPRetryWorker creates a new worker instance
func NewOTPRetryWorker(
	curierUseCase interfaces.CourierUseCase,
	queue otp_queue.OTPQueueRepository,
) types.Worker {
	return &otpRetryWorker{
		curierUseCase: curierUseCase,
		queue:         queue,
	}
}

// Name returns the worker name
func (w *otpRetryWorker) Name() string {
	return "otp-retry-worker"
}

// Start periodically retries failed OTP deliveries
func (w *otpRetryWorker) Start(ctx context.Context, interval time.Duration) {
	logger.GetLogger().Infof("[%s] started with interval %s", w.Name(), interval.String())
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.retryFailedOTPs(ctx)
		case <-ctx.Done():
			logger.GetLogger().Info("OTPRetryWorker stopped")
			return
		}
	}
}

// retryFailedOTPs retries failed OTP deliveries
func (w *otpRetryWorker) retryFailedOTPs(ctx context.Context) {
	err := w.curierUseCase.RetryFailedOTPs(ctx, time.Now())
	if err != nil {
		logger.GetLogger().Errorf("Failed to retry OTPs: %v", err)
	}
}
