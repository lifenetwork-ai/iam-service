package ucases

import (
	"context"
	"testing"
	"time"

	kratos "github.com/ory/kratos-client-go"
	"gorm.io/driver/sqlite"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	mock_repositories "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/repositories"
	mock_services "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/services"
	mock_types "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/rate_limiter/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

// TestChangeIdentifier tests the ChangeIdentifier use case method
func TestChangeIdentifier(t *testing.T) {
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
			result, err := ucase.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierEmail.String(), newEmail, constants.IdentifierEmail.String())

			// Then
			assertSuccessfulUpdate(t, result, err, testData.flowID, newEmail)
		})

		t.Run("should fail when email already exists", func(t *testing.T) {
			// Given
			existingEmail := "existing@example.com"
			setupIdentifierExistsFlow(ctx, mockDeps, testData, existingEmail, constants.IdentifierEmail.String())

			// When
			result, err := ucase.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierEmail.String(), existingEmail, constants.IdentifierEmail.String())

			// Then
			assertIdentifierExists(t, result, err)
		})

		t.Run("should fail with invalid email format", func(t *testing.T) {
			// Given
			invalidEmail := "invalid-email"

			// When
			result, err := ucase.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierEmail.String(), invalidEmail, constants.IdentifierEmail.String())

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
			result, err := ucase.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierPhone.String(), newPhone, constants.IdentifierPhone.String())

			// Then
			assertSuccessfulUpdate(t, result, err, testData.flowID, newPhone)
		})

		// New: invalid phone format
		t.Run("should fail with invalid phone format", func(t *testing.T) {
			invalidPhone := "12345abc"
			result, err := ucase.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierPhone.String(), invalidPhone, constants.IdentifierPhone.String())
			assert.Error(t, err)
			assert.Nil(t, result)
			if err != nil {
				assert.Contains(t, err.Error(), "Invalid phone number")
			}
		})
	})

	t.Run("should fail with empty identifier", func(t *testing.T) {
		// When
		result, err := ucase.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
			testData.tenantUserID, constants.IdentifierEmail.String(), "", constants.IdentifierEmail.String())

		// Then
		assertInvalidRequest(t, result, err)
	})

	// Error-path tests for ChangeIdentifier
	t.Run("error paths", func(t *testing.T) {
		baseNew := "newemail@example.com"

		t.Run("exists lookup error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			md := setupTestDependencies(ctrl)
			uc := newTestUserUseCase(md)
			md.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			md.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), testData.tenantID.String(), constants.IdentifierEmail.String(), baseNew).Return(false, assert.AnError)
			result, err := uc.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierEmail.String(), baseNew, constants.IdentifierEmail.String())
			assert.Error(t, err)
			assert.Nil(t, result)
			if err != nil {
				assert.Contains(t, err.Error(), "Failed to check existing identifier")
			}
		})

		t.Run("list identities error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			md := setupTestDependencies(ctrl)
			uc := newTestUserUseCase(md)
			md.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			md.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), testData.tenantID.String(), constants.IdentifierEmail.String(), baseNew).Return(false, nil)
			md.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), testData.tenantID.String(), testData.tenantUserID).Return(nil, assert.AnError)
			result, err := uc.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierEmail.String(), baseNew, constants.IdentifierEmail.String())
			assert.Error(t, err)
			assert.Nil(t, result)
			if err != nil {
				assert.Contains(t, err.Error(), "Failed to check user identities")
			}
		})

		t.Run("kratos init error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			md := setupTestDependencies(ctrl)
			uc := newTestUserUseCase(md)
			md.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			md.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), testData.tenantID.String(), constants.IdentifierEmail.String(), baseNew).Return(false, nil)
			md.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), testData.tenantID.String(), testData.tenantUserID).Return([]*domain.UserIdentity{{ID: "id1", Type: constants.IdentifierEmail.String()}}, nil)
			md.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), testData.tenantID).Return(nil, assert.AnError)
			result, err := uc.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierEmail.String(), baseNew, constants.IdentifierEmail.String())
			assert.Error(t, err)
			assert.Nil(t, result)
			if err != nil {
				assert.Contains(t, err.Error(), "Failed to initialize registration flow")
			}
		})

		t.Run("tenant get error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			md := setupTestDependencies(ctrl)
			uc := newTestUserUseCase(md)
			md.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			md.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), testData.tenantID.String(), constants.IdentifierEmail.String(), baseNew).Return(false, nil)
			md.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), testData.tenantID.String(), testData.tenantUserID).Return([]*domain.UserIdentity{{ID: "id1", Type: constants.IdentifierEmail.String()}}, nil)
			md.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), testData.tenantID).Return(&kratos.RegistrationFlow{Id: testData.flowID}, nil)
			md.tenantRepo.EXPECT().GetByID(testData.tenantID).Return(nil, assert.AnError)
			result, err := uc.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierEmail.String(), baseNew, constants.IdentifierEmail.String())
			assert.Error(t, err)
			assert.Nil(t, result)
			if err != nil {
				assert.Contains(t, err.Error(), "Failed to get tenant")
			}
		})

		t.Run("kratos submit error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			md := setupTestDependencies(ctrl)
			uc := newTestUserUseCase(md)
			md.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			md.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), testData.tenantID.String(), constants.IdentifierEmail.String(), baseNew).Return(false, nil)
			md.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), testData.tenantID.String(), testData.tenantUserID).Return([]*domain.UserIdentity{{ID: "id1", Type: constants.IdentifierEmail.String()}}, nil)
			flow := &kratos.RegistrationFlow{Id: testData.flowID}
			md.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), testData.tenantID).Return(flow, nil)
			md.tenantRepo.EXPECT().GetByID(testData.tenantID).Return(&domain.Tenant{ID: testData.tenantID, Name: "tenant-name"}, nil)
			md.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), testData.tenantID, flow, gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
			result, err := uc.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierEmail.String(), baseNew, constants.IdentifierEmail.String())
			assert.Error(t, err)
			assert.Nil(t, result)
			if err != nil {
				assert.Contains(t, err.Error(), "Registration failed")
			}
		})

		t.Run("save challenge error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			md := setupTestDependencies(ctrl)
			uc := newTestUserUseCase(md)
			md.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			md.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), testData.tenantID.String(), constants.IdentifierEmail.String(), baseNew).Return(false, nil)
			md.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), testData.tenantID.String(), testData.tenantUserID).Return([]*domain.UserIdentity{{ID: "id1", Type: constants.IdentifierEmail.String()}}, nil)
			flow := &kratos.RegistrationFlow{Id: testData.flowID}
			md.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), testData.tenantID).Return(flow, nil)
			md.tenantRepo.EXPECT().GetByID(testData.tenantID).Return(&domain.Tenant{ID: testData.tenantID, Name: "tenant-name"}, nil)
			md.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), testData.tenantID, flow, gomock.Any(), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
			md.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), testData.flowID, gomock.Any(), gomock.Any()).Return(assert.AnError)
			result, err := uc.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
				testData.tenantUserID, constants.IdentifierEmail.String(), baseNew, constants.IdentifierEmail.String())
			assert.Error(t, err)
			assert.Nil(t, result)
			if err != nil {
				assert.Contains(t, err.Error(), "Failed to save challenge session")
			}
		})
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

