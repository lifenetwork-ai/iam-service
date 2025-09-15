//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/lifenetwork-ai/iam-service/constants"
	kratos_service "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	utypes "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIntegration_ChangeIdentifier_TableDriven(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-change-table")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Set up rate limiter mocks for all operations
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	testCases := []struct {
		name         string
		initialType  string
		initialValue string
		newType      string
		newValue     string
	}{
		{"EmailToEmail", constants.IdentifierEmail.String(), "old1@test.com", constants.IdentifierEmail.String(), "new1@test.com"},
		{"PhoneToPhone", constants.IdentifierPhone.String(), "+84345381013", constants.IdentifierPhone.String(), "+84345381014"},
		{"PhoneToEmail", constants.IdentifierPhone.String(), "+84345381015", constants.IdentifierEmail.String(), "new2@test.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1) Register with initial identifier
			var reg *utypes.IdentityUserAuthResponse
			var derr *domainerrors.DomainError
			if tc.initialType == constants.IdentifierEmail.String() {
				reg, derr = ucase.Register(ctx, tenantID, "en", tc.initialValue, "")
			} else {
				reg, derr = ucase.Register(ctx, tenantID, "en", "", tc.initialValue)
			}
			require.Nil(t, derr)
			require.NotNil(t, reg)

			ver, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
			require.Nil(t, derr)
			require.NotNil(t, ver)
			globalUserID := ver.User.GlobalUserID

			// 2) If needed, add a second type (for email->email and phone->phone scenarios)
			if tc.name != "PhoneToEmail" {
				// Add the opposite type to ensure two identifiers exist
				if tc.initialType == constants.IdentifierEmail.String() {
					add, derr := ucase.AddNewIdentifier(ctx, tenantID, globalUserID, "+84999999999", constants.IdentifierPhone.String())
					require.Nil(t, derr)
					require.NotNil(t, add)
					_, derr = ucase.VerifyRegister(ctx, tenantID, add.FlowID, "000000")
					require.Nil(t, derr)
				} else {
					add, derr := ucase.AddNewIdentifier(ctx, tenantID, globalUserID, "add@test.com", constants.IdentifierEmail.String())
					require.Nil(t, derr)
					require.NotNil(t, add)
					_, derr = ucase.VerifyRegister(ctx, tenantID, add.FlowID, "000000")
					require.Nil(t, derr)
				}
			}

			// 3) Change identifier
			change, derr := ucase.ChangeIdentifier(ctx, globalUserID, tenantID, ver.User.ID, tc.newValue)
			require.Nil(t, derr)
			require.NotNil(t, change)
			_, derr = ucase.VerifyRegister(ctx, tenantID, change.FlowID, "000000")
			require.Nil(t, derr)

			// 4) Verify persistence in IAM
			identities, err := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, globalUserID, tenantID.String())
			require.Nil(t, err)
			if tc.name == "PhoneToEmail" {
				require.Equal(t, 1, len(identities))
				assertIAMHasIdentity(t, identities, constants.IdentifierEmail.String(), tc.newValue)
			} else {
				require.Equal(t, 2, len(identities))
				assertIAMHasIdentity(t, identities, tc.newType, tc.newValue)
			}

			// 5) Verify persistence in Kratos
			servc := deps.kratosService.(*kratos_service.FakeKratosService)
			ids, _ := servc.GetIdentities(ctx, tenantID)
			if tc.name == "PhoneToEmail" {
				assertKratosHasIdentity(t, ids, tc.newType, tc.newValue, true)
			} else {
				assertKratosHasIdentity(t, ids, tc.newType, tc.newValue, true)
			}
		})
	}
}
