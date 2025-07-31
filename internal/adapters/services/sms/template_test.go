package sms

import (
	"bytes"
	"strings"
	"testing"
)

type OTPData struct {
	TenantName string
	OTP        string
	TTL        int
}

func TestOTPTemplate(t *testing.T) {
	tests := []struct {
		name     string
		data     OTPData
		expected []string // strings that should be present in output
	}{
		{
			name: "TTL equals 1 - should use singular 'minute'",
			data: OTPData{
				TenantName: "TestApp",
				OTP:        "123456",
				TTL:        1,
			},
			expected: []string{
				"Dear Valued Customer",
				"TestApp",
				"*123456*",
				"*1* minute.",
				"Do not share this OTP",
				"TestApp takes your account security",
			},
		},
		{
			name: "TTL equals 5 - should use plural 'minutes'",
			data: OTPData{
				TenantName: "MyCompany",
				OTP:        "789012",
				TTL:        5,
			},
			expected: []string{
				"Dear Valued Customer",
				"MyCompany",
				"*789012*",
				"*5* minutes.",
				"Do not share this OTP",
				"MyCompany takes your account security",
			},
		},
		{
			name: "TTL equals 0 - should use plural 'minutes'",
			data: OTPData{
				TenantName: "ZeroApp",
				OTP:        "000000",
				TTL:        0,
			},
			expected: []string{
				"ZeroApp",
				"*000000*",
				"*0* minutes.",
			},
		},
		{
			name: "TTL equals 15 - should use plural 'minutes'",
			data: OTPData{
				TenantName: "LongApp",
				OTP:        "999888",
				TTL:        15,
			},
			expected: []string{
				"LongApp",
				"*999888*",
				"*15* minutes.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := OTPTemplate.Execute(&buf, tt.data)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			output := buf.String()

			// Check that all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't.\nFull output:\n%s", expected, output)
				}
			}
		})
	}
}
