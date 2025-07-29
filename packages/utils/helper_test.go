package utils

import (
	"testing"
	"time"

	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/stretchr/testify/require"
)

func TestComputeBackoffDuration(t *testing.T) {
	tests := []struct {
		retryCount int
		expected   time.Duration
	}{
		{retryCount: 1, expected: 20 * time.Second},                   // 2^1 * 10s
		{retryCount: 2, expected: 40 * time.Second},                   // 2^2 * 10s
		{retryCount: 3, expected: 80 * time.Second},                   // 2^3 * 10s
		{retryCount: 4, expected: 160 * time.Second},                  // 2^4 * 10s
		{retryCount: 5, expected: constants.DefaultChallengeDuration}, // 2^5 * 10s
	}

	for _, tt := range tests {
		t.Run("RetryCount_"+string(rune(tt.retryCount)), func(t *testing.T) {
			delay := ComputeBackoffDuration(tt.retryCount)
			require.Equal(t, tt.expected, delay, "wrong backoff for retry #%d", tt.retryCount)
			t.Logf("Retry #%d → delay = %v", tt.retryCount, delay)
		})
	}
}
