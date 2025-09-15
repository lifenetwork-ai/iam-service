//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIntegration_AdminUseCase_TenantManagement(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	adminUcase, _, _, _, container := startPostgresAndBuildAdminUCase(t, ctx, ctrl, "admin-tenant-test")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Tenant is seeded by startPostgresAndBuildAdminUCase

	t.Run("CreateTenant", func(t *testing.T) {
		newTenant, derr := adminUcase.CreateTenant(ctx, "test-tenant", "https://test.example.com", "https://admin.test.example.com")
		require.Nil(t, derr)
		require.NotNil(t, newTenant)
		require.Equal(t, "test-tenant", newTenant.Name)
		require.Equal(t, "https://test.example.com", newTenant.PublicURL)
		require.Equal(t, "https://admin.test.example.com", newTenant.AdminURL)
		require.NotEqual(t, uuid.Nil, newTenant.ID)
	})

	t.Run("ListTenants", func(t *testing.T) {
		resp, derr := adminUcase.ListTenants(ctx, 1, 10, "")
		require.Nil(t, derr)
		require.NotNil(t, resp)
		require.Greater(t, resp.TotalCount, int64(0))
		require.Greater(t, len(resp.Items), 0)
	})

	t.Run("GetTenantByID", func(t *testing.T) {
		// First create a tenant to get
		newTenant, derr := adminUcase.CreateTenant(ctx, "get-test-tenant", "https://get.test.com", "https://admin.get.test.com")
		require.Nil(t, derr)
		require.NotNil(t, newTenant)

		// Now get it by ID
		foundTenant, derr := adminUcase.GetTenantByID(ctx, newTenant.ID.String())
		require.Nil(t, derr)
		require.NotNil(t, foundTenant)
		require.Equal(t, newTenant.ID, foundTenant.ID)
		require.Equal(t, "get-test-tenant", foundTenant.Name)
	})

	t.Run("UpdateTenant", func(t *testing.T) {
		// First create a tenant to update
		newTenant, derr := adminUcase.CreateTenant(ctx, "update-test-tenant", "https://update.test.com", "https://admin.update.test.com")
		require.Nil(t, derr)
		require.NotNil(t, newTenant)

		// Update the tenant
		updatedTenant, derr := adminUcase.UpdateTenant(ctx, newTenant.ID.String(), "updated-tenant", "https://updated.test.com", "https://admin.updated.test.com")
		require.Nil(t, derr)
		require.NotNil(t, updatedTenant)
		require.Equal(t, newTenant.ID, updatedTenant.ID)
		require.Equal(t, "updated-tenant", updatedTenant.Name)
		require.Equal(t, "https://updated.test.com", updatedTenant.PublicURL)
		require.Equal(t, "https://admin.updated.test.com", updatedTenant.AdminURL)
	})

	t.Run("DeleteTenant", func(t *testing.T) {
		// First create a tenant to delete
		newTenant, derr := adminUcase.CreateTenant(ctx, "delete-test-tenant", "https://delete.test.com", "https://admin.delete.test.com")
		require.Nil(t, derr)
		require.NotNil(t, newTenant)

		// Delete the tenant
		deletedTenant, derr := adminUcase.DeleteTenant(ctx, newTenant.ID.String())
		require.Nil(t, derr)
		require.NotNil(t, deletedTenant)
		require.Equal(t, newTenant.ID, deletedTenant.ID)

		// Verify it's deleted by trying to get it
		_, derr = adminUcase.GetTenantByID(ctx, newTenant.ID.String())
		require.NotNil(t, derr)
		require.Equal(t, "MSG_TENANT_NOT_FOUND", derr.Code)
	})
}

