package provider

import (
	"context"
	"time"
)

// SMSProvider defines the interface that all SMS providers must implement
type SMSProvider interface {
	SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error
	RefreshToken(ctx context.Context, refreshToken string) error
	GetChannelType() string
	HealthCheck(ctx context.Context) error
}
