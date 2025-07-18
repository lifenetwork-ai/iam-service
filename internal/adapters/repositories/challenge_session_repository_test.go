package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/testutil"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/stretchr/testify/assert"
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
			Type:  "email",
			Email: "test@example.com",
			OTP:   "123456",
		}
		ttl := 5 * time.Minute

		// Test SaveChallenge
		err := repo.SaveChallenge(ctx, sessionID, challenge, ttl)
		require.NoError(t, err)

		// Verify item was saved in cache using CacheRepository interface
		var retrieved *domain.ChallengeSession
		err = mockCache.RetrieveItem(sessionKey(sessionID), &retrieved)
		require.NoError(t, err)
		assert.Equal(t, challenge.Type, retrieved.Type)
		assert.Equal(t, challenge.Email, retrieved.Email)
		assert.Equal(t, challenge.OTP, retrieved.OTP)

		// Test GetChallenge
		retrievedChallenge, err := repo.GetChallenge(ctx, sessionID)
		require.NoError(t, err)
		assert.NotNil(t, retrievedChallenge)
		assert.Equal(t, challenge.Type, retrievedChallenge.Type)
		assert.Equal(t, challenge.Email, retrievedChallenge.Email)
		assert.Equal(t, challenge.OTP, retrievedChallenge.OTP)
	})

	t.Run("GetChallenge - Not Found", func(t *testing.T) {
		// Test getting non-existent challenge
		retrieved, err := repo.GetChallenge(ctx, "non-existent-session")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("SaveChallenge - Phone Challenge", func(t *testing.T) {
		// Test data
		sessionID := "test-session-456"
		challenge := &domain.ChallengeSession{
			Type:  constants.IdentifierPhone.String(),
			Phone: "+1234567890",
			OTP:   "654321",
		}
		ttl := 5 * time.Minute

		// Test SaveChallenge
		err := repo.SaveChallenge(ctx, sessionID, challenge, ttl)
		require.NoError(t, err)

		// Verify item was saved in cache using CacheRepository interface
		var retrieved *domain.ChallengeSession
		err = mockCache.RetrieveItem(sessionKey(sessionID), &retrieved)
		require.NoError(t, err)
		assert.Equal(t, challenge.Type, retrieved.Type)
		assert.Equal(t, challenge.Phone, retrieved.Phone)
		assert.Equal(t, challenge.OTP, retrieved.OTP)

		// Test GetChallenge
		retrievedChallenge, err := repo.GetChallenge(ctx, sessionID)
		require.NoError(t, err)
		assert.NotNil(t, retrievedChallenge)
		assert.Equal(t, challenge.Type, retrievedChallenge.Type)
		assert.Equal(t, challenge.Phone, retrievedChallenge.Phone)
		assert.Equal(t, challenge.OTP, retrievedChallenge.OTP)
	})

	t.Run("RemoveChallenge", func(t *testing.T) {
		// Test removing challenges
		err := mockCache.RemoveItem(sessionKey("test-session-123"))
		require.NoError(t, err)

		err = mockCache.RemoveItem(sessionKey("test-session-456"))
		require.NoError(t, err)

		// Verify items are gone
		retrieved, err := repo.GetChallenge(ctx, "test-session-123")
		assert.Error(t, err)
		assert.Nil(t, retrieved)

		retrieved, err = repo.GetChallenge(ctx, "test-session-456")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})
}
