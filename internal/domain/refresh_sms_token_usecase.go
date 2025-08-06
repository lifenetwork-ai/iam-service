package domain

import (
	"context"

	services "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/services"
)

type RefreshSmsTokenUsecase struct {
	smsService services.SmsTokenRefresher
}

func NewRefreshSmsTokenUsecase(smsService services.SmsTokenRefresher) *RefreshSmsTokenUsecase {
	return &RefreshSmsTokenUsecase{
		smsService: smsService,
	}
}

func (u *RefreshSmsTokenUsecase) RefreshZaloToken(ctx context.Context) error {
	return u.smsService.RefreshZaloToken(ctx)
}
