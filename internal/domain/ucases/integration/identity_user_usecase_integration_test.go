//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	kratos "github.com/ory/kratos-client-go"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/lifenetwork-ai/iam-service/constants"
	kratos_service "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/stretchr/testify/require"
)

func applyPostgresScripts(db *gorm.DB) {
	_, thisFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(thisFile)
	// internal/domain/ucases/integration -> internal/adapters/postgres/scripts
	scriptsDir := filepath.Join(testDir, "../../../adapters/postgres/scripts")
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		panic(fmt.Errorf("read scripts dir: %w", err))
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)
	for _, name := range files {
		path := filepath.Join(scriptsDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			panic(fmt.Errorf("read %s: %w", name, err))
		}
		if err := db.Exec(string(content)).Error; err != nil {
			panic(fmt.Errorf("execute %s: %w", name, err))
		}
	}
}

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
	return
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