func TestIntegration_AdminUseCase_AdminAccountManagement(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	adminUcase, _, _, _, container := startPostgresAndBuildAdminUCase(t, ctx, ctrl, "admin-account-test")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	t.Run("CreateAdminAccount", func(t *testing.T) {
		adminAccount, derr := adminUcase.CreateAdminAccount(ctx, "testadmin@example.com", "securepassword123", "admin")
		require.Nil(t, derr)
		require.NotNil(t, adminAccount)
		require.Equal(t, "testadmin@example.com", adminAccount.Username)
		require.Equal(t, "admin", adminAccount.Role)
		require.NotEqual(t, uuid.Nil, adminAccount.ID)
		require.NotNil(t, adminAccount.CreatedAt)
		require.NotNil(t, adminAccount.UpdatedAt)
	})

	t.Run("GetAdminAccountByUsername", func(t *testing.T) {
		// First create an admin account
		adminAccount, derr := adminUcase.CreateAdminAccount(ctx, "getadmin@example.com", "password123", "admin")
		require.Nil(t, derr)
		require.NotNil(t, adminAccount)

		// Now get it by username
		foundAccount, derr := adminUcase.GetAdminAccountByUsername(ctx, "getadmin@example.com")
		require.Nil(t, derr)
		require.NotNil(t, foundAccount)
		require.Equal(t, adminAccount.ID, foundAccount.ID)
		require.Equal(t, "getadmin@example.com", foundAccount.Username)
		require.Equal(t, "admin", foundAccount.Role)
	})

	t.Run("CreateDuplicateAdminAccount", func(t *testing.T) {
		// Create first admin account
		_, derr := adminUcase.CreateAdminAccount(ctx, "duplicate@example.com", "password123", "admin")
		require.Nil(t, derr)

		// Try to create another with same username
		_, derr = adminUcase.CreateAdminAccount(ctx, "duplicate@example.com", "differentpassword", "admin")
		require.NotNil(t, derr)
		require.Equal(t, "MSG_ADMIN_USERNAME_EXISTS", derr.Code)
	})
}

func TestIntegration_AdminUseCase_UserIdentityManagement(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ucase, adminUcase, deps, _, tenantID, container := startPostgresAndBuildUCase(t, ctx, ctrl, "admin-identity-test")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	// Tenant is seeded by startPostgresAndBuildUCase
	// Allow identity flows without rate limits
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("CheckIdentifierAdmin", func(t *testing.T) {
		// Test checking non-existent identifier
		exists, identifierType, derr := adminUcase.CheckIdentifierAdmin(ctx, tenantID, "nonexistent@example.com")
		require.Nil(t, derr)
		require.False(t, exists)
		// We expect the inferred type even when the identifier does not exist
		require.Equal(t, "email", identifierType)

		// Create a user first via registration to properly set up the identity
		regResp, derr := ucase.Register(ctx, tenantID, "en", "testuser@example.com", "")
		require.Nil(t, derr)
		require.NotNil(t, regResp)

		// Verify the registration
		_, derr = ucase.VerifyRegister(ctx, tenantID, regResp.VerificationFlow.FlowID, "000000")
		require.Nil(t, derr)

		// Now check if identifier exists
		exists, identifierType, derr = adminUcase.CheckIdentifierAdmin(ctx, tenantID, "testuser@example.com")
		require.Nil(t, derr)
		require.True(t, exists)
		require.Equal(t, "email", identifierType)
	})

	t.Run("AddIdentifierAdmin", func(t *testing.T) {
		// Create a user via registration first
		regResp, derr := ucase.Register(ctx, tenantID, "en", "existing@example.com", "")
		require.Nil(t, derr)
		require.NotNil(t, regResp)

		// Verify the registration to get the user
		verifyResp, derr := ucase.VerifyRegister(ctx, tenantID, regResp.VerificationFlow.FlowID, "000000")
		require.Nil(t, derr)
		require.NotNil(t, verifyResp)

		// Add identifier via admin
		req := dto.AdminAddIdentifierPayloadDTO{
			ExistingIdentifier: "existing@example.com",
			NewIdentifier:      "+84312345678", // Valid phone format (VN)
		}

		resp, derr := adminUcase.AddIdentifierAdmin(ctx, tenantID, req)
		require.Nil(t, derr)
		require.NotNil(t, resp)
		require.Equal(t, verifyResp.User.GlobalUserID, resp.GlobalUserID)
		require.Equal(t, "+84312345678", resp.Identifier)

		// Verify it was created in IAM
		identities, err := deps.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, verifyResp.User.GlobalUserID, tenantID.String())
		require.Nil(t, err)
		require.Len(t, identities, 2) // existing + new
	})
}
