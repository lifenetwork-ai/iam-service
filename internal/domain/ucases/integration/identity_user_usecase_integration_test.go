//go:build integration

package integration

import (
	"context"
	"testing"

	kratos "github.com/ory/kratos-client-go"
	"go.uber.org/mock/gomock"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	adaptersrepo "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	kratos_service "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ChangeIdentifier_EmailToEmail_LoginChecks(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-change-identifier-email-to-email")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	var globalUserID string
	oldEmail := "oldemail@test.com"
	newEmail := "newemail@test.com"
	phone := "+84345381013"

	// Tenant is seeded by startPostgresAndBuildUCase
	// Register a phone identity
	registerResp, registerErr := ucase.Register(ctx, tenantID, "en", "", phone)
	require.Nil(t, registerErr)
	require.NotNil(t, registerResp)

	verifyRegisterResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, registerResp.VerificationFlow.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyRegisterResp)
	globalUserID = verifyRegisterResp.User.GlobalUserID

	// Act: add an email identity
	addEmailResp, addEmailErr := ucase.AddNewIdentifier(ctx, tenantID, globalUserID, oldEmail, constants.IdentifierEmail.String())
	require.Nil(t, addEmailErr)
	require.NotNil(t, addEmailResp)

	verifyAddEmailResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, addEmailResp.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyAddEmailResp)

	// Act: change identifier phone→email
	changeResp, changeErr := ucase.ChangeIdentifier(ctx, globalUserID, tenantID, verifyRegisterResp.User.ID, newEmail)
	require.Nil(t, changeErr)
	require.NotNil(t, changeResp)

	verifyChangeResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, changeResp.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyChangeResp)

	// Query identities in IAM, ensure the old email identity is deleted on IAM side
	identities, ucaseErr := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, globalUserID, tenantID.String())
	require.Nil(t, ucaseErr)
	require.NotNil(t, identities)
	require.Equal(t, 2, len(identities))
	assertIAMHasIdentity(t, identities, constants.IdentifierPhone.String(), phone)
	assertIAMHasIdentity(t, identities, constants.IdentifierEmail.String(), newEmail)

	//  query kratos, ensure the old email identity is deleted on Kratos side
	servc := deps.kratosService.(*kratos_service.FakeKratosService)
	ids, _ := servc.GetIdentities(ctx, tenantID)
	new, ok := ids[newEmail]
	require.True(t, ok)
	require.Equal(t, newEmail, new.Traits.(map[string]interface{})[constants.IdentifierEmail.String()])
	_, ok = ids[oldEmail]
	require.False(t, ok)
	assertKratosHasIdentity(t, ids, constants.IdentifierPhone.String(), phone, true)
	assertKratosHasIdentity(t, ids, constants.IdentifierEmail.String(), newEmail, true)
	assertKratosHasIdentity(t, ids, constants.IdentifierEmail.String(), oldEmail, false)

	// Check: login with new email, old phone works, new email fails
	t.Run("login with new email works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, newEmail)
		require.Nil(t, err)
		require.NotNil(t, loginResp)
	})
	t.Run("login with old email fails", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, oldEmail)
		require.Nil(t, loginResp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not registered")
	})
	t.Run("login with phone still works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, phone)
		require.Nil(t, err)
		require.NotNil(t, loginResp)
	})
}

