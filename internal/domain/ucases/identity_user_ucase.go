package ucases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	ratelimiters "github.com/lifenetwork-ai/iam-service/infrastructures/rate_limiter/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	domainservice "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/services"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
	client "github.com/ory/kratos-client-go"
)

var (
	codeInvalidPhone = "MSG_INVALID_PHONE_NUMBER"
	msgInvalidPhone  = "Invalid phone number"
)

type userUseCase struct {
	db                        *gorm.DB
	rateLimiter               ratelimiters.RateLimiter
	tenantRepo                domainrepo.TenantRepository
	globalUserRepo            domainrepo.GlobalUserRepository
	userIdentityRepo          domainrepo.UserIdentityRepository
	userIdentifierMappingRepo domainrepo.UserIdentifierMappingRepository
	challengeSessionRepo      domainrepo.ChallengeSessionRepository
	kratosService             domainservice.KratosService
}

func NewIdentityUserUseCase(
	db *gorm.DB,
	rateLimiter ratelimiters.RateLimiter,
	challengeSessionRepo domainrepo.ChallengeSessionRepository,
	tenantRepo domainrepo.TenantRepository,
	globalUserRepo domainrepo.GlobalUserRepository,
	userIdentityRepo domainrepo.UserIdentityRepository,
	userIdentifierMappingRepo domainrepo.UserIdentifierMappingRepository,
	kratosService domainservice.KratosService,
) interfaces.IdentityUserUseCase {
	return &userUseCase{
		db:                        db,
		rateLimiter:               rateLimiter,
		challengeSessionRepo:      challengeSessionRepo,
		tenantRepo:                tenantRepo,
		globalUserRepo:            globalUserRepo,
		userIdentityRepo:          userIdentityRepo,
		userIdentifierMappingRepo: userIdentifierMappingRepo,
		kratosService:             kratosService,
	}
}

// ChallengeWithPhone challenges the user with a phone number
func (u *userUseCase) ChallengeWithPhone(
	ctx context.Context,
	tenantID uuid.UUID,
	phone string,
) (*types.IdentityUserChallengeResponse, *domainerrors.DomainError) {
	// Get tenant
	_, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_GET_TENANT_FAILED", "Failed to get tenant")
	}

	phone, _, err = utils.NormalizePhoneE164(phone, constants.DefaultRegion)
	if err != nil {
		code := codeInvalidPhone
		msg := msgInvalidPhone
		return nil, domainerrors.NewValidationError(code, msg, nil)
	}

	// Check rate limit for phone challenges
	key := fmt.Sprintf("challenge:phone:tenant:%s:%s", phone, tenantID.String())
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// Check if the identifier exists in the database
	_, err = u.userIdentityRepo.GetByTypeAndValue(ctx, nil, tenantID.String(), constants.IdentifierPhone.String(), phone)
	if err != nil {
		return nil, domainerrors.NewNotFoundError("MSG_IDENTITY_NOT_FOUND", "Phone number").WithDetails([]interface{}{
			map[string]string{
				"field": "phone",
				"error": "Phone number not registered in the system",
			},
		})
	}

	// Initialize verification flow with Kratos
	flow, err := u.kratosService.InitializeLoginFlow(ctx, tenantID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_VERIFICATION_FLOW_FAILED", "Failed to initialize verification flow")
	}

	// Submit login flow to Kratos
	_, err = u.kratosService.SubmitLoginFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), &phone, nil, nil)
	if err != nil {
		return nil, domainerrors.NewUnauthorizedError("MSG_LOGIN_FAILED", "Login failed").WithCause(err)
	}

	// Create challenge session
	err = u.challengeSessionRepo.SaveChallenge(ctx, flow.Id, &domain.ChallengeSession{
		IdentifierType: constants.IdentifierPhone.String(),
		Identifier:     phone,
	}, constants.DefaultChallengeDuration)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_SAVING_SESSION_FAILED", "Saving challenge session failed")
	}

	return &types.IdentityUserChallengeResponse{
		FlowID:      flow.Id,
		Receiver:    phone,
		ChallengeAt: time.Now().Unix(),
	}, nil
}

// ChallengeWithEmail challenges the user with an email
func (u *userUseCase) ChallengeWithEmail(
	ctx context.Context,
	tenantID uuid.UUID,
	email string,
) (*types.IdentityUserChallengeResponse, *domainerrors.DomainError) {
	// Normalize and validate email
	email = strings.ToLower(email)
	if !utils.IsEmail(email) {
		return nil, domainerrors.NewValidationError(
			"MSG_INVALID_EMAIL",
			"Invalid email",
			[]interface{}{
				map[string]string{
					"field": constants.IdentifierEmail.String(),
					"error": "Invalid email",
				},
			},
		)
	}

	// Check rate limit for email challenges
	key := fmt.Sprintf("challenge:email:tenant:%s:%s", email, tenantID.String())
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// Initialize login flow with Kratos
	flow, err := u.kratosService.InitializeLoginFlow(ctx, tenantID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_VERIFICATION_FLOW_FAILED", "Failed to initialize login flow")
	}

	// Submit login flow to Kratos
	_, err = u.kratosService.SubmitLoginFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), &email, nil, nil)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_VERIFICATION_FLOW_FAILED", "Failed to submit login flow")
	}

	// Create challenge session
	err = u.challengeSessionRepo.SaveChallenge(ctx, flow.Id, &domain.ChallengeSession{
		IdentifierType: constants.IdentifierEmail.String(),
		Identifier:     email,
	}, constants.DefaultChallengeDuration)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_SAVING_SESSION_FAILED", "Saving challenge session failed")
	}

	// Return challenge session
	return &types.IdentityUserChallengeResponse{
		FlowID:      flow.Id,
		Receiver:    email,
		ChallengeAt: time.Now().Unix(),
	}, nil
}

