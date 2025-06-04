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

// // Save challenge session to cache for 5 minutes
// 	cacheKey := &cachingTypes.Keyer{Raw: sessionID}
// 	cacheValue := challengeSession{
// 		Type:  "email",
// 		Email: email,
// 		OTP:   otp,
// 	}
// 	if err := u.cacheRepo.SaveItem(cacheKey, cacheValue, 5*time.Minute); err != nil {
// 		return nil, &dto.ErrorDTOResponse{
// 			Status:  http.StatusInternalServerError,
// 			Code:    "MSG_CACHING_FAILED",
// 			Message: "Caching failed",
// 			Details: []interface{}{err.Error()},
// 		}
// 	}

func (r *challengeSessionRepository) SaveChallenge(_ context.Context, sessionID string, challenge *domain.ChallengeSession, ttl time.Duration) error {
	cacheKey := &cachingTypes.Keyer{Raw: sessionID}
	if err := r.cache.SaveItem(cacheKey, challenge, ttl); err != nil {
		return err
	}
	return nil
}

func (r *challengeSessionRepository) GetChallenge(_ context.Context, sessionID string) (*domain.ChallengeSession, error) {
	cacheKey := &cachingTypes.Keyer{Raw: sessionID}
	var challenge domain.ChallengeSession
	if err := r.cache.RetrieveItem(cacheKey, &challenge); err != nil {
		return nil, err
	}
	return &challenge, nil
}
