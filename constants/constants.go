package constants

import "time"

// Pagination
const (
	DEFAULT_PAGE_TEXT    = "page"
	DEFAULT_SIZE_TEXT    = "size"
	DEFAULT_PAGE         = "1"
	DEFAULT_PAGE_SIZE    = "10"
	DEFAULT_MIN_PAGESIZE = 5
	DEFAULT_MAX_PAGESIZE = 100
)

// Cache config
const (
	DefaultExpiration = 30 * time.Second
	CleanupInterval   = 1 * time.Minute
)

// Webhook constants
const WebhookTimeout = 5 * time.Second

// Concurrency
const MaxConcurrency = 10

// Order direction
type OrderDirection string

func (t OrderDirection) String() string {
	return string(t)
}

const (
	Asc  OrderDirection = "ASC"
	Desc OrderDirection = "DESC"
)

// DefaultChallengeDuration is the default duration for a challenge session
const DefaultChallengeDuration = 5 * time.Minute

// Challenge Types
const (
	ChallengeTypeLogin            = "login"
	ChallengeTypeRegister         = "register"
	ChallengeTypeChangeIdentifier = "change_identifier"
)

// HTTP Headers
const ContentTypeJson = "application/json"

// Courier constants
const (
	// Supported channels
	ChannelSMS      = "sms"
	ChannelWhatsApp = "whatsapp"
	ChannelZalo     = "zalo"

	// Tenant names
	TenantLifeAI   = "life_ai"
	TenantGenetica = "genetica"

	// Retry related constants
	MaxOTPRetryCount   = 3
	RetryDelayDuration = 30 * time.Second
	BaseRetryDuration  = 10 * time.Second // Base duration for exponential backoff

	// OTP worker interval
	OTPWorkerInterval = 10 * time.Second
)