func TestIntegration_ChangeIdentifier_PhoneToPhone_LoginChecks(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-change-identifier-phone-to-phone")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	var globalUserID string
	oldPhone := "+84345381013"
	newPhone := "+84345381014"
	email := "oldemail@test.com"

	// Tenant is seeded by startPostgresAndBuildUCase

	// Register a phone identity
	registerResp, registerErr := ucase.Register(ctx, tenantID, "en", "", oldPhone)
	require.Nil(t, registerErr)
	require.NotNil(t, registerResp)

	verifyRegisterResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, registerResp.VerificationFlow.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyRegisterResp)
	globalUserID = verifyRegisterResp.User.GlobalUserID

	// Act: add an email identity
	addEmailResp, addEmailErr := ucase.AddNewIdentifier(ctx, tenantID, globalUserID, email, constants.IdentifierEmail.String())
	require.Nil(t, addEmailErr)
	require.NotNil(t, addEmailResp)

	verifyAddEmailResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, addEmailResp.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyAddEmailResp)

	// Act: change identifier phone→phone
	changeResp, changeErr := ucase.ChangeIdentifier(ctx, globalUserID, tenantID, verifyRegisterResp.User.ID, newPhone)
	require.Nil(t, changeErr)
	require.NotNil(t, changeResp)

	verifyChangeResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, changeResp.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyChangeResp)

	// Query identities in IAM, ensure the old phone identity is deleted on IAM side
	identities, ucaseErr := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, globalUserID, tenantID.String())
	require.Nil(t, ucaseErr)
	require.NotNil(t, identities)
	require.Equal(t, 2, len(identities))
	assertIAMHasIdentity(t, identities, constants.IdentifierEmail.String(), email)
	assertIAMHasIdentity(t, identities, constants.IdentifierPhone.String(), newPhone)

	//  query kratos, ensure the old phone identity is deleted on Kratos side
	servc := deps.kratosService.(*kratos_service.FakeKratosService)
	ids, _ := servc.GetIdentities(ctx, tenantID)
	new, ok := ids[newPhone]
	require.True(t, ok)
	require.Equal(t, newPhone, new.Traits.(map[string]interface{})[constants.IdentifierPhone.String()])
	_, ok = ids[oldPhone]
	require.False(t, ok)
	assertKratosHasIdentity(t, ids, constants.IdentifierPhone.String(), newPhone, true)
	assertKratosHasIdentity(t, ids, constants.IdentifierEmail.String(), email, true)
	assertKratosHasIdentity(t, ids, constants.IdentifierPhone.String(), oldPhone, false)

	// Check: login with new phone, old phone fails, email still works
	t.Run("login with new phone works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, newPhone)
		require.Nil(t, err)
		require.NotNil(t, loginResp)
	})

	t.Run("login with old phone fails", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, oldPhone)
		require.NotNil(t, err)
		require.Nil(t, loginResp)
		require.Contains(t, err.Error(), "not found")
	})

	t.Run("login with email still works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, email)
		require.Nil(t, err)
		require.NotNil(t, loginResp)
	})
}

func TestIntegration_ChangeIdentifier_PhoneToEmail_LoginChecks(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-change-identifier-phone-to-email")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	oldPhone := "+84345381013"
	newEmail := "newemail@test.com"

	// Tenant is seeded by startPostgresAndBuildUCase

	// Register a phone identity (only phone exists initially)
	registerResp, registerErr := ucase.Register(ctx, tenantID, "en", "", oldPhone)
	require.Nil(t, registerErr)
	require.NotNil(t, registerResp)

	verifyRegisterResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, registerResp.VerificationFlow.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyRegisterResp)
	globalUserID := verifyRegisterResp.User.GlobalUserID

	// Act: change identifier phone→email (since only one identifier exists, it should replace it)
	changeResp, changeErr := ucase.ChangeIdentifier(ctx, globalUserID, tenantID, verifyRegisterResp.User.ID, newEmail)
	require.Nil(t, changeErr)
	require.NotNil(t, changeResp)

	verifyChangeResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, changeResp.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyChangeResp)

	// Query identities in IAM, ensure there is ONLY the new email identity (old phone removed)
	identities, ucaseErr := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, globalUserID, tenantID.String())
	require.Nil(t, ucaseErr)
	require.NotNil(t, identities)
	require.Equal(t, 1, len(identities))
	require.Equal(t, constants.IdentifierEmail.String(), identities[0].Type)
	require.Equal(t, newEmail, identities[0].Value)

	// Query Kratos, ensure old phone is deleted and new email exists
	servc := deps.kratosService.(*kratos_service.FakeKratosService)
	ids, _ := servc.GetIdentities(ctx, tenantID)
	newId, ok := ids[newEmail]
	require.True(t, ok)
	require.Equal(t, newEmail, newId.Traits.(map[string]interface{})[constants.IdentifierEmail.String()])
	_, ok = ids[oldPhone]
	require.False(t, ok)

	// Check: login with new email works, old phone fails
	t.Run("login with new email works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, newEmail)
		require.Nil(t, err)
		require.NotNil(t, loginResp)
	})

	t.Run("login with old phone fails", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, oldPhone)
		require.NotNil(t, err)
		require.Nil(t, loginResp)
		require.Contains(t, err.Error(), "not found")
	})
}

