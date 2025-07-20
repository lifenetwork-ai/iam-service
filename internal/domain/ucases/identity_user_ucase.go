package ucases

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	ratelimiters "github.com/lifenetwork-ai/iam-service/infrastructures/ratelimiters/types"
	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	kratos_types "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos/types"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	middleware "github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	ucasetypes "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
	client "github.com/ory/kratos-client-go"
)

type userUseCase struct {
	db                        *gorm.DB
	rateLimiter               ratelimiters.RateLimiter
	tenantRepo                repositories.TenantRepository
	globalUserRepo            repositories.GlobalUserRepository
	userIdentityRepo          repositories.UserIdentityRepository
	userIdentifierMappingRepo repositories.UserIdentifierMappingRepository
	challengeSessionRepo      repositories.ChallengeSessionRepository
	kratosService             kratos_types.KratosService
}

func NewIdentityUserUseCase(
	db *gorm.DB,
	rateLimiter ratelimiters.RateLimiter,
	challengeSessionRepo repositories.ChallengeSessionRepository,
	tenantRepo repositories.TenantRepository,
	globalUserRepo repositories.GlobalUserRepository,
	userIdentityRepo repositories.UserIdentityRepository,
	userIdentifierMappingRepo repositories.UserIdentifierMappingRepository,
	kratosService kratos_types.KratosService,
) ucasetypes.IdentityUserUseCase {
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
) (*dto.IdentityUserChallengeDTO, *domainerrors.DomainError) {
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
	key := "challenge:phone:" + phone
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

	return &dto.IdentityUserChallengeDTO{
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
) (*dto.IdentityUserChallengeDTO, *domainerrors.DomainError) {
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
	key := "challenge:email:" + email
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
	return &dto.IdentityUserChallengeDTO{
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
) (*dto.IdentityUserAuthDTO, *domainerrors.DomainError) {
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
	return &dto.IdentityUserAuthDTO{
		SessionID:       registrationResult.Session.Id,
		SessionToken:    *registrationResult.SessionToken,
		Active:          *registrationResult.Session.Active,
		ExpiresAt:       registrationResult.Session.ExpiresAt,
		IssuedAt:        registrationResult.Session.IssuedAt,
		AuthenticatedAt: registrationResult.Session.AuthenticatedAt,
		User: dto.IdentityUserDTO{
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
) (*dto.IdentityUserAuthDTO, *domainerrors.DomainError) {
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
	return &dto.IdentityUserAuthDTO{
		SessionID:       loginResult.Session.Id,
		SessionToken:    *loginResult.SessionToken,
		Active:          *loginResult.Session.Active,
		ExpiresAt:       loginResult.Session.ExpiresAt,
		IssuedAt:        loginResult.Session.IssuedAt,
		AuthenticatedAt: loginResult.Session.AuthenticatedAt,
		User: dto.IdentityUserDTO{
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
) (*dto.IdentityUserAuthDTO, *domainerrors.DomainError) {
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
	_, err = u.kratosService.SubmitVerificationFlow(ctx, tenantID, flow, code)
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
	return &dto.IdentityUserAuthDTO{
		SessionID:       session.Id,
		SessionToken:    sessionValue.Flow,
		Active:          *session.Active,
		ExpiresAt:       session.ExpiresAt,
		IssuedAt:        session.IssuedAt,
		AuthenticatedAt: session.AuthenticatedAt,
		User: dto.IdentityUserDTO{
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
	payload dto.IdentityUserRegisterDTO,
) (*dto.IdentityUserAuthDTO, *domainerrors.DomainError) {
	// Validate phone number if provided
	if payload.Phone != "" && !utils.IsPhoneNumber(payload.Phone) {
		return nil, domainerrors.NewValidationError("MSG_INVALID_PHONE_NUMBER", "Invalid phone number format", []any{"Phone number must be in international format (e.g., +1234567890)"})
	}

	// Check rate limit for registration attempts
	key := "register:" + payload.Email
	if payload.Phone != "" {
		key = "register:" + payload.Phone
	}
	if err := utils.CheckRateLimitDomain(u.rateLimiter, key, constants.MaxAttemptsPerWindow, constants.RateLimitWindow); err != nil {
		return nil, domainerrors.WrapInternal(err, "MSG_RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// Check if identifier (email/phone) already exists in IAM
	if payload.Email != "" {
		exists, err := u.userIdentityRepo.ExistsWithinTenant(ctx, tenantID.String(), constants.IdentifierEmail.String(), payload.Email)
		if err != nil {
			return nil, domainerrors.WrapInternal(err, "MSG_IAM_LOOKUP_FAILED", "Failed to check existing email identity")
		}
		if exists {
			return nil, domainerrors.NewConflictError("MSG_EMAIL_ALREADY_EXISTS", "Email has already been registered", nil)
		}
	}

	if payload.Phone != "" {
		exists, err := u.userIdentityRepo.ExistsWithinTenant(ctx, tenantID.String(), constants.IdentifierPhone.String(), payload.Phone)
		if err != nil {
			return nil, domainerrors.WrapInternal(err, "MSG_IAM_LOOKUP_FAILED", "Failed to check existing phone identity")
		}

		if exists {
			return nil, domainerrors.NewConflictError("MSG_PHONE_ALREADY_EXISTS", "Phone number has already been registered", nil)
		}
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
	if payload.Email != "" {
		traits[constants.IdentifierEmail.String()] = payload.Email
	}
	if payload.Phone != "" {
		traits[constants.IdentifierPhone.String()] = payload.Phone
	}

	// Submit registration flow
	_, err = u.kratosService.SubmitRegistrationFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), traits)
	if err != nil {
		logger.GetLogger().Errorf("Failed to submit registration flow: %v", err)
		return nil, domainerrors.NewValidationError("MSG_REGISTRATION_FAILED", "Registration failed", []interface{}{err.Error()})
	}
	receiver := payload.Email
	if receiver == "" {
		receiver = payload.Phone
	}

	// Return success with verification flow info
	return &dto.IdentityUserAuthDTO{
		VerificationNeeded: true,
		VerificationFlow: &dto.IdentityUserChallengeDTO{
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
) (*dto.IdentityUserAuthDTO, *domainerrors.DomainError) {
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
	return &dto.IdentityUserAuthDTO{
		SessionID:       loginResult.Session.Id,
		SessionToken:    *loginResult.SessionToken,
		Active:          *loginResult.Session.Active,
		ExpiresAt:       loginResult.Session.ExpiresAt,
		IssuedAt:        loginResult.Session.IssuedAt,
		AuthenticatedAt: loginResult.Session.AuthenticatedAt,
		User: dto.IdentityUserDTO{
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
	sessionToken, err := u.extractSessionToken(ctx)
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
) (*dto.IdentityUserAuthDTO, *domainerrors.DomainError) {
	// Get session
	session, err := u.kratosService.GetSession(ctx, tenantID, accessToken)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get session: %v", err)
		return nil, domainerrors.NewUnauthorizedError("MSG_INVALID_SESSION", "Invalid session").WithCause(err)
	}

	// Return authentication response
	return &dto.IdentityUserAuthDTO{
		SessionID:       session.Id,
		SessionToken:    accessToken,
		Active:          *session.Active,
		ExpiresAt:       session.ExpiresAt,
		IssuedAt:        session.IssuedAt,
		AuthenticatedAt: session.AuthenticatedAt,
		User: dto.IdentityUserDTO{
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
) (*dto.IdentityUserDTO, *domainerrors.DomainError) {
	// Get session token from context
	sessionToken, sessionTokenErr := u.extractSessionToken(ctx)
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

	return &user, nil
}

// extractSessionToken extracts and validates the session token from context
func (u *userUseCase) extractSessionToken(ctx context.Context) (string, *domainerrors.DomainError) {
	sessionTokenVal := ctx.Value(middleware.SessionTokenKey)
	if sessionTokenVal == nil {
		return "", domainerrors.NewUnauthorizedError("MSG_UNAUTHORIZED", "Unauthorized").WithDetails([]interface{}{
			map[string]string{"field": "session_token", "error": "Session token not found"},
		})
	}

	sessionToken, ok := sessionTokenVal.(string)
	if !ok || sessionToken == "" {
		return "", domainerrors.NewUnauthorizedError("MSG_UNAUTHORIZED", "Unauthorized").WithDetails([]interface{}{
			map[string]string{"field": "session_token", "error": "Invalid session token format"},
		})
	}

	return sessionToken, nil
}
