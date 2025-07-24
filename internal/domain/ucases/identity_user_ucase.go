package ucases

import (
	"context"
	"fmt"
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

	if !utils.IsPhoneNumber(phone) {
		return nil, domainerrors.NewValidationError(
			"MSG_INVALID_PHONE_NUMBER",
			"Invalid phone number",
			[]interface{}{
				map[string]string{
					"field": "phone",
					"error": "Invalid phone number",
				},
			},
		)
	}

	// Check rate limit for phone challenges
	key := fmt.Sprintf("challenge:phone:tenant:%s:%s", phone, tenantID.String())
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
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
		Type:  constants.IdentifierPhone.String(),
		Phone: phone,
		Flow:  flow.Id,
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
		Type:  constants.IdentifierEmail.String(),
		Email: email,
		Flow:  flow.Id,
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

	email := extractStringFromTraits(traits, constants.IdentifierEmail.String(), "")
	phone := extractStringFromTraits(traits, constants.IdentifierPhone.String(), "")
	tenantName := extractStringFromTraits(traits, "tenant", "")
	newTenantUserID := registrationResult.Session.Identity.Id

	// Determine identifier and type
	identifier := email
	if phone != "" {
		identifier = phone
	}
	identifierType, err := utils.GetIdentifierType(identifier)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_GET_IDENTIFIER_TYPE_FAILED", "Failed to get identifier type")
	}

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

	// Bind IAM to registration
	if err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return u.bindIAMToRegistration(ctx, tx, tenant, newTenantUserID, identifier, identifierType)
	}); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_IAM_REGISTRATION_FAILED", "Failed to bind IAM to registration")
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
			ID:       newTenantUserID,
			UserName: extractStringFromTraits(traits, constants.IdentifierUsername.String(), ""),
			Email:    email,
			Phone:    phone,
		},
		AuthenticationMethods: utils.Map(registrationResult.Session.AuthenticationMethods, func(method client.SessionAuthenticationMethod) string {
			return *method.Method
		}),
	}, nil
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
	if identity, err := u.userIdentityRepo.GetByTypeAndValue(ctx, tx, identifierType, identifier); err == nil {
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
	}

	// Create identities
	if err := u.userIdentityRepo.FirstOrCreate(tx, &domain.UserIdentity{
		GlobalUserID: globalUser.ID,
		Type:         identifierType,
		Value:        identifier,
	}); err != nil {
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
		logger.GetLogger().Errorf("Failed to get registration flow: %v", err)
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
	identifier := sessionValue.Phone
	if sessionValue.Email != "" {
		identifier = sessionValue.Email
	}
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

// ChallengeVerify verifies a challenge
func (u *userUseCase) ChallengeVerify(
	ctx context.Context,
	tenantID uuid.UUID,
	sessionID string,
	code string,
) (*types.IdentityUserAuthResponse, *domainerrors.DomainError) {
	// Check rate limit for verification attempts
	key := fmt.Sprintf("verify:%s", sessionID)
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// Get the challenge session
	sessionValue, err := u.challengeSessionRepo.GetChallenge(ctx, sessionID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_CHALLENGE_SESSION_NOT_FOUND", "Challenge session not found")
	}

	if sessionValue == nil {
		return nil, domainerrors.NewNotFoundError("MSG_CHALLENGE_SESSION_NOT_FOUND", "Challenge session").WithDetails([]interface{}{
			map[string]string{"field": "session", "error": "Session not found"},
		})
	}

	// Get the verification flow
	flow, err := u.kratosService.GetVerificationFlow(ctx, tenantID, sessionValue.Flow)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get verification flow: %v", err)
		return nil, domainerrors.WrapInternal(err, "MSG_GET_FLOW_FAILED", "Failed to get verification flow")
	}

	// Submit verification flow with code
	_, err = u.kratosService.SubmitVerificationFlow(ctx, tenantID, flow.Id, &code)
	if err != nil {
		logger.GetLogger().Errorf("Failed to submit verification flow with code: %v", err)
		return nil, domainerrors.NewValidationError("MSG_VERIFICATION_FAILED", "Verification failed", []interface{}{err.Error()})
	}

	// Get session
	session, err := u.kratosService.GetSession(ctx, tenantID, sessionValue.Flow)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get session: %v", err)
		return nil, domainerrors.NewUnauthorizedError("MSG_INVALID_SESSION", "Invalid session").WithCause(err)
	}

	// Return authentication response
	return &types.IdentityUserAuthResponse{
		SessionID:       session.Id,
		SessionToken:    sessionValue.Flow,
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

// Register registers a new user
func (u *userUseCase) Register(
	ctx context.Context,
	tenantID uuid.UUID,
	email string,
	phone string,
) (*types.IdentityUserAuthResponse, *domainerrors.DomainError) {
	// Validate phone number if provided
	if phone != "" && !utils.IsPhoneNumber(phone) {
		return nil, domainerrors.NewValidationError("MSG_INVALID_PHONE_NUMBER", "Invalid phone number format", []any{"Phone number must be in international format (e.g., +1234567890)"})
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
	}
	if email != "" {
		traits[constants.IdentifierEmail.String()] = email
	}
	if phone != "" {
		traits[constants.IdentifierPhone.String()] = phone
	}

	// Submit registration flow
	_, err = u.kratosService.SubmitRegistrationFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), traits)
	if err != nil {
		logger.GetLogger().Errorf("Failed to submit registration flow: %v", err)
		return nil, domainerrors.NewValidationError("MSG_REGISTRATION_FAILED", "Registration failed", []interface{}{err.Error()})
	}
	receiver := email
	if receiver == "" {
		receiver = phone
	}

	// Return success with verification flow info
	return &types.IdentityUserAuthResponse{
		VerificationNeeded: true,
		VerificationFlow: &types.IdentityUserChallengeResponse{
			FlowID:      flow.Id,
			Receiver:    receiver,
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

	// Set user id
	user.ID = session.Identity.Id

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
	if identifierType == constants.IdentifierEmail.String() && !utils.IsEmail(identifier) {
		return nil, domainerrors.NewValidationError("MSG_INVALID_EMAIL", "Invalid email", nil)
	}
	if identifierType == constants.IdentifierPhone.String() && !utils.IsPhoneNumber(identifier) {
		return nil, domainerrors.NewValidationError("MSG_INVALID_PHONE_NUMBER", "Invalid phone number", nil)
	}

	// 2. Check if identifier already exists globally
	exists, err := u.userIdentityRepo.ExistsWithinTenant(ctx, tenantID.String(), identifierType, identifier)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_IAM_LOOKUP_FAILED", "Failed to check existing identifier")
	}
	if exists {
		return nil, domainerrors.NewConflictError("MSG_IDENTIFIER_ALREADY_EXISTS", "Identifier has already been registered", nil)
	}

	// 2b. Check if user already has this identifier type
	hasType, err := u.userIdentityRepo.ExistsByGlobalUserIDAndType(ctx, globalUserID, identifierType)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_CHECK_TYPE_EXIST_FAILED", "Failed to check user identity type")
	}
	if hasType {
		return nil, domainerrors.NewConflictError("MSG_IDENTIFIER_TYPE_EXISTS", fmt.Sprintf("User already has an identifier of type %s", identifierType), nil)
	}

	// 3. Rate limit
	key := fmt.Sprintf("challenge:add:%s:%s", identifierType, identifier)
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// 4. Init Kratos Registration Flow
	flow, err := u.kratosService.InitializeRegistrationFlow(ctx, tenantID)
	if err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_INIT_REG_FLOW_FAILED", "Failed to initialize registration flow")
	}

	// 5. Submit minimal traits to trigger OTP (email or phone)
	traits := map[string]interface{}{
		identifierType: identifier,
	}
	if _, err := u.kratosService.SubmitRegistrationFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), traits); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_ADD_IDENTIFIER_FAILED", "Adding identifier failed").WithCause(err)
	}

	// 6. Save challenge session
	session := &domain.ChallengeSession{
		Type:         identifierType,
		Flow:         flow.Id,
		GlobalUserID: globalUserID,
	}
	if identifierType == constants.IdentifierEmail.String() {
		session.Email = identifier
	}
	if identifierType == constants.IdentifierPhone.String() {
		session.Phone = identifier
	}

	if err := u.challengeSessionRepo.SaveChallenge(ctx, flow.Id, session, constants.DefaultChallengeDuration); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_SAVE_CHALLENGE_FAILED", "Failed to save challenge session")
	}

	// 7. Return response
	return &types.IdentityUserChallengeResponse{
		FlowID:      flow.Id,
		Receiver:    identifier,
		ChallengeAt: time.Now().Unix(),
	}, nil
}