// assertIAMHasIdentity asserts that the identity exists in IAMs
func assertIAMHasIdentity(t *testing.T, identities []*domain.UserIdentity, identifierType string, identifierValue string) {
	for _, identity := range identities {
		if identity.Type == identifierType && identity.Value == identifierValue {
			return
		}
	}
	require.Fail(t, "identity not found")
}

// assertKratosHasIdentity asserts that the identity exists in Kratos
func assertKratosHasIdentity(t *testing.T, identities map[string]*kratos.Identity, identifierType string, identifierValue string, hasIdentity bool) {
	for _, identity := range identities {
		if identity.Traits.(map[string]interface{})[identifierType] == identifierValue {
			if hasIdentity {
				return
			} else {
				require.Fail(t, "identity found but should not have")
			}
		}
	}
	if hasIdentity {
		require.Fail(t, "identity not found")
	}
}

func TestIntegration_Profile_ReturnsMappedFields(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-profile")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Tenant is seeded by startPostgresAndBuildUCase

	phone := "+84321234567"
	email := "profile@test.com"

	// Register via phone and verify -> creates user and session token
	reg, regErr := ucase.Register(ctx, tenantID, "en", "", phone)
	require.Nil(t, regErr)
	require.NotNil(t, reg)
	ver, verErr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, verErr)
	require.NotNil(t, ver)
	globalUserID := ver.User.GlobalUserID

	// Add email identifier and verify
	add, addErr := ucase.AddNewIdentifier(ctx, tenantID, globalUserID, email, constants.IdentifierEmail.String())
	require.Nil(t, addErr)
	require.NotNil(t, add)
	_, addVerErr := ucase.VerifyRegister(ctx, tenantID, add.FlowID, "000000")
	require.Nil(t, addVerErr)

	// Verify persistence in IAM
	identitiesIAM, iamErr := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, globalUserID, tenantID.String())
	require.Nil(t, iamErr)
	require.NotNil(t, identitiesIAM)
	assertIAMHasIdentity(t, identitiesIAM, constants.IdentifierPhone.String(), phone)
	assertIAMHasIdentity(t, identitiesIAM, constants.IdentifierEmail.String(), email)

	// Verify persistence in Kratos
	servc := deps.kratosService.(*kratos_service.FakeKratosService)
	ids, _ := servc.GetIdentities(ctx, tenantID)
	assertKratosHasIdentity(t, ids, constants.IdentifierPhone.String(), phone, true)
	assertKratosHasIdentity(t, ids, constants.IdentifierEmail.String(), email, true)

	// Call Profile with session token from the first verification
	ctxWithToken := context.WithValue(ctx, constants.SessionTokenKey, ver.SessionToken)
	resp, derr := ucase.Profile(ctxWithToken, tenantID)
	require.Nil(t, derr)
	require.NotNil(t, resp)
	// Should reflect DB-mapped email and phone regardless of Kratos traits
	require.Equal(t, email, resp.Email)
	require.Equal(t, phone, resp.Phone)
	require.Equal(t, ver.User.ID, resp.ID)
	require.Equal(t, ver.User.GlobalUserID, resp.GlobalUserID)
}

func TestIntegration_Profile_UnauthorizedWithoutToken(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, _, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-profile-unauth")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Tenant is seeded by startPostgresAndBuildUCase

	_, derr := ucase.Profile(ctx, tenantID)
	require.NotNil(t, derr)
	require.Equal(t, "MSG_UNAUTHORIZED", derr.Code)
}

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

// Force a failure inside bindIAMToRegistration by inserting a conflicting identity before verify
func TestIntegration_VerifyRegister_BindIAMToRegistration_FailsOnDuplicateIdentity(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, db, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-bind-fail")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	email := "duplicate-bind@test.com"

	// Start registration (creates challenge session with identifier)
	reg, derr := ucase.Register(ctx, tenantID, "en", email, "")
	require.Nil(t, derr)

	// Insert a pre-existing identity with same tenant/type/value to force unique violation
	preGU := &domain.GlobalUser{}
	require.NoError(t, adaptersrepo.NewGlobalUserRepository(db).Create(db, preGU))
	// Use InsertOnce to persist properly respecting constraints
	_, err := deps.userIdentityRepo.InsertOnceByKratosUserAndType(ctx, db, tenantID.String(), uuid.NewString(), preGU.ID, constants.IdentifierEmail.String(), email)
	require.NoError(t, err)

	// Now verify should hit bindIAMToRegistration -> identity insert fails -> wrapped MSG_IAM_REGISTRATION_FAILED
	auth, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, auth)
	require.NotNil(t, derr)
	require.Equal(t, "MSG_IAM_REGISTRATION_FAILED", derr.Code)
}

