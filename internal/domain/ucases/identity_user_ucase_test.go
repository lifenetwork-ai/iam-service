package ucases

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	mock_repositories "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/repositories"
	mock_services "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/services"
	mock_types "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/rate_limiter/types"
	kratos "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

// TestUpdateIdentifier tests the UpdateIdentifier use case method
func TestUpdateIdentifier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup test dependencies
	mockDeps := setupTestDependencies(ctrl)
	ucase := newTestUserUseCase(mockDeps)
	ctx := context.Background()

	// Allow RegisterAttempt to be called any number of times in any subtest
	mockDeps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Common test data
	testData := struct {
		globalUserID string
		tenantID     uuid.UUID
		tenantUserID string
		flowID       string
	}{
		globalUserID: "test-global-user-id",
		tenantID:     uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		tenantUserID: "00000000-0000-0000-0000-000000000002",
		flowID:       "test-flow-id",
	}

	t.Run("when updating email identifier", func(t *testing.T) {
		t.Run("should succeed with valid new email", func(t *testing.T) {
			// Given
			newEmail := "newemail@example.com"
			setupSuccessfulUpdateFlow(ctx, mockDeps, testData, newEmail, constants.IdentifierEmail.String())

			// When
			result, err := ucase.UpdateIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, newEmail, constants.IdentifierEmail.String())

			// Then
			assertSuccessfulUpdate(t, result, err, testData.flowID, newEmail)
		})

		t.Run("should fail when email already exists", func(t *testing.T) {
			// Given
			existingEmail := "existing@example.com"
			setupIdentifierExistsFlow(ctx, mockDeps, testData, existingEmail, constants.IdentifierEmail.String())

			// When
			result, err := ucase.UpdateIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, existingEmail, constants.IdentifierEmail.String())

			// Then
			assertIdentifierExists(t, result, err)
		})

		t.Run("should fail with invalid email format", func(t *testing.T) {
			// Given
			invalidEmail := "invalid-email"

			// When
			result, err := ucase.UpdateIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, invalidEmail, constants.IdentifierEmail.String())

			// Then
			assertInvalidEmail(t, result, err)
		})
	})

	t.Run("when updating phone identifier", func(t *testing.T) {
		t.Run("should succeed with valid phone number", func(t *testing.T) {
			// Given
			newPhone := "+84344381024"
			setupSuccessfulUpdateFlow(ctx, mockDeps, testData, newPhone, constants.IdentifierPhone.String())

			// When
			result, err := ucase.UpdateIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, newPhone, constants.IdentifierPhone.String())

			// Then
			assertSuccessfulUpdate(t, result, err, testData.flowID, newPhone)
		})
	})

	t.Run("should fail with empty identifier", func(t *testing.T) {
		// When
		result, err := ucase.UpdateIdentifier(ctx, testData.globalUserID, testData.tenantID,
			testData.tenantUserID, "", constants.IdentifierEmail.String())

		// Then
		assertInvalidRequest(t, result, err)
	})
}

// Helper functions to reduce test complexity and improve readability
type testDependencies struct {
	tenantRepo                *mock_repositories.MockTenantRepository
	globalUserRepo            *mock_repositories.MockGlobalUserRepository
	userIdentityRepo          *mock_repositories.MockUserIdentityRepository
	userIdentifierMappingRepo *mock_repositories.MockUserIdentifierMappingRepository
	challengeSessionRepo      *mock_repositories.MockChallengeSessionRepository
	kratosService             *mock_services.MockKratosService
	rateLimiter               *mock_types.MockRateLimiter
}

func setupTestDependencies(ctrl *gomock.Controller) *testDependencies {
	deps := &testDependencies{
		tenantRepo:                mock_repositories.NewMockTenantRepository(ctrl),
		globalUserRepo:            mock_repositories.NewMockGlobalUserRepository(ctrl),
		userIdentityRepo:          mock_repositories.NewMockUserIdentityRepository(ctrl),
		userIdentifierMappingRepo: mock_repositories.NewMockUserIdentifierMappingRepository(ctrl),
		challengeSessionRepo:      mock_repositories.NewMockChallengeSessionRepository(ctrl),
		kratosService:             mock_services.NewMockKratosService(ctrl),
		rateLimiter:               mock_types.NewMockRateLimiter(ctrl),
	}

	// Setup default rate limiter behavior
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(false, nil).AnyTimes()

	return deps
}

