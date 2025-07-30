package sms

import (
	"bytes"

	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

func GetOTPMessage(tenantName, otp string) string {
	buf := bytes.Buffer{}
	err := OTPTemplate.Execute(&buf, map[string]string{
		"TenantName": tenantName,
		"OTP":        otp,
	})
	if err != nil {
		logger.GetLogger().Errorf("Failed to execute OTP template: %v", err)
		return ""
	}
	return buf.String()
}
