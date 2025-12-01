//go:build integration

package integration

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	kratos_service "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	adaptersrepo "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	utypes "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	mock_services "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/services"
	"github.com/stretchr/testify/require"
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

// These tests lock in the desired DeleteIdentifier semantics using a GoMock KratosService:
// - Kratos delete must succeed before IAM deletion proceeds
// - If Kratos delete fails, IAM must remain unchanged

func TestIntegration_DeleteIdentifier_KratosSuccess_DeletesIAM(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange Kratos mock
	kratosMock := mock_services.NewMockKratosService(ctrl)

	// Harness with injectable Kratos
	ucase, _, deps, db, tenantID, container := startPostgresAndBuildUCaseWithKratos(t, ctx, ctrl, "tenant-del-success", kratosMock)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Seed a user with two identifiers (email and phone)
	gu := &domain.GlobalUser{}
	require.NoError(t, adaptersrepo.NewGlobalUserRepository(db).Create(db, gu))

	email := "delsuccess@test.com"
	phone := "+84123456789"
	emailKratosID := uuid.NewString()
	phoneKratosID := uuid.NewString()

	_, err := deps.userIdentityRepo.InsertOnceByKratosUserAndType(ctx, db, tenantID.String(), emailKratosID, gu.ID, constants.IdentifierEmail.String(), email)
	require.NoError(t, err)
	_, err = deps.userIdentityRepo.InsertOnceByKratosUserAndType(ctx, db, tenantID.String(), phoneKratosID, gu.ID, constants.IdentifierPhone.String(), phone)
	require.NoError(t, err)

	// Expect Kratos delete to succeed for the phone identity
	kratosMock.EXPECT().DeleteIdentifierAdmin(gomock.Any(), tenantID, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ uuid.UUID, id uuid.UUID) error {
			require.Equal(t, uuid.MustParse(phoneKratosID), id)
			return nil
		},
	)

	// Act
	derr := ucase.DeleteIdentifier(ctx, gu.ID, tenantID, phoneKratosID, constants.IdentifierPhone.String())

	// Assert
	require.Nil(t, derr)
	identities, qerr := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, gu.ID, tenantID.String())
	require.Nil(t, qerr)
	// Phone should be removed; email should remain
	var hasPhone, hasEmail bool
	for _, id := range identities {
		if id.Type == constants.IdentifierPhone.String() {
			hasPhone = true
		}
		if id.Type == constants.IdentifierEmail.String() {
			hasEmail = true
		}
	}
	require.False(t, hasPhone)
	require.True(t, hasEmail)
}

func TestIntegration_DeleteIdentifier_KratosFailure_AbortsIAMDeletion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kratosMock := mock_services.NewMockKratosService(ctrl)
	ucase, _, deps, db, tenantID, container := startPostgresAndBuildUCaseWithKratos(t, ctx, ctrl, "tenant-del-fail", kratosMock)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Seed a user with two identifiers
	gu := &domain.GlobalUser{}
	require.NoError(t, adaptersrepo.NewGlobalUserRepository(db).Create(db, gu))

	email := "delfail@test.com"
	phone := "+84987654321"
	emailKratosID := uuid.NewString()
	phoneKratosID := uuid.NewString()

	_, err := deps.userIdentityRepo.InsertOnceByKratosUserAndType(ctx, db, tenantID.String(), emailKratosID, gu.ID, constants.IdentifierEmail.String(), email)
	require.NoError(t, err)
	_, err = deps.userIdentityRepo.InsertOnceByKratosUserAndType(ctx, db, tenantID.String(), phoneKratosID, gu.ID, constants.IdentifierPhone.String(), phone)
	require.NoError(t, err)

	// Expect Kratos delete to fail
	kratosMock.EXPECT().DeleteIdentifierAdmin(gomock.Any(), tenantID, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ uuid.UUID, id uuid.UUID) error {
			require.Equal(t, uuid.MustParse(phoneKratosID), id)
			return assertErr("kratos delete failed")
		},
	)

	// Act
	derr := ucase.DeleteIdentifier(ctx, gu.ID, tenantID, phoneKratosID, constants.IdentifierPhone.String())

	// Assert desired semantics (TDD): error returned and IAM unchanged
	require.NotNil(t, derr)
	identities, qerr := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, gu.ID, tenantID.String())
	require.Nil(t, qerr)
	var hasPhone, hasEmail bool
	for _, id := range identities {
		if id.Type == constants.IdentifierPhone.String() {
			hasPhone = true
		}
		if id.Type == constants.IdentifierEmail.String() {
			hasEmail = true
		}
	}
	require.True(t, hasPhone)
	require.True(t, hasEmail)
}

func TestIntegration_DeleteIdentifier_KratosNetworkError_AbortsIAMDeletion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kratosMock := mock_services.NewMockKratosService(ctrl)
	ucase, _, deps, db, tenantID, container := startPostgresAndBuildUCaseWithKratos(t, ctx, ctrl, "tenant-del-neterr", kratosMock)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	gu := &domain.GlobalUser{}
	require.NoError(t, adaptersrepo.NewGlobalUserRepository(db).Create(db, gu))

	email := "delnet@test.com"
	phone := "+84888888888"
	emailKratosID := uuid.NewString()
	phoneKratosID := uuid.NewString()

	_, err := deps.userIdentityRepo.InsertOnceByKratosUserAndType(ctx, db, tenantID.String(), emailKratosID, gu.ID, constants.IdentifierEmail.String(), email)
	require.NoError(t, err)
	_, err = deps.userIdentityRepo.InsertOnceByKratosUserAndType(ctx, db, tenantID.String(), phoneKratosID, gu.ID, constants.IdentifierPhone.String(), phone)
	require.NoError(t, err)

	kratosMock.EXPECT().DeleteIdentifierAdmin(gomock.Any(), tenantID, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ uuid.UUID, id uuid.UUID) error {
			require.Equal(t, uuid.MustParse(phoneKratosID), id)
			return assertErr("network error")
		},
	)

	derr := ucase.DeleteIdentifier(ctx, gu.ID, tenantID, phoneKratosID, constants.IdentifierPhone.String())

	require.NotNil(t, derr)
	identities, qerr := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, gu.ID, tenantID.String())
	require.Nil(t, qerr)
	var hasPhone, hasEmail bool
	for _, id := range identities {
		if id.Type == constants.IdentifierPhone.String() {
			hasPhone = true
		}
		if id.Type == constants.IdentifierEmail.String() {
			hasEmail = true
		}
	}
	require.True(t, hasPhone)
	require.True(t, hasEmail)
}

// assertErr is a tiny helper that returns an error for mock expectations
func assertErr(msg string) error { return &mockErr{msg: msg} }

type mockErr struct{ msg string }

func (e *mockErr) Error() string { return e.msg }
