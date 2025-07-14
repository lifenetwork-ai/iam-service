package constants

import "time"

// Rate limit within a window
const (
	MaxAttemptsPerWindow = 5
	RateLimitWindow      = 5 * time.Minute
)
