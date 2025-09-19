package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lifenetwork-ai/iam-service/constants"
	otp_queue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
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

		errs := deliverToReceivers(ctx, tenant, receivers, constants.OTPSendMaxReceiverConcurrency, w.curierUseCase.DeliverOTP)
		for _, e := range errs {
			logger.GetLogger().Warnf("Failed to deliver OTP to %s: %v", e.Receiver, e.Err)
		}
	}
}

// batchError represents an error occurred while sending OTP to a specific receiver.
type batchError struct {
	Receiver string
	Err      error
}

// deliverToReceivers sends OTP to multiple receivers concurrently with a bounded concurrency.
func deliverToReceivers(
	ctx context.Context,
	tenant string,
	receivers []string,
	maxConcurrency int,
	deliver func(context.Context, string, string) *errors.DomainError,
) []batchError {
	if len(receivers) == 0 {
		return nil
	}
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}
	if maxConcurrency > len(receivers) {
		maxConcurrency = len(receivers)
	}

	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []batchError

loop:
	for _, r := range receivers {
		receiver := r // capture per-iteration

		// Stop scheduling new tasks if context is canceled
		select {
		case <-ctx.Done():
			break loop
		case sem <- struct{}{}:
			// acquired slot
		}

		wg.Add(1)
		go func() {
			defer func() {
				<-sem
				wg.Done()
				if rec := recover(); rec != nil {
					// swallow panic per task; record as error
					mu.Lock()
					errs = append(errs, batchError{Receiver: receiver, Err: fmt.Errorf("panic: %v", rec)})
					mu.Unlock()
				}
			}()

			// If context already canceled, skip deliver
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := deliver(ctx, tenant, receiver); err != nil {
				mu.Lock()
				errs = append(errs, batchError{Receiver: receiver, Err: err})
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return errs
}
