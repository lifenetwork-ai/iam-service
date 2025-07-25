package repositories

import (
	"context"
	"time"

	"github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type challengeSessionRepository struct {
	cache types.CacheRepository
}

func NewChallengeSessionRepository(cache types.CacheRepository) domainrepo.ChallengeSessionRepository {
	return &challengeSessionRepository{
		cache: cache,
	}
}

// SaveChallenge saves a challenge session in the cache with a specified TTL.
func (r *challengeSessionRepository) SaveChallenge(_ context.Context, sessionID string, challenge *domain.ChallengeSession, ttl time.Duration) error {
	cacheKey := &types.Keyer{Raw: sessionID}
	return r.cache.SaveItem(cacheKey, challenge, ttl)
}

// GetChallenge retrieves a challenge session from the cache using the session ID.
// If the session does not exist, it returns an error.
func (r *challengeSessionRepository) GetChallenge(_ context.Context, sessionID string) (*domain.ChallengeSession, error) {
	cacheKey := &types.Keyer{Raw: sessionID}
	var challenge *domain.ChallengeSession
	if err := r.cache.RetrieveItem(cacheKey, &challenge); err != nil {
		return nil, err
	}
	return challenge, nil
}

// DeleteChallenge deletes a challenge session from the cache using the session ID.
func (r *challengeSessionRepository) DeleteChallenge(_ context.Context, sessionID string) error {
	cacheKey := &types.Keyer{Raw: sessionID}
	return r.cache.RemoveItem(cacheKey)
}
