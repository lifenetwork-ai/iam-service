//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Covers rate limit error paths by mocking limiter to return limited
func TestIntegration_RateLimit_Paths(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-rate-limit")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Force limiter to return limited on any key
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// ChallengeWithEmail should fail with rate-limit
	chall, derr := ucase.ChallengeWithEmail(ctx, tenantID, "rl@test.com")
	require.NotNil(t, derr)
	require.Nil(t, chall)
	require.Equal(t, "MSG_RATE_LIMIT_EXCEEDED", derr.Code)

	// ChallengeWithPhone should fail with rate-limit
	chall2, derr := ucase.ChallengeWithPhone(ctx, tenantID, "+84300000000")
	require.NotNil(t, derr)
	require.Nil(t, chall2)
	require.Equal(t, "MSG_RATE_LIMIT_EXCEEDED", derr.Code)

	// Register should fail with rate-limit (email)
	reg, derr := ucase.Register(ctx, tenantID, "en", "rl2@test.com", "")
	require.NotNil(t, derr)
	require.Nil(t, reg)
	require.Equal(t, "MSG_RATE_LIMIT_EXCEEDED", derr.Code)

	// VerifyRegister should fail with rate-limit
	auth, derr := ucase.VerifyRegister(ctx, tenantID, "flow", "000000")
	require.NotNil(t, derr)
	require.Nil(t, auth)
	require.Equal(t, "MSG_RATE_LIMIT_EXCEEDED", derr.Code)

	// VerifyLogin should fail with rate-limit
	auth2, derr := ucase.VerifyLogin(ctx, tenantID, "flow2", "000000")
	require.NotNil(t, derr)
	require.Nil(t, auth2)
	require.Equal(t, "MSG_RATE_LIMIT_EXCEEDED", derr.Code)

	// Note: AddNewIdentifier, ChangeIdentifier, ChallengeVerification perform other validations
	// before rate-limit checks; they are intentionally excluded here.

	// VerifyIdentifier should fail with rate-limit
	verRes, derr := ucase.VerifyIdentifier(ctx, tenantID, "flow3", "000000")
	require.NotNil(t, derr)
	require.Nil(t, verRes)
	require.Equal(t, "MSG_RATE_LIMIT_EXCEEDED", derr.Code)
}