// VerifyRegister verifies the registration flow
func (u *userUseCase) VerifyRegister(
	ctx context.Context,
	tenantID uuid.UUID,
	flowID string,
	code string,
) (*types.IdentityUserAuthResponse, *domainerrors.DomainError) {
	// Check rate limit for verification attempts
	key := "verify:register:" + flowID
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	sessionValue, err := u.challengeSessionRepo.GetChallenge(ctx, flowID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_CHALLENGE_SESSION_NOT_FOUND", "Challenge session not found")
	}

	flow, err := u.kratosService.GetRegistrationFlow(ctx, tenantID, flowID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get registration flow: %v", err)
		return nil, domainerrors.WrapInternal(err, "MSG_GET_FLOW_FAILED", "Failed to get registration flow")
	}

	// Submit registration flow with code
	registrationResult, err := u.kratosService.SubmitRegistrationFlowWithCode(ctx, tenantID, flow, code)
	if err != nil {
		logger.GetLogger().Errorf("Failed to submit registration flow with code: %v", err)
		return nil, domainerrors.NewValidationError("MSG_REGISTRATION_FAILED", "Registration failed", []interface{}{err.Error()})
	}

	// Extract traits
	traits, ok := registrationResult.Session.Identity.Traits.(map[string]interface{})
	if !ok {
		logger.GetLogger().Errorf("Failed to parse identity traits: %v", traits)
		return nil, domainerrors.NewInternalError("MSG_INVALID_TRAITS", "Failed to parse identity traits")
	}

	tenantName := extractStringFromTraits(traits, "tenant", "")
	newIdentityID := registrationResult.Session.Identity.Id

	// Get identifier and type from session value
	identifier := sessionValue.Identifier
	identifierType := sessionValue.IdentifierType

	// Get tenant by name
	tenant, err := u.tenantRepo.GetByName(tenantName)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_GET_TENANT_FAILED", "Failed to get tenant")
	}
	if tenant == nil {
		return nil, domainerrors.NewNotFoundError("MSG_TENANT_NOT_FOUND", "Tenant").WithDetails([]interface{}{
			map[string]string{"field": "tenant", "error": "Tenant not found"},
		})
	}

	switch sessionValue.ChallengeType {
	case constants.ChallengeTypeAddIdentifier:
		// Handle add identifier challenge
		inserted, err := u.userIdentityRepo.InsertOnceByTenantUserAndType(
			ctx, nil, tenantID.String(), sessionValue.GlobalUserID, identifierType, identifier,
		)
		if err != nil || !inserted {
			return nil, domainerrors.NewConflictError("MSG_IDENTIFIER_TYPE_EXISTS", "Identifier of this type already added", nil)
		}

	case constants.ChallengeTypeChangeIdentifier:
		// Handle change identifier challenge
		if err := u.bindIAMToUpdateIdentifier(
			ctx,
			tenant,
			sessionValue.GlobalUserID,
			sessionValue.IdentityID, // the identity id of the old identifier to be deleted
			newIdentityID,
			identifier,
			identifierType,
		); err != nil {
			return nil, domainerrors.WrapInternal(err, "MSG_UPDATE_IDENTIFIER_FAILED", "Failed to update identifier")
		}

	default:
		// Bind IAM to registration
		if err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return u.bindIAMToRegistration(ctx, tx, tenant, newIdentityID, identifier, identifierType)
		}); err != nil {
			return nil, domainerrors.WrapInternal(err, "MSG_IAM_REGISTRATION_FAILED", "Failed to bind IAM to registration")
		}
	}

	// Delete challenge session
	_ = u.challengeSessionRepo.DeleteChallenge(ctx, flowID)

	// Return authentication response
	return &types.IdentityUserAuthResponse{
		SessionID:       registrationResult.Session.Id,
		SessionToken:    *registrationResult.SessionToken,
		Active:          *registrationResult.Session.Active,
		ExpiresAt:       registrationResult.Session.ExpiresAt,
		IssuedAt:        registrationResult.Session.IssuedAt,
		AuthenticatedAt: registrationResult.Session.AuthenticatedAt,
		User: types.IdentityUserResponse{
			ID:       newIdentityID,
			UserName: extractStringFromTraits(traits, constants.IdentifierUsername.String(), ""),
			Email:    extractStringFromTraits(traits, constants.IdentifierEmail.String(), ""),
			Phone:    extractStringFromTraits(traits, constants.IdentifierPhone.String(), ""),
		},
		AuthenticationMethods: utils.Map(registrationResult.Session.AuthenticationMethods, func(method client.SessionAuthenticationMethod) string {
			return *method.Method
		}),
	}, nil
}

