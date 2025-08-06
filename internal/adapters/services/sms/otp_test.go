package sms

import (
	"testing"
	"time"

	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
)

func TestExtractOTPFromMessage(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		// Basic cases
		{"Your OTP is 123456", "123456"},
		{"Code: 9876", "9876"},
		{"otp 555666", "555666"},

		// Template format (with asterisks)
		{"*123456*", "123456"},

		// Edge cases
		{"No OTP here", ""},
		{"", ""},
		{"Phone: 1234567890", ""},

		// Template
		{common.GetOTPMessage("test", "645334", 10*time.Minute), "645334"},
		{common.GetOTPMessage("google", "123456", 1*time.Minute), "123456"},
	}

	for _, tt := range tests {
		result := common.ExtractOTPFromMessage(tt.message)
		if result != tt.expected {
			t.Errorf("Input: %q, got %q, want %q", tt.message, result, tt.expected)
		}
	}
}
