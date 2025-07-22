package workers

import (
	"context"
	"sync"
	"time"

	otp_queue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	"github.com/lifenetwork-ai/iam-service/internal/workers/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type otpRetryWorker struct {
	curierUseCase interfaces.CourierUseCase
	queue         otp_queue.OTPQueueRepository

	mu      sync.Mutex
	running bool
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
			go w.safeRetry(ctx)
		case <-ctx.Done():
			logger.GetLogger().Infof("[%s] stopped", w.Name())
			return
		}
	}
}

// safeRetry checks and prevents concurrent execution
func (w *otpRetryWorker) safeRetry(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		logger.GetLogger().Warnf("[%s] retry already in progress, skipping", w.Name())
		return
	}
	w.running = true
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		w.running = false
		w.mu.Unlock()
	}()

	logger.GetLogger().Infof("[%s] retry started", w.Name())
	err := w.curierUseCase.RetryFailedOTPs(ctx, time.Now())
	if err != nil {
		logger.GetLogger().Errorf("[%s] retry failed: %v", w.Name(), err)
	}
}
