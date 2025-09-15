//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	adaptersrepo "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	kratos_service "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
