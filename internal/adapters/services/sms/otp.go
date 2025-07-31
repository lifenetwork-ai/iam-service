package sms

import (
	"bytes"
	"time"

	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

func GetOTPMessage(tenantName, otp string, ttl time.Duration) string {
	buf := bytes.Buffer{}
	err := OTPTemplate.Execute(&buf, map[string]any{
		"TenantName": tenantName,
		"OTP":        otp,
		"TTL":        int64(ttl.Minutes()),
	})
	if err != nil {
		logger.GetLogger().Errorf("Failed to execute OTP template: %v", err)
		return ""
	}
	return buf.String()
}
