package repositories

import (
	"context"
	"time"

	cachingTypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type challengeSessionRepository struct {
	cache cachingTypes.CacheRepository
}

func NewChallengeSessionRepository(cache cachingTypes.CacheRepository) interfaces.ChallengeSessionRepository {
	return &challengeSessionRepository{
		cache: cache,
	}
}

// SaveChallenge saves a challenge session in the cache with a specified TTL.
func (r *challengeSessionRepository) SaveChallenge(_ context.Context, sessionID string, challenge *domain.ChallengeSession, ttl time.Duration) error {
	cacheKey := &cachingTypes.Keyer{Raw: sessionID}
	return r.cache.SaveItem(cacheKey, challenge, ttl)
}

// GetChallenge retrieves a challenge session from the cache using the session ID.
// If the session does not exist, it returns an error.
func (r *challengeSessionRepository) GetChallenge(_ context.Context, sessionID string) (*domain.ChallengeSession, error) {
	cacheKey := &cachingTypes.Keyer{Raw: sessionID}
	var challenge *domain.ChallengeSession
	if err := r.cache.RetrieveItem(cacheKey, &challenge); err != nil {
		return nil, err
	}
	return challenge, nil
}
