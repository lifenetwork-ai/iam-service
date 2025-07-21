package workers

import (
	"context"
	"time"

	otp_queue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	"github.com/lifenetwork-ai/iam-service/internal/workers/types"
)

type otpDeliveryWorker struct {
	uc    interfaces.CourierUseCase
	queue otp_queue.OTPQueueRepository
}

// NewOTPDeliveryWorker creates a new worker instance
func NewOTPDeliveryWorker(uc interfaces.CourierUseCase, queue otp_queue.OTPQueueRepository) types.Worker {
	return &otpDeliveryWorker{
		uc:    uc,
		queue: queue,
	}
}

// Start periodically checks OTP queue and delivers OTPs
func (w *otpDeliveryWorker) Start(ctx context.Context, interval time.Duration) {
}
