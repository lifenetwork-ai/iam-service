package interfaces

import (
	"context"
	"time"

	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

type CourierUseCase interface {
	ReceiveOTP(ctx context.Context, receiver, body string) *domainerrors.DomainError
	GetAvailableChannels(ctx context.Context, tenantName, receiver string) []string
	DeliverOTP(ctx context.Context, tenantName, receiver string) *domainerrors.DomainError
	RetryFailedOTPs(ctx context.Context, now time.Time) (int, *domainerrors.DomainError)
	ChooseChannel(ctx context.Context, tenantName, receiver, channel string) *domainerrors.DomainError
	GetChannel(ctx context.Context, tenantName, receiver string) (types.ChooseChannelResponse, *domainerrors.DomainError)
}