// Trigger rollbackKratosUpdateIdentifier by causing tx in bindIAMToUpdateIdentifier to fail due to type conflict
func TestIntegration_ChangeIdentifier_TriggersRollbackOnTxFailure(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-rollback-change")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Seed user with both email and phone so that changing identifier type collides (unique tenant+global_user+type)
	email := "both@test.com"
	phone := "+84320019999"

	reg, derr := ucase.Register(ctx, tenantID, "en", email, "")
	require.Nil(t, derr)
	ver, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	// Add a new identifier of type phone so the user has both types
	add, derr := ucase.AddNewIdentifier(ctx, tenantID, ver.User.GlobalUserID, phone, constants.IdentifierPhone.String())
	require.Nil(t, derr)
	_, derr = ucase.VerifyRegister(ctx, tenantID, add.FlowID, "000000")
	require.Nil(t, derr)

	// Prepare target phone value and pre-seed a DIFFERENT user with it to violate UNIQUE (tenant_id, type, value)
	newPhone := "+84320018888"
	// Pre-seed a DIFFERENT user with the target new phone value to violate UNIQUE (tenant_id, type, value)
	// when our change tries to update the phone identity value to this number.
	preReg, derr := ucase.Register(ctx, tenantID, "en", "", newPhone)
	require.Nil(t, derr)
	_, derr = ucase.VerifyRegister(ctx, tenantID, preReg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	// Now attempt to change identifier to another phone number which is already taken in tenant -> should fail fast with conflict
	chg, derr := ucase.ChangeIdentifier(ctx, ver.User.GlobalUserID, tenantID, ver.User.ID, newPhone)
	require.Nil(t, chg)
	require.NotNil(t, derr)
	require.Equal(t, "MSG_IDENTIFIER_ALREADY_EXISTS", derr.Code)
}

// Add another Login/RefreshToken path to cover username/lang traits present
func TestIntegration_LoginAndRefresh_WithUsernameAndLang(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-login-refresh")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	email := "userlang@test.com"

	// Register with lang included already; username will be set on login identity via traits
	reg, derr := ucase.Register(ctx, tenantID, "vi", email, "")
	require.Nil(t, derr)
	_, derr = ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	// Perform password login
	auth, derr := ucase.Login(ctx, tenantID, email, "pw")
	require.Nil(t, derr)
	require.NotNil(t, auth)
	require.True(t, auth.Active)
	// Lang should be populated from traits
	require.Equal(t, "vi", auth.User.Lang)

	// RefreshToken should return current session info using access token
	refreshed, derr := ucase.RefreshToken(ctx, tenantID, auth.SessionToken, "unused-refresh")
	require.Nil(t, derr)
	require.NotNil(t, refreshed)
	require.Equal(t, auth.User.Email, refreshed.User.Email)
	require.Equal(t, "vi", refreshed.User.Lang)
}

// UpdateLang validation: invalid lang and unsupported lang
func TestIntegration_UpdateLang_ValidationErrors(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-updatelang-validation")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Seed user by email
	email := "langval@test.com"
	reg, derr := ucase.Register(ctx, tenantID, "en", email, "")
	require.Nil(t, derr)
	ver, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	// Invalid lang (empty after normalize)
	derr = ucase.UpdateLang(ctx, tenantID, ver.User.ID, " ")
	require.NotNil(t, derr)
	require.Equal(t, "MSG_INVALID_LANG", derr.Code)

	// Unsupported lang
	derr = ucase.UpdateLang(ctx, tenantID, ver.User.ID, "zz")
	require.NotNil(t, derr)
	require.Equal(t, "MSG_UNSUPPORTED_LANG", derr.Code)
}

// ChallengeWithEmail + VerifyLogin success flow
func TestIntegration_ChallengeEmail_Then_VerifyLogin_Succeeds(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-email-login-success")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	email := "emaillogin@test.com"
	reg, derr := ucase.Register(ctx, tenantID, "en", email, "")
	require.Nil(t, derr)
	_, derr = ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	// Challenge with email
	chall, derr := ucase.ChallengeWithEmail(ctx, tenantID, email)
	require.Nil(t, derr)
	require.NotNil(t, chall)

	// VerifyLogin with the flow id
	auth, derr := ucase.VerifyLogin(ctx, tenantID, chall.FlowID, "000000")
	require.Nil(t, derr)
	require.NotNil(t, auth)
	require.True(t, auth.Active)
	require.Equal(t, email, auth.User.Email)
}

// ChangeIdentifier success: email -> phone
func TestIntegration_ChangeIdentifier_EmailToPhone_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-change-email-phone")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	email := "changeid@test.com"
	reg, derr := ucase.Register(ctx, tenantID, "en", email, "")
	require.Nil(t, derr)
	ver, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	// Change to phone
	newPhone := "+84321110001"
	chg, derr := ucase.ChangeIdentifier(ctx, ver.User.GlobalUserID, tenantID, ver.User.ID, newPhone)
	require.Nil(t, derr)
	require.NotNil(t, chg)

	// Verify change
	res, derr := ucase.VerifyRegister(ctx, tenantID, chg.FlowID, "000000")
	require.Nil(t, derr)
	require.NotNil(t, res)
	require.True(t, res.Active)
	require.Equal(t, newPhone, res.User.Phone)

	// Ensure login by new phone works
	auth, derr := ucase.Login(ctx, tenantID, newPhone, "any")
	require.Nil(t, derr)
	require.NotNil(t, auth)
	require.True(t, auth.Active)
}

