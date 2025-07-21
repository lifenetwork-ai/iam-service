package ucases

import (
	"context"
	"time"

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
		defaultTTL: 5 * time.Minute,
	}
}

func (u *courierUseCase) ReceiveOTP(ctx context.Context, receiver, body string) *domainerrors.DomainError {
	return nil
}

func (u *courierUseCase) DeliverOTP(ctx context.Context, receiver, channel string) *domainerrors.DomainError {
	return nil
}

func (u *courierUseCase) RetryFailedOTPs(ctx context.Context, now time.Time) *domainerrors.DomainError {
	return nil
}