// --- ChangeIdentifier helpers ---
func setupSuccessfulUpdateFlow(
	ctx context.Context,
	deps *testDependencies,
	data struct {
		globalUserID         string
		tenantID             uuid.UUID
		tenantUserID, flowID string
	},
	newIdentifier, identifierType string,
) {
	// New identifier must not exist
	deps.userIdentityRepo.EXPECT().
		ExistsWithinTenant(gomock.Any(), data.tenantID.String(), identifierType, newIdentifier).
		Return(false, nil).
		Times(1)

	// List must return non-empty slice so ChangeIdentifier does not fail
	deps.userIdentityRepo.EXPECT().
		ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).
		Return([]*domain.UserIdentity{
			{ID: "identity-id", Type: identifierType},
			{ID: "other-id", Type: map[bool]string{true: constants.IdentifierEmail.String(), false: constants.IdentifierPhone.String()}[identifierType != constants.IdentifierEmail.String()]},
		}, nil).
		AnyTimes()

	// Tenant lookup
	deps.tenantRepo.EXPECT().
		GetByID(gomock.Eq(data.tenantID)).
		Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil).
		AnyTimes()

	// KratosService mocks
	flow := &kratos.RegistrationFlow{Id: data.flowID}
	deps.kratosService.EXPECT().
		InitializeRegistrationFlow(gomock.Any(), gomock.Eq(data.tenantID)).
		Return(flow, nil).
		AnyTimes()
	deps.kratosService.EXPECT().
		SubmitRegistrationFlow(gomock.Any(), gomock.Eq(data.tenantID), gomock.Any(), gomock.Eq("code"), gomock.Any()).
		Return(&kratos.SuccessfulNativeRegistration{}, nil).
		AnyTimes()

	// ChallengeSessionRepository mock
	deps.challengeSessionRepo.EXPECT().
		SaveChallenge(gomock.Any(), gomock.Eq(data.flowID), gomock.Any(), gomock.Eq(5*time.Minute)).
		Return(nil).
		AnyTimes()
}

