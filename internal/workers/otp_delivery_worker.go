package workers

import (
	"context"
	"sync"
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

	mu      sync.Mutex
	running bool
}

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

func (w *otpDeliveryWorker) Name() string {
	return "otp-delivery-worker"
}

func (w *otpDeliveryWorker) Start(ctx context.Context, interval time.Duration) {
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

func (w *otpDeliveryWorker) safeProcess(ctx context.Context) {
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

	w.processPendingOTPs(ctx)
}

func (w *otpDeliveryWorker) processPendingOTPs(ctx context.Context) {
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
			err := w.curierUseCase.DeliverOTP(ctx, tenant, receiver, constants.ChannelSMS)
			if err != nil {
				logger.GetLogger().Warnf("Failed to deliver OTP to %s: %v", receiver, err)
			}
		}
	}
}