// bindIAMToUpdateIdentifier handles updating to a different identifier
// keeping the same GlobalUserID but mapping to the new TenantUserID
// All write operations are wrapped in a transaction to ensure atomicity
func (u *userUseCase) bindIAMToUpdateIdentifier(
	ctx context.Context,
	tenant *domain.Tenant,
	globalUserID string,
	oldIdentityID string, // the tenant user id of the old identifier to be deleted
	newIdentityID string,
	newIdentifier string,
	newIdentifierType string,
) error {
	// Pre-fetch the current mapping before starting the transaction
	identifierMapping, err := u.userIdentifierMappingRepo.GetByTenantIDAndTenantUserID(
		ctx,
		tenant.ID.String(),
		oldIdentityID,
	)
	if err != nil {
		return fmt.Errorf("get existing mapping: %w", err)
	}

	// Get the old identifier
	oldIdentity, err := u.userIdentityRepo.GetByID(ctx, nil, oldIdentityID)
	if err != nil {
		return fmt.Errorf("Old identity with id: %s not found: %w", oldIdentityID, err)
	}

	// Begin transaction
	txErr := u.db.Transaction(func(tx *gorm.DB) error {
		// Create identity for the new identifier with the existing GlobalUserID
		if err := u.userIdentityRepo.Update(tx, &domain.UserIdentity{
			ID:           oldIdentity.ID,
			GlobalUserID: oldIdentity.GlobalUserID,
			TenantID:     tenant.ID.String(),
			Type:         newIdentifierType,
			Value:        newIdentifier,
		}); err != nil {
			return fmt.Errorf("create new identity: %w", err)
		}

		if err := u.userIdentifierMappingRepo.Update(tx, &domain.UserIdentifierMapping{
			ID:           identifierMapping.ID,
			TenantUserID: newIdentityID,
		}); err != nil {
			return fmt.Errorf("update mapping: %w", err)
		}

		// If we reach here, all operations succeeded and the transaction will be committed
		return nil
	})

	if txErr != nil {
		if cleanUpErr := u.rollbackKratosUpdateIdentifier(ctx, tenant, newIdentityID); cleanUpErr != nil {
			logger.GetLogger().Errorf("Failed to clean up: %v", cleanUpErr)
		}
		return fmt.Errorf("failed to bind IAM to update identifier: %v", txErr)
	}

	// Delete the old identifier from Kratos
	if err := u.kratosService.DeleteIdentifierAdmin(ctx, tenant.ID, uuid.MustParse(oldIdentityID)); err != nil {
		return fmt.Errorf("failed to delete old identifier: %w", err)
	}

	return nil
}

// rollbackKratosUpdateIdentifier rolls back the changes to Kratos when the update identifier flow fails
func (u *userUseCase) rollbackKratosUpdateIdentifier(
	ctx context.Context,
	tenant *domain.Tenant,
	newTenantUserID string,
) error {
	if err := u.kratosService.DeleteIdentifierAdmin(ctx, tenant.ID, uuid.MustParse(newTenantUserID)); err != nil {
		return fmt.Errorf("failed to delete identifier with tenantID: %s and newTenantUserID: %s", tenant.ID.String(), newTenantUserID)
	}

	return nil
}

// bindIAMToRegistration binds the IAM records to the registration flow
func (u *userUseCase) bindIAMToRegistration(
	ctx context.Context,
	tx *gorm.DB,
	tenant *domain.Tenant,
	newTenantUserID string,
	identifier string,
	identifierType string,
) error {
	var globalUserID string

	// Lookup existing identity
	if identity, err := u.userIdentityRepo.GetByTypeAndValue(ctx, tx, tenant.ID.String(), identifierType, identifier); err == nil {
		globalUserID = identity.GlobalUserID
	}

	var globalUser *domain.GlobalUser
	if globalUserID != "" {
		globalUser = &domain.GlobalUser{ID: globalUserID}

		// Check if already mapped
		exists, err := u.userIdentifierMappingRepo.ExistsByTenantAndTenantUserID(ctx, tx, tenant.ID.String(), newTenantUserID)
		if err != nil {
			return fmt.Errorf("check mapping exists: %w", err)
		}
		if exists {
			return nil
		}
	} else {
		globalUser = &domain.GlobalUser{}
		if err := u.globalUserRepo.Create(tx, globalUser); err != nil {
			return fmt.Errorf("create global user: %w", err)
		}
		globalUserID = globalUser.ID
	}

	// Create identities
	_, err := u.userIdentityRepo.InsertOnceByTenantUserAndType(
		ctx, tx,
		tenant.ID.String(),
		globalUserID,
		identifierType,
		identifier,
	)
	if err != nil {
		return fmt.Errorf("create identity: %w", err)
	}

	// Create mapping
	if err := u.userIdentifierMappingRepo.Create(tx, &domain.UserIdentifierMapping{
		GlobalUserID: globalUser.ID,
		TenantID:     tenant.ID.String(),
		TenantUserID: newTenantUserID,
	}); err != nil {
		return fmt.Errorf("create mapping: %w", err)
	}

	return nil
}

// VerifyLogin verifies the login flow
func (u *userUseCase) VerifyLogin(
	ctx context.Context,
	tenantID uuid.UUID,
	flowID string,
	code string,
) (*types.IdentityUserAuthResponse, *domainerrors.DomainError) {
	// Check rate limit for verification attempts
	key := "verify:login:" + flowID
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// Get the login flow
	flow, err := u.kratosService.GetLoginFlow(ctx, tenantID, flowID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get login flow: %v", err)
		return nil, domainerrors.WrapInternal(err, "MSG_GET_FLOW_FAILED", "Failed to get login flow")
	}

	// Get the challenge session to retrieve the phone number
	sessionValue, err := u.challengeSessionRepo.GetChallenge(ctx, flowID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_CHALLENGE_SESSION_NOT_FOUND", "Challenge session not found")
	}

	if sessionValue == nil {
		return nil, domainerrors.NewNotFoundError("MSG_CHALLENGE_SESSION_NOT_FOUND", "Challenge session").WithDetails([]interface{}{
			map[string]string{"field": "session", "error": "Session not found"},
		})
	}

	// Submit login flow with code
	identifier := sessionValue.Identifier
	loginResult, err := u.kratosService.SubmitLoginFlow(
		ctx, tenantID, flow, constants.MethodTypeCode.String(), &identifier, nil, &code,
	)
	if err != nil {
		return nil, domainerrors.NewValidationError("MSG_LOGIN_FAILED", "Login failed", []interface{}{err.Error()})
	}

	// Delete challenge session
	_ = u.challengeSessionRepo.DeleteChallenge(ctx, flowID)

	// Return authentication response
	return &types.IdentityUserAuthResponse{
		SessionID:       loginResult.Session.Id,
		SessionToken:    *loginResult.SessionToken,
		Active:          *loginResult.Session.Active,
		ExpiresAt:       loginResult.Session.ExpiresAt,
		IssuedAt:        loginResult.Session.IssuedAt,
		AuthenticatedAt: loginResult.Session.AuthenticatedAt,
		User: types.IdentityUserResponse{
			ID:       loginResult.Session.Identity.Id,
			UserName: extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), constants.IdentifierUsername.String(), ""),
			Email:    extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), constants.IdentifierEmail.String(), ""),
			Phone:    extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), constants.IdentifierPhone.String(), ""),
		},
		AuthenticationMethods: utils.Map(loginResult.Session.AuthenticationMethods, func(method client.SessionAuthenticationMethod) string {
			return *method.Method
		}),
	}, nil
}

