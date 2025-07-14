package ratelimiters

import (
	"context"
	"time"

	cachetypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	"github.com/lifenetwork-ai/iam-service/infrastructures/ratelimiters/types"
)

type fixedWindowRateLimiter struct {
	cache cachetypes.CacheClient
}

func NewFixedWindowRateLimiter(cache cachetypes.CacheClient) types.RateLimiter {
	return &fixedWindowRateLimiter{cache: cache}
}

func (r *fixedWindowRateLimiter) IsLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	var count int
	err := r.cache.Get(ctx, key, &count)
	if err != nil {
		// Not found => no limit hit yet
		return false, nil
	}
	return count >= limit, nil
}

func (r *fixedWindowRateLimiter) RegisterAttempt(ctx context.Context, key string, window time.Duration) error {
	var count int
	err := r.cache.Get(ctx, key, &count)
	if err != nil {
		// First attempt → set to 1 with expiration
		return r.cache.Set(ctx, key, 1, window)
	}
	// Increment within window
	return r.cache.Set(ctx, key, count+1, window)
}

func (r *fixedWindowRateLimiter) ResetAttempts(ctx context.Context, key string) error {
	return r.cache.Del(ctx, key)
}
