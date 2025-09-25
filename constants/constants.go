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
const WebhookTimeout = 15 * time.Second

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
	ChallengeTypeAddIdentifier    = "add_identifier"
	ChallengeTypeVerifyIdentifier = "verify_identifier"
)

// HTTP Headers
const (
	HeaderContentTypeJson = "application/json" // Value for the header
	HeaderKeyContentType  = "Content-Type"     // Header name
)

// Courier constants
const (
	// Supported channels
	ChannelSMS      = "sms"
	ChannelWhatsApp = "whatsapp"
	ChannelZalo     = "zalo"

	// Retry related constants
	MaxOTPRetryCount   = 5
	RetryDelayDuration = 30 * time.Second
	BaseRetryDuration  = 3 * time.Second // Base duration for exponential backoff

	// OTP worker interval
	OTPDeliveryWorkerInterval = 1 * time.Second
	OTPRetryWorkerInterval    = 5 * time.Second

	OTPSendMaxReceiverConcurrency = 10

	// Zalo refresh token worker interval
	ZaloRefreshTokenWorkerInterval = 4 * time.Hour

	// Dev bypass OTP constants
	DevBypassOTPFetchTimeout = 10 * time.Second // how long we wait for the fresh OTP to appear in cache
	DevBypassOTPPollInterval = 200 * time.Millisecond
)

const (
	EnglishLanguage    = "en"
	VietnameseLanguage = "vi"
)

const (
	DefaultRegion = "VN"
)

const (
	DefaultSMSChannel = "webhook"
)