func setupIdentifierExistsFlow(ctx context.Context, deps *testDependencies, data struct {
	globalUserID         string
	tenantID             uuid.UUID
	tenantUserID, flowID string
},
	identifier, identifierType string) {
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), data.tenantID.String(), identifierType, identifier).Return(true, nil).AnyTimes()
	deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), gomock.Eq(data.tenantID.String()), gomock.Eq(data.tenantUserID)).Return([]*domain.UserIdentity{
		{ID: "identity-id", Type: identifierType},
	}, nil).AnyTimes()
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
	// DeleteIdentifier uses ListByTenantAndTenantUserID; return two identifiers
	deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), gomock.Eq(data.tenantID.String()), gomock.Eq(data.tenantUserID)).Return([]*domain.UserIdentity{
		{ID: existingIdentity.ID, Type: identifierType},
		{ID: "other-id", Type: map[bool]string{true: constants.IdentifierEmail.String(), false: constants.IdentifierPhone.String()}[identifierType != constants.IdentifierEmail.String()]},
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
	// Only one identifier present
	deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), gomock.Eq(data.tenantID.String()), gomock.Eq(data.tenantUserID)).Return([]*domain.UserIdentity{
		existingIdentity,
	}, nil).AnyTimes()
}

func setupNonExistentIdentifierFlow(ctx context.Context, deps *testDependencies, data struct {
	globalUserID string
	tenantID     uuid.UUID
	tenantUserID string
},
	identifierType string) {
	// No identifiers present
	deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), gomock.Eq(data.tenantID.String()), gomock.Eq(data.tenantUserID)).Return([]*domain.UserIdentity{}, nil).AnyTimes()
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
			err := ucase.DeleteIdentifier(ctx, testData.globalUserID, testData.tenantID, testData.tenantUserID, identifierType)
			assert.Error(t, err)
			if err != nil {
				assert.Contains(t, err.Error(), "does not have an identifier of type")
			}
		})

		t.Run("when user has multiple identifiers", func(t *testing.T) {
			// Common expectation: user has both email + phone
			mockDeps.userIdentityRepo.EXPECT().
				ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), gomock.Eq(testData.tenantID.String()), gomock.Eq(testData.tenantUserID)).
				Return([]*domain.UserIdentity{
					{ID: "id1", Type: constants.IdentifierEmail.String()},
					{ID: "id2", Type: constants.IdentifierPhone.String()},
				}, nil).AnyTimes()

			t.Run("should allow replacing same type", func(t *testing.T) {
				// replacing email -> new email should succeed
				// Use the struct with flowID for setupSuccessfulUpdateFlow
				testDataWithFlow := struct {
					globalUserID string
					tenantID     uuid.UUID
					tenantUserID string
					flowID       string
				}{
					globalUserID: testData.globalUserID,
					tenantID:     testData.tenantID,
					tenantUserID: testData.tenantUserID,
					flowID:       "test-flow-id",
				}
				// Re-configure the ExistsWithinTenant for this specific ChangeIdentifier scenario only
				setupSuccessfulUpdateFlow(ctx, mockDeps, testDataWithFlow, "newemail@example.com", constants.IdentifierEmail.String())
				result, err := ucase.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
					testData.tenantUserID, constants.IdentifierEmail.String(), "newemail@example.com", constants.IdentifierEmail.String())
				assertSuccessfulUpdate(t, result, err, testDataWithFlow.flowID, "newemail@example.com")
			})

			t.Run("should fail when replacing with different type", func(t *testing.T) {
				// Expect ExistsWithinTenant pre-check for the new phone identifier to pass (not existing)
				mockDeps.userIdentityRepo.EXPECT().
					ExistsWithinTenant(gomock.Any(), testData.tenantID.String(), constants.IdentifierPhone.String(), "+84344381024").
					Return(false, nil)

				// trying to replace email with phone using a valid phone number to reach rule check
				result, err := ucase.ChangeIdentifier(ctx, testData.globalUserID, testData.tenantID,
					testData.tenantUserID, constants.IdentifierEmail.String(), "+84344381024", constants.IdentifierPhone.String())

				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "cross-type change not allowed")
			})
		})

	})

	// Error paths for DeleteIdentifier
	t.Run("repo list error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		md := setupTestDependencies(ctrl)
		uc := newTestUserUseCase(md)
		md.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), testData.tenantID.String(), testData.tenantUserID).Return(nil, assert.AnError)
		err := uc.DeleteIdentifier(ctx, testData.globalUserID, testData.tenantID, testData.tenantUserID, constants.IdentifierEmail.String())
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "Failed to get user identifiers")
		}
	})

	t.Run("repo delete error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		md := setupTestDependencies(ctrl)
		uc := newTestUserUseCase(md)
		md.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), testData.tenantID.String(), testData.tenantUserID).Return([]*domain.UserIdentity{{ID: "id1", Type: constants.IdentifierEmail.String()}, {ID: "id2", Type: constants.IdentifierPhone.String()}}, nil)
		md.userIdentityRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(assert.AnError)
		err := uc.DeleteIdentifier(ctx, testData.globalUserID, testData.tenantID, testData.tenantUserID, constants.IdentifierEmail.String())
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "Failed to delete identifier")
		}
	})

	t.Run("kratos delete error is logged only", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		md := setupTestDependencies(ctrl)
		uc := newTestUserUseCase(md)
		md.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), testData.tenantID.String(), testData.tenantUserID).Return([]*domain.UserIdentity{{ID: "id1", Type: constants.IdentifierEmail.String()}, {ID: "id2", Type: constants.IdentifierPhone.String()}}, nil)
		md.userIdentityRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
		md.kratosService.EXPECT().DeleteIdentifierAdmin(gomock.Any(), testData.tenantID, gomock.Any()).Return(assert.AnError)
		err := uc.DeleteIdentifier(ctx, testData.globalUserID, testData.tenantID, testData.tenantUserID, constants.IdentifierEmail.String())
		assert.Nil(t, err)
	})
}

