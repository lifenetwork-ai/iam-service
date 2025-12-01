package instances

import (
	"context"
	"sync"

	ratelimiters "github.com/lifenetwork-ai/iam-service/infrastructures/rate_limiter"
	"github.com/lifenetwork-ai/iam-service/infrastructures/rate_limiter/types"
)

var (
	rateLimiterOnce     sync.Once
	rateLimiterInstance types.RateLimiter
)

// RateLimiterInstance returns a singleton instance of the rate limiter.
func RateLimiterInstance() types.RateLimiter {
	rateLimiterOnce.Do(func() {
		// Initialize cache repository
		cacheRepo := CacheRepositoryInstance(context.Background())

		// Use fixed window limiter
		rateLimiterInstance = ratelimiters.NewFixedWindowRateLimiter(cacheRepo)
	})

	return rateLimiterInstance
}