// Register registers a new user
func (u *userUseCase) Register(
	ctx context.Context,
	tenantID uuid.UUID,
	lang string,
	email string,
	phone string,
) (*types.IdentityUserAuthResponse, *domainerrors.DomainError) {
	// Normalize and validate phone number if provided
	if phone != "" {
		normalizedPhone, _, err := utils.NormalizePhoneE164(phone, constants.DefaultRegion)
		if err != nil {
			code := codeInvalidPhone
			msg := msgInvalidPhone
			return nil, domainerrors.NewValidationError(code, msg, nil)
		}
		phone = normalizedPhone
	}

	// Normalize and validate email address if provided
	email = strings.ToLower(email)
	if email != "" && !utils.IsEmail(email) {
		return nil, domainerrors.NewValidationError("MSG_INVALID_EMAIL_ADDRESS", "Invalid email address format", []any{"Email address must be a valid format"})
	}

	// Check rate limit for registration attempts
	var key string
	if phone != "" {
		key = fmt.Sprintf("register:phone:tenant:%s:%s", phone, tenantID.String())
	} else {
		key = fmt.Sprintf("register:email:tenant:%s:%s", email, tenantID.String())
	}
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// Check if identifier (email/phone) already exists in IAM
	var identifierType string
	var identifierValue string
	if email != "" {
		identifierType = constants.IdentifierEmail.String()
		identifierValue = email
	} else {
		identifierType = constants.IdentifierPhone.String()
		identifierValue = phone
	}

	exists, err := u.userIdentityRepo.ExistsWithinTenant(ctx, tenantID.String(), identifierType, identifierValue)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_IAM_LOOKUP_FAILED", "Failed to check existing identifier")
	}
	if exists {
		return nil, domainerrors.NewConflictError("MSG_IDENTIFIER_ALREADY_EXISTS", "Identifier has already been registered", nil)
	}

	// Initialize registration flow with Kratos
	flow, err := u.kratosService.InitializeRegistrationFlow(ctx, tenantID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to initialize registration flow: %v", err)
		return nil, domainerrors.WrapInternal(err, "MSG_INITIALIZE_REGISTRATION_FAILED", "Failed to initialize registration flow")
	}

	tenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to initialize registration flow: %v", err)
		return nil, domainerrors.WrapInternal(err, "MSG_GET_TENANT_FAILED", "Failed to get tenant")
	}

	// Prepare traits
	traits := map[string]interface{}{
		"tenant": tenant.Name,
		"lang":   lang,
	}
	if email != "" {
		traits[constants.IdentifierEmail.String()] = identifierValue
	}
	if phone != "" {
		traits[constants.IdentifierPhone.String()] = identifierValue
	}

	// Submit registration flow
	_, err = u.kratosService.SubmitRegistrationFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), traits)
	if err != nil {
		logger.GetLogger().Errorf("Failed to submit registration flow: %v", err)
		return nil, domainerrors.NewValidationError("MSG_REGISTRATION_FAILED", "Registration failed", []interface{}{err.Error()})
	}

	// Save challenge session
	session := &domain.ChallengeSession{
		ChallengeType:  constants.ChallengeTypeRegister,
		Identifier:     identifierValue,
		IdentifierType: identifierType,
	}

	if err := u.challengeSessionRepo.SaveChallenge(ctx, flow.Id, session, constants.DefaultChallengeDuration); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_SAVE_CHALLENGE_FAILED", "Failed to save challenge session")
	}

	// Return success with verification flow info
	return &types.IdentityUserAuthResponse{
		VerificationNeeded: true,
		VerificationFlow: &types.IdentityUserChallengeResponse{
			FlowID:      flow.Id,
			Receiver:    identifierValue,
			ChallengeAt: time.Now().Unix(),
		},
	}, nil
}

// Login logs in a user with username and password
func (u *userUseCase) Login(
	ctx context.Context,
	tenantID uuid.UUID,
	username string,
	password string,
) (*types.IdentityUserAuthResponse, *domainerrors.DomainError) {
	// Initialize login flow
	flow, err := u.kratosService.InitializeLoginFlow(ctx, tenantID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to initialize login flow: %v", err)
		return nil, domainerrors.WrapInternal(err, "MSG_INITIALIZE_LOGIN_FAILED", "Failed to initialize login flow")
	}

	// Submit login flow to Kratos
	loginResult, err := u.kratosService.SubmitLoginFlow(ctx, tenantID, flow, constants.MethodTypePassword.String(), &username, &password, nil)
	if err != nil {
		logger.GetLogger().Errorf("Failed to submit login flow: %v", err)
		return nil, domainerrors.NewUnauthorizedError("MSG_LOGIN_FAILED", "Login failed").WithCause(err)
	}

	// Return authentication response
	return &types.IdentityUserAuthResponse{
		SessionID:       loginResult.Session.Id,
		SessionToken:    *loginResult.SessionToken,
		Active:          *loginResult.Session.Active,
		ExpiresAt:       loginResult.Session.ExpiresAt,
		IssuedAt:        loginResult.Session.IssuedAt,
		AuthenticatedAt: loginResult.Session.AuthenticatedAt,
		User: types.IdentityUserResponse{
			ID:       loginResult.Session.Identity.Id,
			UserName: extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), constants.IdentifierUsername.String(), ""),
			Email:    extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), constants.IdentifierEmail.String(), ""),
			Phone:    extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), constants.IdentifierPhone.String(), ""),
		},
		AuthenticationMethods: utils.Map(loginResult.Session.AuthenticationMethods, func(method client.SessionAuthenticationMethod) string {
			return *method.Method
		}),
	}, nil
}

