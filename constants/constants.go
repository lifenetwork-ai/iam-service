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
const (
	MaxWebhookWorkers = 10
	WebhookTimeout    = 5 * time.Second
)

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
