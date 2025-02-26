package services

import (
	"context"
)

type SMSService interface {
	Send(ctx context.Context, email, subject, body string) error
}

func NewSMSService() SMSService {
	return nil
}
