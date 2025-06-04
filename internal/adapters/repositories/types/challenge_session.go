package interfaces

import (
	"context"
	"time"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type ChallengeSessionRepository interface {
	SaveChallenge(ctx context.Context, sessionID string, challenge *domain.ChallengeSession, ttl time.Duration) error
	GetChallenge(ctx context.Context, sessionID string) (*domain.ChallengeSession, error)
}
