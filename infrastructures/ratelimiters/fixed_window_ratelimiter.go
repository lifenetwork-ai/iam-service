package ratelimiters

import (
	"time"

	cachetypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	"github.com/lifenetwork-ai/iam-service/infrastructures/ratelimiters/types"
)

type fixedWindowRateLimiter struct {
	cacheRepo cachetypes.CacheRepository // dùng interface mới
}

func NewFixedWindowRateLimiter(cache cachetypes.CacheRepository) types.RateLimiter {
	return &fixedWindowRateLimiter{cacheRepo: cache}
}

func (r *fixedWindowRateLimiter) IsLimited(key string, limit int, window time.Duration) (bool, error) {
	var count int
	cacheKey := &cachetypes.Keyer{Raw: key}
	err := r.cacheRepo.RetrieveItem(cacheKey, &count)
	if err != nil {
		// not found or expired
		return false, nil
	}
	return count >= limit, nil
}

func (r *fixedWindowRateLimiter) RegisterAttempt(key string, window time.Duration) error {
	var count int
	cacheKey := &cachetypes.Keyer{Raw: key}
	err := r.cacheRepo.RetrieveItem(cacheKey, &count)
	if err != nil {
		// first time
		return r.cacheRepo.SaveItem(cacheKey, 1, window)
	}
	return r.cacheRepo.SaveItem(cacheKey, count+1, window)
}

func (r *fixedWindowRateLimiter) ResetAttempts(key string) error {
	cacheKey := &cachetypes.Keyer{Raw: key}
	return r.cacheRepo.RemoveItem(cacheKey)
}
