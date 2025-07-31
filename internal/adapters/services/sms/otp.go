package sms

import (
	"bytes"
	"fmt"
	"time"

	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

func GetOTPMessage(tenantName, otp string, ttl time.Duration) string {
	buf := bytes.Buffer{}
	err := OTPTemplate.Execute(&buf, map[string]string{
		"TenantName": tenantName,
		"OTP":        otp,
		"TTL":        fmt.Sprintf("%d", int(ttl.Minutes())),
	})
	if err != nil {
		logger.GetLogger().Errorf("Failed to execute OTP template: %v", err)
		return ""
	}
	return buf.String()
}