// Covers: ChallengeVerification + VerifyIdentifier success and invalid code
func TestIntegration_VerifyIdentifier_Branches(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, _, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-verify-identifier-branches")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Use phone to avoid any email normalization issues
	phone := "+84321112223"
	reg, derr := ucase.Register(ctx, tenantID, "en", "", phone)
	require.Nil(t, derr)
	_, derr = ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)

	chall, derr := ucase.ChallengeVerification(ctx, tenantID, phone)
	require.Nil(t, derr)
	require.NotNil(t, chall)

	// Invalid code branch: configure fake to reject OTP (failed_challenge)
	fake := deps.kratosService.(*kratos_service.FakeKratosService)
	fake.SetFaults(kratos_service.Faults{RejectOTP: true})
	res1, derr := ucase.VerifyIdentifier(ctx, tenantID, chall.FlowID, "111111")
	require.NotNil(t, derr)
	require.Nil(t, res1)

	// Valid code branch: reset faults
	fake.SetFaults(kratos_service.Faults{})
	res2, derr := ucase.VerifyIdentifier(ctx, tenantID, chall.FlowID, "000000")
	require.Nil(t, derr)
	require.NotNil(t, res2)
	require.True(t, res2.Verified)
}

// Ensures Register cleans up orphan IAM identity (exists in DB but missing in Kratos)
func TestIntegration_Register_CleansUpOrphanIAM(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, _, deps, db, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "tenant-orphan-cleanup")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Allow flows without rate limits
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Seed an orphan identity in IAM (Kratos has no such identity)
	gu := &domain.GlobalUser{}
	require.NoError(t, adaptersrepo.NewGlobalUserRepository(db).Create(db, gu))

	phone := "+84300000001"
	orphanKratosID := uuid.NewString() // Not present in FakeKratosService

	inserted, err := deps.userIdentityRepo.InsertOnceByKratosUserAndType(ctx, db, tenantID.String(), orphanKratosID, gu.ID, constants.IdentifierPhone.String(), phone)
	require.NoError(t, err)
	require.True(t, inserted)

	// Attempt to register same phone — should clean up orphan and proceed
	reg, derr := ucase.Register(ctx, tenantID, "en", "", phone)
	require.Nil(t, derr)
	require.NotNil(t, reg)
	require.True(t, reg.VerificationNeeded)
	require.NotEmpty(t, reg.VerificationFlow.FlowID)

	// Verify registration completes successfully
	auth, derr := ucase.VerifyRegister(ctx, tenantID, reg.VerificationFlow.FlowID, "000000")
	require.Nil(t, derr)
	require.NotNil(t, auth)
	require.True(t, auth.Active)
	require.Equal(t, phone, auth.User.Phone)
}

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

// Registration persistence tests
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
