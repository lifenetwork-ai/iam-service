package workers

import (
	"context"
	"sync"
	"time"

	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms"
	"github.com/lifenetwork-ai/iam-service/internal/workers/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type zaloRefreshTokenWorker struct {
	smsService *sms.SMSService

	mu      sync.Mutex
	running bool
}

// NewZaloRefreshTokenWorker creates a new worker instance
func NewZaloRefreshTokenWorker(
	smsService *sms.SMSService,
) types.Worker {
	return &zaloRefreshTokenWorker{
		smsService: smsService,
	}
}

// Name returns the worker name
func (w *zaloRefreshTokenWorker) Name() string {
	return "zalo-refresh-token-worker"
}

// Start periodically retries failed OTP deliveries
func (w *zaloRefreshTokenWorker) Start(ctx context.Context, interval time.Duration) {
	logger.GetLogger().Infof("[%s] started with interval %s", w.Name(), interval.String())
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go w.safeProcess(ctx)
		case <-ctx.Done():
			logger.GetLogger().Infof("[%s] stopped", w.Name())
			return
		}
	}
}

// safeProcess checks and prevents concurrent execution
func (w *zaloRefreshTokenWorker) safeProcess(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		logger.GetLogger().Warnf("[%s] still processing, skipping this tick", w.Name())
		return
	}
	w.running = true
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		w.running = false
		w.mu.Unlock()
	}()

	w.processZaloToken(ctx)
}

func (w *zaloRefreshTokenWorker) processZaloToken(ctx context.Context) {
	provider, err := w.smsService.GetProvider(constants.ChannelZalo)
	if err != nil {
		logger.GetLogger().Errorf("[%s] failed to get zalo provider: %v", w.Name(), err)
		return
	}

	err = provider.RefreshToken(ctx, "")
	if err != nil {
		logger.GetLogger().Errorf("[%s] failed to refresh zalo token: %v", w.Name(), err)
		return
	}

	logger.GetLogger().Infof("[%s] refreshed zalo token", w.Name())
}