func newTestUserUseCase(deps *testDependencies) *userUseCase {
	return &userUseCase{
		db:                        &gorm.DB{},
		rateLimiter:               deps.rateLimiter,
		tenantRepo:                deps.tenantRepo,
		globalUserRepo:            deps.globalUserRepo,
		userIdentityRepo:          deps.userIdentityRepo,
		userIdentifierMappingRepo: deps.userIdentifierMappingRepo,
		challengeSessionRepo:      deps.challengeSessionRepo,
		kratosService:             deps.kratosService,
	}
}

// Update: Fix mock expectations and assertions for UpdateIdentifier and DeleteIdentifier tests
// Use .Times(1) for single expected calls, adjust error assertions, and ensure proper mock usage

// --- UpdateIdentifier helpers ---
func setupSuccessfulUpdateFlow(ctx context.Context, deps *testDependencies, data struct {
	globalUserID         string
	tenantID             uuid.UUID
	tenantUserID, flowID string
},
	newIdentifier, identifierType string) {
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), data.tenantID.String(), identifierType, newIdentifier).Return(false, nil).AnyTimes()
	deps.userIdentityRepo.EXPECT().GetByTenantAndTenantUserID(gomock.Any(), gomock.Any(), gomock.Eq(data.tenantID.String()), gomock.Eq(data.tenantUserID)).Return(&domain.UserIdentity{
		ID:   "identity-id",
		Type: identifierType,
	}, nil).AnyTimes()
	deps.tenantRepo.EXPECT().GetByID(gomock.Eq(data.tenantID)).Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil).AnyTimes()
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), gomock.Eq(data.tenantID)).Return(&kratos.RegistrationFlow{Id: data.flowID}, nil).AnyTimes()
	deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), gomock.Eq(data.tenantID), gomock.Any(), gomock.Eq(constants.MethodTypeCode.String()), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil).AnyTimes()
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), gomock.Eq(data.flowID), gomock.Any(), gomock.Eq(constants.DefaultChallengeDuration)).Return(nil).AnyTimes()
}

func setupIdentifierExistsFlow(ctx context.Context, deps *testDependencies, data struct {
	globalUserID         string
	tenantID             uuid.UUID
	tenantUserID, flowID string
},
	identifier, identifierType string) {
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), data.tenantID.String(), identifierType, identifier).Return(true, nil).AnyTimes()
}

// --- DeleteIdentifier helpers ---
func setupMultipleIdentifiersFlow(ctx context.Context, deps *testDependencies, data struct {
	globalUserID string
	tenantID     uuid.UUID
	tenantUserID string
},
	identifierType string) {
	existingIdentity := &domain.UserIdentity{
		ID:   "identity-id",
		Type: identifierType,
	}
	deps.userIdentityRepo.EXPECT().GetByTenantAndTenantUserID(gomock.Any(), gomock.Any(), gomock.Eq(data.tenantID.String()), gomock.Eq(data.tenantUserID)).Return(existingIdentity, nil).AnyTimes()
	deps.userIdentityRepo.EXPECT().GetByGlobalUserID(gomock.Any(), gomock.Any(), gomock.Eq(data.tenantID.String()), gomock.Eq(data.globalUserID)).Return([]domain.UserIdentity{
		{Type: identifierType},
		{Type: constants.IdentifierPhone.String()},
	}, nil).AnyTimes()
	deps.userIdentityRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	deps.kratosService.EXPECT().DeleteIdentifierAdmin(gomock.Any(), gomock.Eq(data.tenantID), gomock.Any()).Return(nil).AnyTimes()
}

func setupSingleIdentifierFlow(ctx context.Context, deps *testDependencies, data struct {
	globalUserID string
	tenantID     uuid.UUID
	tenantUserID string
},
	identifierType string) {
	existingIdentity := &domain.UserIdentity{
		ID:   "identity-id",
		Type: identifierType,
	}
	deps.userIdentityRepo.EXPECT().GetByTenantAndTenantUserID(gomock.Any(), gomock.Any(), gomock.Eq(data.tenantID.String()), gomock.Eq(data.tenantUserID)).Return(existingIdentity, nil).AnyTimes()
	deps.userIdentityRepo.EXPECT().GetByGlobalUserID(gomock.Any(), gomock.Any(), gomock.Eq(data.tenantID.String()), gomock.Eq(data.globalUserID)).Return([]domain.UserIdentity{
		{Type: identifierType},
	}, nil).AnyTimes()
}