// Integration-style: ChangeIdentifier -> VerifyRegister successful flow
func TestChangeIdentifierThenVerifyRegister_Success(t *testing.T) {
	t.Skip("covered by isolated integration tests")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup test dependencies
	mockDeps := setupTestDependencies(ctrl)
	ucase := newTestUserUseCase(mockDeps)
	// Use in-memory sqlite so gorm.Transaction works
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	// Allow rate limiter default behaviors
	mockDeps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Common test data
	data := struct {
		globalUserID  string
		tenantID      uuid.UUID
		tenantUserID  string
		flowID        string
		newIdentifier string
	}{
		globalUserID:  "global-user-1",
		tenantID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		tenantUserID:  "00000000-0000-0000-0000-00000000aaaa",
		flowID:        "flow-123",
		newIdentifier: "newemail@example.com",
	}

	// ChangeIdentifier expectations
	// new identifier must not exist
	mockDeps.userIdentityRepo.EXPECT().
		ExistsWithinTenant(gomock.Any(), data.tenantID.String(), constants.IdentifierEmail.String(), data.newIdentifier).
		Return(false, nil)
	// user currently has email identity (and at least one more to allow replace same type)
	mockDeps.userIdentityRepo.EXPECT().
		ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), data.tenantID.String(), data.tenantUserID).
		Return([]*domain.UserIdentity{{ID: "identity-email-id", Type: constants.IdentifierEmail.String()}}, nil)
	// tenant lookup
	mockDeps.tenantRepo.EXPECT().GetByID(data.tenantID).Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	// kratos init + submit to trigger OTP
	regFlow := &kratos.RegistrationFlow{Id: data.flowID}
	mockDeps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), data.tenantID).Return(regFlow, nil)
	mockDeps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), data.tenantID, regFlow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
	// save challenge session
	mockDeps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), data.flowID, gomock.Any(), gomock.Any()).Return(nil)

	// Execute ChangeIdentifier
	changeResp, changeErr := ucase.ChangeIdentifier(ctx, data.globalUserID, data.tenantID, data.tenantUserID, constants.IdentifierEmail.String(), data.newIdentifier, constants.IdentifierEmail.String())
	assert.NoError(t, changeErr)
	assert.NotNil(t, changeResp)
	assert.Equal(t, data.flowID, changeResp.FlowID)

	// VerifyRegister expectations
	// challenge session returned with ChangeIdentifier info
	mockDeps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), data.flowID).Return(&domain.ChallengeSession{
		GlobalUserID:   data.globalUserID,
		TenantUserID:   data.tenantUserID,
		Identifier:     data.newIdentifier,
		IdentifierType: constants.IdentifierEmail.String(),
		ChallengeType:  constants.ChallengeTypeChangeIdentifier,
		IdentityID:     "identity-email-id",
	}, nil)
	// kratos: get registration flow and submit with code
	mockDeps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), data.tenantID, data.flowID).Return(regFlow, nil)
	newTenantUserID := "00000000-0000-0000-0000-00000000bbbb"
	verifyResult := &kratos.SuccessfulNativeRegistration{
		Session: &kratos.Session{
			Id:                    "session-id",
			Active:                ptr(true),
			ExpiresAt:             ptr(time.Now().Add(30 * time.Minute)),
			IssuedAt:              ptr(time.Now()),
			AuthenticatedAt:       ptr(time.Now()),
			Identity:              &kratos.Identity{Id: newTenantUserID, Traits: map[string]interface{}{"tenant": "tenant-name", string(constants.IdentifierEmail): data.newIdentifier}},
			AuthenticationMethods: []kratos.SessionAuthenticationMethod{{Method: ptr("code")}},
		},
		SessionToken: ptr("token"),
	}
	mockDeps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), data.tenantID, regFlow, gomock.Any()).Return(verifyResult, nil)
	// repo interactions inside bindIAMToUpdateIdentifier
	mockDeps.userIdentifierMappingRepo.EXPECT().GetByTenantIDAndTenantUserID(gomock.Any(), data.tenantID.String(), data.tenantUserID).Return(&domain.UserIdentifierMapping{ID: "map-id", GlobalUserID: data.globalUserID, TenantID: data.tenantID.String(), TenantUserID: data.tenantUserID}, nil)
	mockDeps.userIdentityRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	mockDeps.userIdentifierMappingRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	// delete old identifier in kratos
	mockDeps.kratosService.EXPECT().DeleteIdentifierAdmin(gomock.Any(), data.tenantID, gomock.Any()).Return(nil)
	// get tenant by name
	mockDeps.tenantRepo.EXPECT().GetByName("tenant-name").Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	// delete challenge at the end
	mockDeps.challengeSessionRepo.EXPECT().DeleteChallenge(gomock.Any(), data.flowID).Return(nil)

	// Execute VerifyRegister
	verifyResp, verifyErr := ucase.VerifyRegister(ctx, data.tenantID, data.flowID, "123456")
	assert.NoError(t, verifyErr)
	assert.NotNil(t, verifyResp)
	assert.Equal(t, newTenantUserID, verifyResp.User.ID)
}

