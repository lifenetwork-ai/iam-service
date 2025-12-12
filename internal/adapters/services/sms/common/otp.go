package common

import (
	"bytes"
	"regexp"
	"time"

	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

func GetOTPMessage(tenantName, otp string, _ time.Duration) string {
	// Select client and brandname based on tenant
	var normalizedTenantName string
	if tenantName == constants.TenantLifeAI {
		normalizedTenantName = constants.TenantLifeAI
	} else {
		normalizedTenantName = constants.TenantGenetica
	}
	buf := bytes.Buffer{}
	err := SMSOTPTemplate.Execute(&buf, map[string]any{
		"TenantName": normalizedTenantName,
		"OTP":        otp,
	})
	if err != nil {
		logger.GetLogger().Errorf("Failed to execute OTP template: %v", err)
		return ""
	}
	return buf.String()
}

func ExtractOTPFromMessage(message string) string {
	// More robust regex to handle common OTP patterns
	patterns := []string{
		`\b(\d{6})\b`,         // 6 digits with word boundaries
		`code[:\s]*(\d{4,8})`, // "code: 123456" or "code 123456"
		`otp[:\s]*(\d{4,8})`,  // "otp: 123456"
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern) // case insensitive
		if matches := re.FindStringSubmatch(message); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}
