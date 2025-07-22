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
		// Use in-memory cache client (you can switch to Redis later)
		cacheRepo := CacheRepositoryInstance(context.Background())

		// Use fixed window limiter
		rateLimiterInstance = ratelimiters.NewFixedWindowRateLimiter(cacheRepo)
	})

	return rateLimiterInstance
}
