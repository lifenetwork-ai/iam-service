package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/testutil"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/stretchr/testify/require"
)

// sessionKey implements fmt.Stringer for the cache key
type sessionKey string

func (s sessionKey) String() string {
	return string(s)
}

func TestChallengeSessionRepository(t *testing.T) {
	mockCache := testutil.NewMockCache()
	repo := NewChallengeSessionRepository(mockCache)
	ctx := context.Background()

	t.Run("SaveChallenge and GetChallenge", func(t *testing.T) {
		// Test data
		sessionID := "test-session-123"
		challenge := &domain.ChallengeSession{
			IdentifierType: "email",
			Identifier:     "test@example.com",
			OTP:            "123456",
		}
		ttl := 5 * time.Minute

		// Test SaveChallenge
		err := repo.SaveChallenge(ctx, sessionID, challenge, ttl)
		require.NoError(t, err)

		// Verify item was saved in cache using CacheRepository interface
		var retrieved *domain.ChallengeSession
		err = mockCache.RetrieveItem(sessionKey(sessionID), &retrieved)
		require.NoError(t, err)
		require.Equal(t, challenge.IdentifierType, retrieved.IdentifierType)
		require.Equal(t, challenge.Identifier, retrieved.Identifier)
		require.Equal(t, challenge.OTP, retrieved.OTP)

		// Test GetChallenge
		retrievedChallenge, err := repo.GetChallenge(ctx, sessionID)
		require.NoError(t, err)
		require.NotNil(t, retrievedChallenge)
		require.Equal(t, challenge.IdentifierType, retrievedChallenge.IdentifierType)
		require.Equal(t, challenge.Identifier, retrievedChallenge.Identifier)
		require.Equal(t, challenge.OTP, retrievedChallenge.OTP)
	})

	t.Run("GetChallenge - Not Found", func(t *testing.T) {
		// Test getting non-existent challenge
		retrieved, err := repo.GetChallenge(ctx, "non-existent-session")
		require.Error(t, err)
		require.Nil(t, retrieved)
	})

	t.Run("SaveChallenge - Phone Challenge", func(t *testing.T) {
		// Test data
		sessionID := "test-session-456"
		challenge := &domain.ChallengeSession{
			IdentifierType: constants.IdentifierPhone.String(),
			Identifier:     "+1234567890",
			OTP:            "654321",
		}
		ttl := 5 * time.Minute

		// Test SaveChallenge
		err := repo.SaveChallenge(ctx, sessionID, challenge, ttl)
		require.NoError(t, err)

		// Verify item was saved in cache using CacheRepository interface
		var retrieved *domain.ChallengeSession
		err = mockCache.RetrieveItem(sessionKey(sessionID), &retrieved)
		require.NoError(t, err)
		require.Equal(t, challenge.IdentifierType, retrieved.IdentifierType)
		require.Equal(t, challenge.Identifier, retrieved.Identifier)
		require.Equal(t, challenge.OTP, retrieved.OTP)

		// Test GetChallenge
		retrievedChallenge, err := repo.GetChallenge(ctx, sessionID)
		require.NoError(t, err)
		require.NotNil(t, retrievedChallenge)
		require.Equal(t, challenge.IdentifierType, retrievedChallenge.IdentifierType)
		require.Equal(t, challenge.Identifier, retrievedChallenge.Identifier)
		require.Equal(t, challenge.OTP, retrievedChallenge.OTP)
	})

	t.Run("RemoveChallenge", func(t *testing.T) {
		// Test removing challenges
		err := mockCache.RemoveItem(sessionKey("test-session-123"))
		require.NoError(t, err)

		err = mockCache.RemoveItem(sessionKey("test-session-456"))
		require.NoError(t, err)

		// Verify items are gone
		retrieved, err := repo.GetChallenge(ctx, "test-session-123")
		require.Error(t, err)
		require.Nil(t, retrieved)

		retrieved, err = repo.GetChallenge(ctx, "test-session-456")
		require.Error(t, err)
		require.Nil(t, retrieved)
	})
}
