package email_service

import (
	"context"
)

type EmailService interface {
	Send(ctx context.Context, email, subject, body string) error
}

func NewEmailService() EmailService {
	return nil
}