// Logout logs out a user
func (u *userUseCase) Logout(
	ctx context.Context,
	tenantID uuid.UUID,
) *domainerrors.DomainError {
	// Get session token from context
	sessionToken, err := extractSessionToken(ctx)
	if err != nil {
		return err
	}

	// Check if session is active
	session, kratosErr := u.kratosService.GetSession(ctx, tenantID, sessionToken)

	if kratosErr != nil {
		return domainerrors.NewUnauthorizedError("MSG_INVALID_SESSION", "Invalid session").WithCause(kratosErr)
	}

	if !*session.Active {
		return domainerrors.NewUnauthorizedError("MSG_INVALID_SESSION", "Invalid session")
	}

	// Revoke session
	if err := u.kratosService.Logout(ctx, tenantID, sessionToken); err != nil {
		logger.GetLogger().Errorf("Failed to logout: %v", err)
		return domainerrors.WrapInternal(err, "MSG_LOGOUT_FAILED", "Failed to logout")
	}

	return nil
}

// RefreshToken refreshes a user's session token
func (u *userUseCase) RefreshToken(
	ctx context.Context,
	tenantID uuid.UUID,
	accessToken string,
	refreshToken string,
) (*types.IdentityUserAuthResponse, *domainerrors.DomainError) {
	// Get session
	session, err := u.kratosService.GetSession(ctx, tenantID, accessToken)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get session: %v", err)
		return nil, domainerrors.NewUnauthorizedError("MSG_INVALID_SESSION", "Invalid session").WithCause(err)
	}

	// Return authentication response
	return &types.IdentityUserAuthResponse{
		SessionID:       session.Id,
		SessionToken:    accessToken,
		Active:          *session.Active,
		ExpiresAt:       session.ExpiresAt,
		IssuedAt:        session.IssuedAt,
		AuthenticatedAt: session.AuthenticatedAt,
		User: types.IdentityUserResponse{
			ID:       session.Identity.Id,
			UserName: extractStringFromTraits(session.Identity.Traits.(map[string]interface{}), constants.IdentifierUsername.String(), ""),
			Email:    extractStringFromTraits(session.Identity.Traits.(map[string]interface{}), constants.IdentifierEmail.String(), ""),
			Phone:    extractStringFromTraits(session.Identity.Traits.(map[string]interface{}), constants.IdentifierPhone.String(), ""),
		},
		AuthenticationMethods: utils.Map(session.AuthenticationMethods, func(method client.SessionAuthenticationMethod) string {
			return *method.Method
		}),
	}, nil
}

// Profile gets a user's profile
func (u *userUseCase) Profile(
	ctx context.Context,
	tenantID uuid.UUID,
) (*types.IdentityUserResponse, *domainerrors.DomainError) {
	// Get session token from context
	sessionToken, sessionTokenErr := extractSessionToken(ctx)
	if sessionTokenErr != nil {
		return nil, sessionTokenErr
	}

	// Get session
	session, err := u.kratosService.WhoAmI(ctx, tenantID, sessionToken)
	if err != nil {
		logger.GetLogger().Errorf("Failed to extract user traits: %v", err)
		return nil, domainerrors.NewUnauthorizedError("MSG_INVALID_SESSION", "Invalid session").WithCause(err)
	}

	// Extract user traits
	user, err := extractUserFromTraits(session.Identity.Traits, "", "")
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_EXTRACT_USER_FAILED", "Failed to extract user traits")
	}

	// If user has no email or phone, return error
	identifier := user.Email
	if identifier == "" {
		identifier = user.Phone
	}

	identifierType, err := utils.GetIdentifierType(identifier)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_GET_IDENTIFIER_TYPE_FAILED", "Failed to get identifier type")
	}

	// Lookup global user ID
	globalUserID, err := u.userIdentityRepo.FindGlobalUserIDByIdentity(ctx, tenantID.String(), identifierType, identifier)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_LOOKUP_FAILED", "Failed to lookup user ID")
	}

	// Set user id
	user.ID = session.Identity.Id
	user.GlobalUserID = globalUserID

	return &user, nil
}