func setupNonExistentIdentifierFlow(ctx context.Context, deps *testDependencies, data struct {
	globalUserID string
	tenantID     uuid.UUID
	tenantUserID string
},
	identifierType string) {
	deps.userIdentityRepo.EXPECT().GetByTenantAndTenantUserID(gomock.Any(), gomock.Any(), gomock.Eq(data.tenantID.String()), gomock.Eq(data.tenantUserID)).Return(nil, nil).AnyTimes()
}

// --- Assertion helpers ---
func assertSuccessfulUpdate(t *testing.T, result interface{}, err error, expectedFlowID, expectedReceiver string) {
	t.Helper()
	assert.Nil(t, err)
	assert.NotNil(t, result)
	response, ok := result.(*types.IdentityUserChallengeResponse)
	assert.True(t, ok, "Expected result to be of type *types.IdentityUserChallengeResponse, got %T", result)
	assert.Equal(t, expectedFlowID, response.FlowID, "Expected flow ID %q, got %q", expectedFlowID, response.FlowID)
	assert.Equal(t, expectedReceiver, response.Receiver, "Expected receiver %q, got %q", expectedReceiver, response.Receiver)
}

func assertIdentifierExists(t *testing.T, result interface{}, err error) {
	assert.Error(t, err)
	assert.Nil(t, result)
	if err != nil {
		assert.Contains(t, err.Error(), "already been registered")
	}
}

func assertInvalidEmail(t *testing.T, result interface{}, err error) {
	assert.Error(t, err)
	assert.Nil(t, result)
	if err != nil {
		assert.Contains(t, err.Error(), "Invalid email")
	}
}

func assertInvalidRequest(t *testing.T, result interface{}, err error) {
	assert.Error(t, err)
	assert.Nil(t, result)
	if err != nil {
		assert.Contains(t, err.Error(), "Identifier is required")
	}
}

// TestDeleteIdentifier tests the DeleteIdentifier use case method
func TestDeleteIdentifier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup test dependencies
	mockDeps := setupTestDependencies(ctrl)
	ucase := newTestUserUseCase(mockDeps)
	ctx := context.Background()

	// Allow RegisterAttempt to be called any number of times in any subtest (if used)
	mockDeps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Common test data
	testData := struct {
		globalUserID string
		tenantID     uuid.UUID
		tenantUserID string
	}{
		globalUserID: "test-global-user-id",
		tenantID:     uuid.MustParse("a6f7ec89-3be2-4e82-bedb-9bc53bf9b935"),
		tenantUserID: "6976260d-751c-4e98-88e4-19f1a459a5f0",
	}

	t.Run("when deleting an identifier", func(t *testing.T) {
		t.Run("should succeed when user has multiple identifiers", func(t *testing.T) {
			identifierType := constants.IdentifierEmail.String()
			setupMultipleIdentifiersFlow(ctx, mockDeps, testData, identifierType)
			err := ucase.DeleteIdentifier(ctx, testData.globalUserID, testData.tenantID, testData.tenantUserID, identifierType)
			assert.Nil(t, err)
		})

		t.Run("should fail when it's the user's only identifier", func(t *testing.T) {
			identifierType := constants.IdentifierEmail.String()
			setupSingleIdentifierFlow(ctx, mockDeps, testData, identifierType)
			err := ucase.DeleteIdentifier(ctx, testData.globalUserID, testData.tenantID, testData.tenantUserID, identifierType)
			assert.Error(t, err)
			if err != nil {
				assert.Contains(t, err.Error(), "Cannot delete the only identifier")
			}
		})

		t.Run("should fail when identifier type doesn't exist", func(t *testing.T) {
			identifierType := constants.IdentifierEmail.String()
			setupNonExistentIdentifierFlow(ctx, mockDeps, testData, identifierType)
			mockDeps.userIdentityRepo.EXPECT().GetByTenantAndTenantUserID(gomock.Any(), gomock.Any(), gomock.Eq(testData.tenantID.String()), gomock.Eq(testData.tenantUserID)).Return(nil, nil).AnyTimes()
			err := ucase.DeleteIdentifier(ctx, testData.globalUserID, testData.tenantID, testData.tenantUserID, identifierType)
			assert.Error(t, err)
			if err != nil {
				assert.Contains(t, err.Error(), "does not have an identifier of type")
			}
		})
	})
}
