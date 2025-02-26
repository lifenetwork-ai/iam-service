package services

import (
	"context"
	"fmt"
)

type EmailService interface {
	SendOTP(
		ctx context.Context,
		destination string,
		otp string,
	) error
}

type EmailServiceSetting struct{}

func NewEmailService() EmailService {
	return &EmailServiceSetting{}
}

func (c *EmailServiceSetting) SendOTP(
	ctx context.Context,
	destination string,
	otp string,
) error {
	organizationId := ctx.Value("organizationId")
	if organizationId == nil {
		return fmt.Errorf("organization ID is required")
	}

	return nil
}
