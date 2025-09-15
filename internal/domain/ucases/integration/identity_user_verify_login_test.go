//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Covers: ChallengeWithEmail -> VerifyLogin
func TestIntegration_VerifyLogin_WithEmail(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-verify-login")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Allow flows without rate limits
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	email := "verifylogin@test.com"

	// Register with email
	reg, derr := ucase.Register(ctx, tenantID, "en", email, "")
	require.Nil(t, derr)
	require.NotNil(t, reg)
	// Verify registration to create identity and session
	ver, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)
	require.NotNil(t, ver)

	// Challenge login with email
	chall, derr := ucase.ChallengeWithEmail(ctx, tenantID, email)
	require.Nil(t, derr)
	require.NotNil(t, chall)
	require.NotEmpty(t, chall.FlowID)

	// Verify login with code
	auth, derr := ucase.VerifyLogin(ctx, tenantID, chall.FlowID, "000000")
	require.Nil(t, derr)
	require.NotNil(t, auth)
	require.True(t, auth.Active)
	require.Equal(t, email, auth.User.Email)
	require.NotEmpty(t, auth.SessionToken)

	// Quick smoke: Profile with the new token works
	ctxWithToken := context.WithValue(ctx, constants.SessionTokenKey, auth.SessionToken)
	prof, derr := ucase.Profile(ctxWithToken, tenantID)
	require.Nil(t, derr)
	require.NotNil(t, prof)
	require.Equal(t, email, prof.Email)
}

// Covers: Logout success and Profile invalidation
func TestIntegration_Logout_Succeeds_And_Profile_Invalid(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-logout-success")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	email := "logoutsucc@test.com"
	reg, derr := ucase.Register(ctx, tenantID, "en", email, "")
	require.Nil(t, derr)
	ver, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	// Put session token in context and logout
	ctxWithToken := context.WithValue(ctx, constants.SessionTokenKey, ver.SessionToken)
	derr = ucase.Logout(ctxWithToken, tenantID)
	require.Nil(t, derr)

	// Profile should now be invalid
	prof, derr := ucase.Profile(ctxWithToken, tenantID)
	require.Nil(t, prof)
	require.NotNil(t, derr)
}