// AddNewIdentifier adds a new identifier (email or phone) to a user
func (u *userUseCase) AddNewIdentifier(
	ctx context.Context,
	tenantID uuid.UUID,
	globalUserID string,
	identifier string,
	identifierType string,
) (*types.IdentityUserChallengeResponse, *domainerrors.DomainError) {
	// 1. Validate input
	if identifierType == constants.IdentifierEmail.String() {
		// Normalize and validate email
		identifier = strings.ToLower(identifier)
		if !utils.IsEmail(identifier) {
			return nil, domainerrors.NewValidationError("MSG_INVALID_EMAIL", "Invalid email", nil)
		}
	}

	if identifierType == constants.IdentifierPhone.String() {
		normalizedPhone, _, err := utils.NormalizePhoneE164(identifier, constants.DefaultRegion)
		if err != nil {
			code := codeInvalidPhone
			msg := msgInvalidPhone
			return nil, domainerrors.NewValidationError(code, msg, nil)
		}
		identifier = normalizedPhone
	}

	// 2. Check if identifier already exists globally
	exists, err := u.userIdentityRepo.ExistsWithinTenant(ctx, tenantID.String(), identifierType, identifier)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_IAM_LOOKUP_FAILED", "Failed to check existing identifier")
	}
	if exists {
		return nil, domainerrors.NewConflictError("MSG_IDENTIFIER_ALREADY_EXISTS", "Identifier has already been registered", nil)
	}

	// 3. Check if user already has this identifier type
	hasType, err := u.userIdentityRepo.ExistsByTenantGlobalUserIDAndType(ctx, tenantID.String(), globalUserID, identifierType)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_CHECK_TYPE_EXIST_FAILED", "Failed to check user identity type")
	}
	if hasType {
		return nil, domainerrors.NewConflictError("MSG_IDENTIFIER_TYPE_EXISTS", fmt.Sprintf("User already has an identifier of type %s", identifierType), nil)
	}

	// 4. Rate limit
	key := fmt.Sprintf("challenge:add:tenant:%s:%s", identifier, tenantID.String())
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// 5. Init Kratos Registration Flow
	flow, err := u.kratosService.InitializeRegistrationFlow(ctx, tenantID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_INIT_REG_FLOW_FAILED", "Failed to initialize registration flow")
	}

	tenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to initialize registration flow: %v", err)
		return nil, domainerrors.WrapInternal(err, "MSG_GET_TENANT_FAILED", "Failed to get tenant")
	}

	// Prepare traits
	traits := map[string]interface{}{
		"tenant": tenant.Name,
	}
	traits[identifierType] = identifier

	// 6. Submit minimal traits to trigger OTP (email or phone)
	if _, err := u.kratosService.SubmitRegistrationFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), traits); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_REGISTRATION_FAILED", "Registration failed").WithCause(err)
	}

	// 7. Save challenge session
	session := &domain.ChallengeSession{
		GlobalUserID:   globalUserID,
		ChallengeType:  constants.ChallengeTypeAddIdentifier,
		Identifier:     identifier,
		IdentifierType: identifierType,
	}

	if err := u.challengeSessionRepo.SaveChallenge(ctx, flow.Id, session, constants.DefaultChallengeDuration); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_SAVE_CHALLENGE_FAILED", "Failed to save challenge session")
	}

	// 8. Return response
	return &types.IdentityUserChallengeResponse{
		FlowID:      flow.Id,
		Receiver:    identifier,
		ChallengeAt: time.Now().Unix(),
	}, nil
}

func (u *userUseCase) CheckIdentifier(
	ctx context.Context,
	tenantID uuid.UUID,
	identifier string,
) (bool, string, *domainerrors.DomainError) {
	identifier = strings.ToLower(identifier)
	identifier = strings.TrimSpace(identifier)

	// 1. Detect type
	idType, err := utils.GetIdentifierType(identifier) // "email" | "phone_number"
	if err != nil {
		return false, "", domainerrors.NewValidationError("MSG_INVALID_IDENTIFIER_TYPE", "Invalid identifier type", map[string]interface{}{"identifier": identifier}).WithCause(err)
	}

	// 2. Validate
	switch idType {
	case constants.IdentifierEmail.String():
		if !utils.IsEmail(identifier) {
			return false, idType, domainerrors.NewValidationError("MSG_INVALID_EMAIL", "Invalid email", nil)
		}
	case constants.IdentifierPhone.String():
		normalizedPhone, _, err := utils.NormalizePhoneE164(identifier, constants.DefaultRegion)
		if err != nil {
			code := codeInvalidPhone
			msg := msgInvalidPhone
			return false, idType, domainerrors.NewValidationError(code, msg, nil)
		}
		identifier = normalizedPhone
	default:
		return false, "", domainerrors.NewValidationError("MSG_INVALID_IDENTIFIER_TYPE", "Invalid identifier type", nil)
	}

	// 3. Repo check (tenant-scoped)
	ok, repoErr := u.userIdentityRepo.ExistsWithinTenant(ctx, tenantID.String(), idType, identifier)
	if repoErr != nil {
		return false, idType, domainerrors.WrapInternal(repoErr, "MSG_LOOKUP_FAILED", "Failed to check identifier")
	}

	return ok, idType, nil
}

