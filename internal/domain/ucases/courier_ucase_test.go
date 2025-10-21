package ucases

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsVietnamesePhone(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{
			name:     "E164 format Vietnamese number",
			phone:    "+84344381024",
			expected: true,
		},
		{
			name:     "Vietnamese number with spaces",
			phone:    " +84 344 381 024 ",
			expected: true,
		},
		{
			name:     "Non-Vietnamese E164 number (Thailand)",
			phone:    "+66812345678",
			expected: false,
		},
		{
			name:     "Non-Vietnamese E164 number (US)",
			phone:    "+14155552671",
			expected: false,
		},
		{
			name:     "Empty string",
			phone:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVietnamesePhone(tt.phone)
			assert.Equal(t, tt.expected, result, "isVietnamesePhone(%s) should be %v", tt.phone, tt.expected)
		})
	}
}
