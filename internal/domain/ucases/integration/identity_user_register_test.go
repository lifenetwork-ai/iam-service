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

func TestIntegration_Register_WithEmail_PersistsInIAMAndKratos(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-register-email")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Rate limiter expectations
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	email := "reg_email@test.com"
	reg, derr := ucase.Register(ctx, tenantID, "en", email, "")
	require.Nil(t, derr)
	require.NotNil(t, reg)

	ver, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)
	require.NotNil(t, ver)
	globalUserID := ver.User.GlobalUserID

	identities, err := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, globalUserID, tenantID.String())
	require.Nil(t, err)
	assertIAMHasIdentity(t, identities, constants.IdentifierEmail.String(), email)

	servc := deps.kratosService.(*kratos_service.FakeKratosService)
	ids, _ := servc.GetIdentities(ctx, tenantID)
	assertKratosHasIdentity(t, ids, constants.IdentifierEmail.String(), email, true)
}

func TestIntegration_Register_WithPhone_PersistsInIAMAndKratos(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-register-phone")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Rate limiter expectations
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	phone := "+84345678901"
	reg, derr := ucase.Register(ctx, tenantID, "en", "", phone)
	require.Nil(t, derr)
	require.NotNil(t, reg)

	ver, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)
	require.NotNil(t, ver)
	globalUserID := ver.User.GlobalUserID

	identities, err := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, globalUserID, tenantID.String())
	require.Nil(t, err)
	assertIAMHasIdentity(t, identities, constants.IdentifierPhone.String(), phone)

	servc := deps.kratosService.(*kratos_service.FakeKratosService)
	ids, _ := servc.GetIdentities(ctx, tenantID)
	assertKratosHasIdentity(t, ids, constants.IdentifierPhone.String(), phone, true)
}
