//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/lifenetwork-ai/iam-service/constants"
	kratos_service "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Covers: Login (password flow)
func TestIntegration_Login_WithPassword(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-login")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	email := "loginpw@test.com"

	// Register with email and verify (to create identity in Kratos)
	reg, derr := ucase.Register(ctx, tenantID, "en", email, "")
	require.Nil(t, derr)
	_, derr = ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	// Perform password login (FakeKratos accepts any password)
	auth, derr := ucase.Login(ctx, tenantID, email, "some-password")
	require.Nil(t, derr)
	require.NotNil(t, auth)
	require.True(t, auth.Active)
	require.Equal(t, email, auth.User.Email)
}

// Covers: DeleteIdentifier prevents deleting the only identifier
func TestIntegration_DeleteIdentifier_OnlyIdentifier_Fails(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-delete-only")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	email := "only@test.com"
	reg, derr := ucase.Register(ctx, tenantID, "en", email, "")
	require.Nil(t, derr)
	ver, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	// Get kratos id for the only identifier (email)
	ids, err := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, ver.User.GlobalUserID, tenantID.String())
	require.Nil(t, err)
	require.Len(t, ids, 1)

	derr = ucase.DeleteIdentifier(ctx, ver.User.GlobalUserID, tenantID, ids[0].KratosUserID, constants.IdentifierEmail.String())
	require.NotNil(t, derr)
	require.Equal(t, "MSG_CANNOT_DELETE_ONLY_IDENTIFIER", derr.Code)
}

// Covers: ChallengeVerification non-existent identifier returns not found
func TestIntegration_ChallengeVerification_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, _, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-chall-notfound")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	chall, derr := ucase.ChallengeVerification(ctx, tenantID, "ghost@test.com")
	require.NotNil(t, derr)
	require.Nil(t, chall)
	require.Equal(t, "MSG_IDENTITY_NOT_FOUND", derr.Code)
}

// Covers: VerifyLogin error branch (GetLoginFlow failure)
func TestIntegration_VerifyLogin_GetFlowFails(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-verifylogin-fail")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	fake := deps.kratosService.(*kratos_service.FakeKratosService)
	fake.SetFaults(kratos_service.Faults{NetworkError: true})

	auth, derr := ucase.VerifyLogin(ctx, tenantID, "any-flow", "000000")
	require.NotNil(t, derr)
	require.Nil(t, auth)
	require.Equal(t, "MSG_GET_FLOW_FAILED", derr.Code)
}

// Covers: UpdateLang error path when Kratos update fails for a peer identity
func TestIntegration_UpdateLang_FailsOnKratosUpdate(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-updatelang-fail")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Seed user with phone
	phone := "+84320010003"
	reg, derr := ucase.Register(ctx, tenantID, "en", "", phone)
	require.Nil(t, derr)
	ver, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	// Force Kratos update failure
	fake := deps.kratosService.(*kratos_service.FakeKratosService)
	fake.SetFaults(kratos_service.Faults{FailUpdate: true})

	// Attempt UpdateLang should fail (use supported lang to reach Kratos update error branch)
	derr = ucase.UpdateLang(ctx, tenantID, ver.User.ID, "vi")
	require.NotNil(t, derr)
	require.Equal(t, "MSG_UPDATE_LANG_FAILED", derr.Code)
}