// Integration-style negative path: ChangeIdentifier -> VerifyRegister fails, no mutations
func TestChangeIdentifierThenVerifyRegister_Failure_NoMutations(t *testing.T) {
	t.Skip("covered by isolated integration tests")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeps := setupTestDependencies(ctrl)
	ucase := newTestUserUseCase(mockDeps)
	// Use in-memory sqlite so gorm.Transaction works
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	mockDeps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	data := struct {
		globalUserID  string
		tenantID      uuid.UUID
		tenantUserID  string
		flowID        string
		newIdentifier string
	}{
		globalUserID:  "global-user-1",
		tenantID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		tenantUserID:  "00000000-0000-0000-0000-00000000cccc",
		flowID:        "flow-neg",
		newIdentifier: "newemail@example.com",
	}

	// ChangeIdentifier set up
	mockDeps.userIdentityRepo.EXPECT().
		ExistsWithinTenant(gomock.Any(), data.tenantID.String(), constants.IdentifierEmail.String(), data.newIdentifier).
		Return(false, nil)
	mockDeps.userIdentityRepo.EXPECT().
		ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), data.tenantID.String(), data.tenantUserID).
		Return([]*domain.UserIdentity{{ID: "identity-email-id", Type: constants.IdentifierEmail.String()}}, nil)
	mockDeps.tenantRepo.EXPECT().GetByID(data.tenantID).Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	regFlow := &kratos.RegistrationFlow{Id: data.flowID}
	mockDeps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), data.tenantID).Return(regFlow, nil)
	mockDeps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), data.tenantID, regFlow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
	mockDeps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), data.flowID, gomock.Any(), gomock.Any()).Return(nil)

	// Execute ChangeIdentifier
	changeResp, changeErr := ucase.ChangeIdentifier(ctx, data.globalUserID, data.tenantID, data.tenantUserID, constants.IdentifierEmail.String(), data.newIdentifier, constants.IdentifierEmail.String())
	assert.NoError(t, changeErr)
	assert.NotNil(t, changeResp)

	// VerifyRegister negative path setup: return error before any mutation
	mockDeps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), data.flowID).Return(&domain.ChallengeSession{
		GlobalUserID:   data.globalUserID,
		TenantUserID:   data.tenantUserID,
		Identifier:     data.newIdentifier,
		IdentifierType: constants.IdentifierEmail.String(),
		ChallengeType:  constants.ChallengeTypeChangeIdentifier,
		IdentityID:     "identity-email-id",
	}, nil)
	mockDeps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), data.tenantID, data.flowID).Return(regFlow, nil)
	mockDeps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), data.tenantID, regFlow, gomock.Any()).Return(nil, assert.AnError)

	// Ensure no mutations happen
	mockDeps.userIdentityRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
	mockDeps.userIdentifierMappingRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
	mockDeps.userIdentifierMappingRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
	mockDeps.kratosService.EXPECT().DeleteIdentifierAdmin(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	mockDeps.challengeSessionRepo.EXPECT().DeleteChallenge(gomock.Any(), data.flowID).Times(0)

	// Execute VerifyRegister and assert error
	verifyResp, verifyErr := ucase.VerifyRegister(ctx, data.tenantID, data.flowID, "000000")
	assert.Error(t, verifyErr)
	assert.Nil(t, verifyResp)
}

// helpers
func ptr[T any](v T) *T { return &v }