// DeleteIdentifier deletes a user's identifier (email or phone)
// Prevents deletion when user only has one identifier
func (u *userUseCase) DeleteIdentifier(
	ctx context.Context,
	globalUserID string,
	tenantID uuid.UUID,
	tenantUserID string,
	identifierType string,
) *domainerrors.DomainError {
	// 1. Validate identifier type
	if identifierType != constants.IdentifierEmail.String() && identifierType != constants.IdentifierPhone.String() {
		return domainerrors.NewValidationError("MSG_INVALID_IDENTIFIER_TYPE", "Invalid identifier type", nil)
	}

	// 2. Get all user identities
	identities, err := u.userIdentityRepo.GetByGlobalUserID(ctx, nil, tenantID.String(), globalUserID)
	if err != nil {
		return domainerrors.WrapInternal(err, "MSG_GET_IDENTIFIERS_FAILED", "Failed to get user identifiers")
	}

	// 3. Find the specific identifier to delete
	var identifierToDelete *domain.UserIdentity
	if len(identities) == 0 {
		return domainerrors.NewConflictError("MSG_IDENTIFIER_TYPE_NOT_EXISTS", fmt.Sprintf("User does not have an identifier of type %s", identifierType), nil)
	}
	for _, id := range identities {
		if id.Type == identifierType {
			identifierToDelete = &id
			break
		}
	}

	if identifierToDelete == nil {
		return domainerrors.NewConflictError("MSG_IDENTIFIER_TYPE_NOT_EXISTS", fmt.Sprintf("User does not have an identifier of type %s", identifierType), nil)
	}

	// 4. Check if this is the only identifier
	// Filter to only count email and phone identifiers
	identifierCount := 0
	for _, id := range identities {
		if id.Type == constants.IdentifierEmail.String() || id.Type == constants.IdentifierPhone.String() {
			identifierCount++
		}
	}

	if identifierCount <= 1 {
		return domainerrors.NewConflictError("MSG_CANNOT_DELETE_ONLY_IDENTIFIER", "Cannot delete the only identifier", nil)
	}

	// 5. Delete the identifier from the database
	if err := u.userIdentityRepo.Delete(nil, identifierToDelete.ID); err != nil {
		return domainerrors.WrapInternal(err, "MSG_DELETE_IDENTIFIER_FAILED", "Failed to delete identifier")
	}

	// 6. Delete the identifier from Kratos
	if err := u.kratosService.DeleteIdentifierAdmin(ctx, tenantID, uuid.MustParse(tenantUserID)); err != nil {
		// Log the error but don't fail the operation since we've already deleted from our database
		logger.GetLogger().Errorf("Failed to delete identifier from Kratos: %v", err)
	}

	return nil
}

// ChangeIdentifier changes a user's identifier from one type to another.
// Rules:
// - If user has exactly one identifier type (email OR phone): allow switching to the other type.
// - If user has more than one identifier type: will replace the same type (email→email or phone→phone).
func (u *userUseCase) ChangeIdentifier(
	ctx context.Context,
	globalUserID string,
	tenantID uuid.UUID,
	tenantUserID string,
	newIdentifier string,
) (*types.IdentityUserChallengeResponse, *domainerrors.DomainError) {
	// 1. Validate identifier type
	newIdentifierType, err := utils.GetIdentifierType(newIdentifier)
	if err != nil {
		return nil, domainerrors.NewValidationError("MSG_INVALID_IDENTIFIER_TYPE", "Invalid identifier type", nil)
	}
	if newIdentifierType != constants.IdentifierEmail.String() && newIdentifierType != constants.IdentifierPhone.String() {
		return nil, domainerrors.NewValidationError("MSG_INVALID_IDENTIFIER_TYPE", "Invalid identifier type", nil).WithCause(err)
	}

	// 2. Validate new identifier value
	if newIdentifier == "" {
		return nil, domainerrors.NewValidationError("MSG_INVALID_REQUEST", "Identifier is required", nil).WithCause(err)
	}
	if newIdentifierType == constants.IdentifierEmail.String() && !utils.IsEmail(newIdentifier) {
		return nil, domainerrors.NewValidationError("MSG_INVALID_EMAIL", "Invalid email", nil).WithCause(err)
	}
	if newIdentifierType == constants.IdentifierPhone.String() {
		normalizedPhone, _, err := utils.NormalizePhoneE164(newIdentifier, constants.DefaultRegion)
		if err != nil {
			return nil, domainerrors.NewValidationError(codeInvalidPhone, msgInvalidPhone, nil).WithCause(err)
		}
		newIdentifier = normalizedPhone
	}

	// 3. Check if new identifier already exists globally
	exists, err := u.userIdentityRepo.ExistsWithinTenant(ctx, tenantID.String(), newIdentifierType, newIdentifier)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_IAM_LOOKUP_FAILED", "Failed to check existing identifier")
	}
	if exists {
		return nil, domainerrors.NewConflictError("MSG_IDENTIFIER_ALREADY_EXISTS", "Identifier has already been registered", nil)
	}

	// 4. Check user's current identifiers
	identities, err := u.userIdentityRepo.GetByGlobalUserID(ctx, nil, tenantID.String(), globalUserID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_CHECK_TYPE_EXIST_FAILED", "Failed to check user identities")
	}
	if len(identities) == 0 {
		return nil, domainerrors.NewConflictError("MSG_NO_IDENTIFIER_EXISTS", "User has no identifier", nil)
	}

	// Find the identifier to be changed
	// If user has only one identifier, use it
	// If user has multiple identifiers, use the one with the same type
	var identity *domain.UserIdentity
	if len(identities) == 1 {
		identity = &identities[0]
	} else {
		for _, id := range identities {
			if id.Type == newIdentifierType {
				identity = &id
				break
			}
		}
	}
	// Should not happen
	if identity == nil {
		return nil, domainerrors.NewInternalError("MSG_INTERNAL_ERROR", "Cannot find identifier to be changed")
	}

	// 5. Rate limit
	key := fmt.Sprintf("challenge:change:%s:%s", newIdentifierType, newIdentifier)
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// 6. Initialize a registration flow
	flow, err := u.kratosService.InitializeRegistrationFlow(ctx, tenantID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_INIT_REG_FLOW_FAILED", "Failed to initialize registration flow")
	}

	tenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to initialize registration flow: %v", err)
		return nil, domainerrors.WrapInternal(err, "MSG_GET_TENANT_FAILED", "Failed to get tenant")
	}

	// 7. Submit minimal traits to trigger OTP
	traits := map[string]interface{}{
		"tenant": tenant.Name,
	}
	traits[newIdentifierType] = newIdentifier

	if _, err := u.kratosService.SubmitRegistrationFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), traits); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_REGISTRATION_FAILED", "Registration failed").WithCause(err)
	}

	// 8. Save challenge session
	session := &domain.ChallengeSession{
		GlobalUserID:   globalUserID,
		TenantUserID:   tenantUserID,
		IdentifierType: newIdentifierType,
		Identifier:     newIdentifier,
		ChallengeType:  constants.ChallengeTypeChangeIdentifier,
		IdentityID:     identity.ID, // the identity id of the old identifier to be deleted
	}
	if err := u.challengeSessionRepo.SaveChallenge(ctx, flow.Id, session, constants.DefaultChallengeDuration); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_SAVE_CHALLENGE_FAILED", "Failed to save challenge session")
	}

	// 9. Return response
	return &types.IdentityUserChallengeResponse{
		FlowID:      flow.Id,
		Receiver:    newIdentifier,
		ChallengeAt: time.Now().Unix(),
	}, nil
}

