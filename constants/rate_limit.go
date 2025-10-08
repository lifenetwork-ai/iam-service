package constants

import "time"

// Rate limit within a window
const (
	MaxAttemptsPerWindow = 5
	RateLimitWindow      = 5 * time.Minute
)

// Actions for rate limiting
const (
	LoginWithPhoneAction = "login_phone"
	LoginWithEmailAction = "login_email"
)