func (u *userUseCase) ChallengeVerification(
	ctx context.Context,
	tenantID uuid.UUID,
	identifier string,
) (*types.IdentityUserChallengeResponse, *domainerrors.DomainError) {
	// 1. Detect & normalize
	identifier = strings.TrimSpace(strings.ToLower(identifier))
	idType, err := utils.GetIdentifierType(identifier) // "email" | "phone_number"
	if err != nil {
		return nil, domainerrors.NewValidationError("MSG_INVALID_IDENTIFIER_TYPE", "Invalid identifier type", nil)
	}
	switch idType {
	case constants.IdentifierEmail.String():
		if !utils.IsEmail(identifier) {
			return nil, domainerrors.NewValidationError("MSG_INVALID_EMAIL", "Invalid email", nil)
		}
	case constants.IdentifierPhone.String():
		identifier, _, err = utils.NormalizePhoneE164(identifier, constants.DefaultRegion)
		if err != nil {
			return nil, domainerrors.NewValidationError(codeInvalidPhone, msgInvalidPhone, nil)
		}
	default:
		return nil, domainerrors.NewValidationError("MSG_INVALID_IDENTIFIER_TYPE", "Invalid identifier type", nil)
	}

	// Make sure identifier exists in the system
	ok, repoErr := u.userIdentityRepo.ExistsWithinTenant(ctx, tenantID.String(), idType, identifier)
	if repoErr != nil {
		return nil, domainerrors.WrapInternal(repoErr, "MSG_LOOKUP_FAILED", "Failed to check identifier")
	}
	if !ok {
		// Only allow verifying existing identifiers
		return nil, domainerrors.NewNotFoundError("MSG_IDENTITY_NOT_FOUND", "Identifier not found in tenant").WithDetails([]interface{}{
			map[string]string{"field": idType, "error": "Identifier not registered in the system"},
		})
	}

	// 2. Rate limit
	key := fmt.Sprintf("challenge:verification:%s:%s:%s", idType, identifier, tenantID.String())
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// 3. Init verification flow
	flowID, err := u.kratosService.InitializeVerificationFlow(ctx, tenantID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_VERIFICATION_FLOW_FAILED", "Failed to initialize verification flow")
	}

	// 4. Submit verification without code -> trigger send OTP
	var codePtr *string // nil
	idPtr := &identifier
	_, err = u.kratosService.SubmitVerificationFlow(
		ctx, tenantID, flowID, idPtr, constants.IdentifierType(idType), codePtr,
	)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_SEND_VERIFICATION_FAILED", "Failed to send verification code")
	}

	// 5. Save challenge session
	session := &domain.ChallengeSession{
		ChallengeType:  constants.ChallengeTypeVerifyIdentifier,
		Identifier:     identifier,
		IdentifierType: idType,
	}
	if err := u.challengeSessionRepo.SaveChallenge(ctx, flowID, session, constants.DefaultChallengeDuration); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_SAVING_SESSION_FAILED", "Saving challenge session failed")
	}

	// 6. Response
	return &types.IdentityUserChallengeResponse{
		FlowID:      flowID,
		Receiver:    identifier,
		ChallengeAt: time.Now().Unix(),
	}, nil
}

func (u *userUseCase) VerifyIdentifier(
	ctx context.Context,
	tenantID uuid.UUID,
	flowID string,
	code string,
) (*types.IdentityVerificationResponse, *domainerrors.DomainError) {
	// 1. Rate limit
	key := "verify:identifier:" + flowID
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// 2. Load session
	sessionValue, err := u.challengeSessionRepo.GetChallenge(ctx, flowID)
	if err != nil || sessionValue == nil {
		return nil, domainerrors.NewNotFoundError("MSG_CHALLENGE_SESSION_NOT_FOUND", "Challenge session")
	}

	// 3. Submit verification with code
	id := sessionValue.Identifier
	result, err := u.kratosService.SubmitVerificationFlow(
		ctx, tenantID, flowID, &id, constants.IdentifierType(sessionValue.IdentifierType), &code,
	)
	if err != nil {
		return nil, domainerrors.NewValidationError("MSG_VERIFICATION_FAILED", "Verification failed", []interface{}{err.Error()})
	}

	// 4. Check state/result
	verified := false
	if result != nil {
		if s, ok := result.State.(string); ok && strings.EqualFold(s, constants.StatePassedChallenge) {
			verified = true
		}
	}

	if !verified {
		// Do NOT delete the session here; allow user to retry
		return nil, domainerrors.NewValidationError("MSG_VERIFICATION_FAILED", "Invalid or expired verification code", nil)
	}

	// 5. Cleanup session
	_ = u.challengeSessionRepo.DeleteChallenge(ctx, flowID)

	// 6. Response
	return &types.IdentityVerificationResponse{
		FlowID:         flowID,
		Identifier:     sessionValue.Identifier,
		IdentifierType: sessionValue.IdentifierType,
		Verified:       verified,
		VerifiedAt:     time.Now().Unix(),
	}, nil
}
