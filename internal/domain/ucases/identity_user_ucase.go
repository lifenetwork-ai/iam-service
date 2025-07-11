package ucases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	middleware "github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	ucasetypes "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
	client "github.com/ory/kratos-client-go"
)

const (
	// DefaultChallengeDuration is the default duration for a challenge session
	DefaultChallengeDuration = 5 * time.Minute // TODO: this should be configurable
)

type userUseCase struct {
	db                        *gorm.DB
	tenantRepo                repositories.TenantRepository
	kratosService             services.KratosService
	globalUserRepo            repositories.GlobalUserRepository
	userIdentityRepo          repositories.UserIdentityRepository
	userIdentifierMappingRepo repositories.UserIdentifierMappingRepository
	challengeSessionRepo      repositories.ChallengeSessionRepository
}

func NewIdentityUserUseCase(
	db *gorm.DB,
	challengeSessionRepo repositories.ChallengeSessionRepository,
	tenantRepo repositories.TenantRepository,
	globalUserRepo repositories.GlobalUserRepository,
	userIdentityRepo repositories.UserIdentityRepository,
	userIdentifierMappingRepo repositories.UserIdentifierMappingRepository,
	kratosService services.KratosService,
) ucasetypes.IdentityUserUseCase {
	return &userUseCase{
		db:                        db,
		challengeSessionRepo:      challengeSessionRepo,
		tenantRepo:                tenantRepo,
		kratosService:             kratosService,
		globalUserRepo:            globalUserRepo,
		userIdentityRepo:          userIdentityRepo,
		userIdentifierMappingRepo: userIdentifierMappingRepo,
	}
}

// ChallengeWithPhone challenges the user with a phone number
func (u *userUseCase) ChallengeWithPhone(
	ctx context.Context,
	tenantID uuid.UUID,
	phone string,
) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse) {
	_, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_GET_TENANT_FAILED",
			Message: "Failed to get tenant",
		}
	}

	if !utils.IsPhoneNumber(phone) {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_PHONE_NUMBER",
			Message: "Invalid phone number",
			Details: []interface{}{
				map[string]string{
					"field": "phone",
					"error": "Invalid phone number",
				},
			},
		}
	}

	// Initialize verification flow with Kratos
	flow, err := u.kratosService.InitializeLoginFlow(ctx, tenantID)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "VERIFICATION_FLOW_FAILED",
			Message: "Failed to initialize verification flow",
			Details: []any{err.Error()},
		}
	}

	// Submit login flow to Kratos
	_, err = u.kratosService.SubmitLoginFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), &phone, nil, nil)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "LOGIN_FAILED",
			Message: "Login failed",
			Details: []any{err.Error()},
		}
	}

	// Create challenge session
	err = u.challengeSessionRepo.SaveChallenge(ctx, flow.Id, &domain.ChallengeSession{
		Type:  "phone",
		Phone: phone,
		Flow:  flow.Id,
	}, DefaultChallengeDuration)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_SAVING_SESSION_FAILED",
			Message: "Saving challenge session failed",
			Details: []interface{}{err.Error()},
		}
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
) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse) {
	if !utils.IsEmail(email) {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_EMAIL",
			Message: "Invalid email",
			Details: []any{
				map[string]string{
					"field": "email",
					"error": "Invalid email",
				},
			},
		}
	}

	// Initialize login flow with Kratos
	flow, err := u.kratosService.InitializeLoginFlow(ctx, tenantID)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "VERIFICATION_FLOW_FAILED",
			Message: "Failed to initialize login flow",
			Details: []any{err.Error()},
		}
	}

	// Submit login flow to Kratos
	_, err = u.kratosService.SubmitLoginFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), &email, nil, nil)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "VERIFICATION_FLOW_FAILED",
			Message: "Failed to submit login flow",
			Details: []any{err.Error()},
		}
	}

	// Create challenge session
	sessionID := uuid.New().String()
	err = u.challengeSessionRepo.SaveChallenge(ctx, sessionID, &domain.ChallengeSession{
		Type:  "email",
		Email: email,
		Flow:  flow.Id,
	}, DefaultChallengeDuration)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_SAVING_SESSION_FAILED",
			Message: "Saving challenge session failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Return challenge session
	return &dto.IdentityUserChallengeDTO{
		FlowID:      flow.Id,
		Receiver:    email,
		ChallengeAt: time.Now().Unix(),
	}, nil
}

func (u *userUseCase) VerifyRegister(
	ctx context.Context,
	tenantID uuid.UUID,
	flowID string,
	code string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	flow, err := u.kratosService.GetRegistrationFlow(ctx, tenantID, flowID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get registration flow: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_GET_FLOW_FAILED",
			Message: "Failed to get registration flow",
		}
	}

	// Submit registration flow with code
	registrationResult, err := u.kratosService.SubmitRegistrationFlowWithCode(ctx, tenantID, flow, code)
	if err != nil {
		logger.GetLogger().Errorf("Failed to submit registration flow with code: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_REGISTRATION_FAILED",
			Message: "Registration failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Extract traits
	traits, ok := registrationResult.Session.Identity.Traits.(map[string]interface{})
	if !ok {
		logger.GetLogger().Errorf("Failed to parse identity traits: %v", traits)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "INVALID_TRAITS",
			Message: "Failed to parse identity traits",
		}
	}

	email := extractStringFromTraits(traits, "email", "")
	phone := extractStringFromTraits(traits, "phone_number", "")
	tenant := extractStringFromTraits(traits, "tenant", "")
	tenantUserID := registrationResult.Session.Identity.Id

	// IAM mapping logic
	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var globalUserID string

		// Try to find existing global user
		if email != "" {
			identity, err := u.userIdentityRepo.GetByTypeAndValue(ctx, tx, constants.IdentifierEmail.String(), email)
			if err == nil {
				globalUserID = identity.GlobalUserID
			}
		}
		if globalUserID == "" && phone != "" {
			identity, err := u.userIdentityRepo.GetByTypeAndValue(ctx, tx, constants.IdentifierPhone.String(), phone)
			if err == nil {
				globalUserID = identity.GlobalUserID
			}
		}

		var globalUser *domain.GlobalUser
		if globalUserID != "" {
			globalUser = &domain.GlobalUser{ID: globalUserID}

			// Check if mapping already exists
			exists, err := u.userIdentifierMappingRepo.ExistsByTenantAndTenantUserID(ctx, tx, tenant, tenantUserID)
			if err != nil {
				return fmt.Errorf("check mapping exists: %w", err)
			}
			if exists {
				return nil
			}
		} else {
			// Create global user
			globalUser = &domain.GlobalUser{}
			if err := u.globalUserRepo.Create(tx, globalUser); err != nil {
				return fmt.Errorf("create global user: %w", err)
			}
		}

		// Create UserIdentity records
		if email != "" {
			if err := u.userIdentityRepo.FirstOrCreate(tx, &domain.UserIdentity{
				GlobalUserID: globalUser.ID,
				Type:         constants.IdentifierEmail.String(),
				Value:        email,
			}); err != nil {
				return fmt.Errorf("email identity: %w", err)
			}
		}
		if phone != "" {
			if err := u.userIdentityRepo.FirstOrCreate(tx, &domain.UserIdentity{
				GlobalUserID: globalUser.ID,
				Type:         constants.IdentifierPhone.String(),
				Value:        phone,
			}); err != nil {
				return fmt.Errorf("phone identity: %w", err)
			}
		}

		// Create mapping
		if err := u.userIdentifierMappingRepo.Create(tx, &domain.UserIdentifierMapping{
			GlobalUserID: globalUser.ID,
			Tenant:       tenant,
			TenantUserID: tenantUserID,
		}); err != nil {
			return fmt.Errorf("create mapping: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "IAM_REGISTRATION_FAILED",
			Message: "Failed to persist IAM records",
			Details: []any{err.Error()},
		}
	}

	// Return authentication response
	return &dto.IdentityUserAuthDTO{
		SessionID:       registrationResult.Session.Id,
		SessionToken:    *registrationResult.SessionToken,
		Active:          *registrationResult.Session.Active,
		ExpiresAt:       registrationResult.Session.ExpiresAt,
		IssuedAt:        registrationResult.Session.IssuedAt,
		AuthenticatedAt: registrationResult.Session.AuthenticatedAt,
		User: dto.IdentityUserDTO{
			ID:       tenantUserID,
			UserName: extractStringFromTraits(traits, "username", ""),
			Email:    email,
			Phone:    phone,
		},
		AuthenticationMethods: utils.Map(registrationResult.Session.AuthenticationMethods, func(method client.SessionAuthenticationMethod) string {
			return *method.Method
		}),
	}, nil
}

// VerifyLogin verifies the login flow
func (u *userUseCase) VerifyLogin(
	ctx context.Context,
	tenantID uuid.UUID,
	flowID string,
	code string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	// Get the login flow
	flow, err := u.kratosService.GetLoginFlow(ctx, tenantID, flowID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get registration flow: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_GET_FLOW_FAILED",
			Message: "Failed to get login flow",
		}
	}

	// Get the challenge session to retrieve the phone number
	sessionValue, err := u.challengeSessionRepo.GetChallenge(ctx, flowID)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "MSG_CHALLENGE_SESSION_NOT_FOUND",
			Message: "Challenge session not found",
			Details: []interface{}{err.Error()},
		}
	}

	if sessionValue == nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "MSG_CHALLENGE_SESSION_NOT_FOUND",
			Message: "Challenge session not found",
			Details: []interface{}{
				map[string]string{"field": "session", "error": "Session not found"},
			},
		}
	}

	// Submit login flow with code
	loginResult, err := u.kratosService.SubmitLoginFlow(
		ctx, tenantID, flow, constants.MethodTypeCode.String(), &sessionValue.Phone, nil, &code,
	)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_LOGIN_FAILED",
			Message: "Login failed",
			Details: []interface{}{err.Error()},
		}
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
			UserName: extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), "username", ""),
			Email:    extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), "email", ""),
			Phone:    extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), "phone_number", ""),
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
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	// Get the challenge session
	sessionValue, err := u.challengeSessionRepo.GetChallenge(ctx, sessionID)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "MSG_CHALLENGE_SESSION_NOT_FOUND",
			Message: "Challenge session not found",
			Details: []interface{}{err.Error()},
		}
	}

	if sessionValue == nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "MSG_CHALLENGE_SESSION_NOT_FOUND",
			Message: "Challenge session not found",
			Details: []interface{}{
				map[string]string{"field": "session", "error": "Session not found"},
			},
		}
	}

	// Get the verification flow
	flow, err := u.kratosService.GetVerificationFlow(ctx, tenantID, sessionValue.Flow)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get verification flow: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_GET_FLOW_FAILED",
			Message: "Failed to get verification flow",
			Details: []interface{}{err.Error()},
		}
	}

	// Submit verification flow with code
	_, err = u.kratosService.SubmitVerificationFlow(ctx, tenantID, flow, code)
	if err != nil {
		logger.GetLogger().Errorf("Failed to submit verification flow with code: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_VERIFICATION_FAILED",
			Message: "Verification failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Get session
	session, err := u.kratosService.GetSession(ctx, tenantID, sessionValue.Flow)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get session: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_INVALID_SESSION",
			Message: "Invalid session",
			Details: []interface{}{err.Error()},
		}
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
			UserName: extractStringFromTraits(session.Identity.Traits.(map[string]interface{}), "username", ""),
			Email:    extractStringFromTraits(session.Identity.Traits.(map[string]interface{}), "email", ""),
			Phone:    extractStringFromTraits(session.Identity.Traits.(map[string]interface{}), "phone_number", ""),
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
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	// Validate phone number if provided
	if payload.Phone != "" && !utils.IsPhoneNumber(payload.Phone) {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_PHONE_NUMBER",
			Message: "Invalid phone number format",
			Details: []any{"Phone number must be in international format (e.g., +1234567890)"},
		}
	}

	// Check if identifier (email/phone) already exists in IAM
	if payload.Email != "" {
		globalUserID, err := u.userIdentityRepo.FindGlobalUserIDByIdentity(ctx, constants.IdentifierEmail.String(), payload.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "IAM_LOOKUP_FAILED",
				Message: "Failed to check existing email identity",
				Details: []any{err.Error()},
			}
		}
		if globalUserID != "" {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusConflict,
				Code:    "EMAIL_ALREADY_EXISTS",
				Message: "Email has already been registered",
			}
		}
	}

	if payload.Phone != "" {
		globalUserID, err := u.userIdentityRepo.FindGlobalUserIDByIdentity(ctx, constants.IdentifierPhone.String(), payload.Phone)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "IAM_LOOKUP_FAILED",
				Message: "Failed to check existing phone identity",
				Details: []any{err.Error()},
			}
		}
		if globalUserID != "" {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusConflict,
				Code:    "PHONE_ALREADY_EXISTS",
				Message: "Phone number has already been registered",
			}
		}
	}

	// Initialize registration flow with Kratos
	flow, err := u.kratosService.InitializeRegistrationFlow(ctx, tenantID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to initialize registration flow: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INITIALIZE_REGISTRATION_FAILED",
			Message: "Failed to initialize registration flow",
			Details: []interface{}{err.Error()},
		}
	}

	tenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to initialize registration flow: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_GET_TENANT_FAILED",
			Message: "Failed to get tenant",
		}
	}

	// Prepare traits
	traits := map[string]interface{}{
		"tenant": tenant.Name,
	}
	if payload.Email != "" {
		traits["email"] = payload.Email
	}
	if payload.Phone != "" {
		traits["phone_number"] = payload.Phone
	}

	// Submit registration flow
	_, err = u.kratosService.SubmitRegistrationFlow(ctx, tenantID, flow, constants.MethodTypeCode.String(), traits)
	if err != nil {
		logger.GetLogger().Errorf("Failed to submit registration flow: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_REGISTRATION_FAILED",
			Message: "Registration failed",
			Details: []interface{}{err.Error()},
		}
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

// LogIn logs in a user with username and password
func (u *userUseCase) LogIn(
	ctx context.Context,
	tenantID uuid.UUID,
	username string,
	password string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	// Initialize login flow
	flow, err := u.kratosService.InitializeLoginFlow(ctx, tenantID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to initialize login flow: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INITIALIZE_LOGIN_FAILED",
			Message: "Failed to initialize login flow",
			Details: []interface{}{err.Error()},
		}
	}

	// Submit login flow to Kratos
	loginResult, err := u.kratosService.SubmitLoginFlow(ctx, tenantID, flow, constants.MethodTypePassword.String(), &username, &password, nil)
	if err != nil {
		logger.GetLogger().Errorf("Failed to submit login flow: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_LOGIN_FAILED",
			Message: "Login failed",
			Details: []interface{}{err.Error()},
		}
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
			UserName: extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), "username", ""),
			Email:    extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), "email", ""),
			Phone:    extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), "phone_number", ""),
		},
		AuthenticationMethods: utils.Map(loginResult.Session.AuthenticationMethods, func(method client.SessionAuthenticationMethod) string {
			return *method.Method
		}),
	}, nil
}

// LogOut logs out a user
func (u *userUseCase) LogOut(
	ctx context.Context,
	tenantID uuid.UUID,
) *dto.ErrorDTOResponse {
	// Get session token from context
	sessionToken := ctx.Value("session_token").(string)
	if sessionToken == "" {
		return &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_UNAUTHORIZED",
			Message: "Unauthorized",
			Details: []interface{}{
				map[string]string{"field": "session_token", "error": "Session token not found"},
			},
		}
	}

	// Revoke session
	if err := u.kratosService.Logout(ctx, tenantID, sessionToken); err != nil {
		logger.GetLogger().Errorf("Failed to logout: %v", err)
		return &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_LOGOUT_FAILED",
			Message: "Failed to logout",
			Details: []interface{}{err.Error()},
		}
	}

	return nil
}

// RefreshToken refreshes a user's session token
func (u *userUseCase) RefreshToken(
	ctx context.Context,
	tenantID uuid.UUID,
	accessToken string,
	refreshToken string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	// Get session
	session, err := u.kratosService.GetSession(ctx, tenantID, accessToken)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get session: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_INVALID_SESSION",
			Message: "Invalid session",
			Details: []interface{}{err.Error()},
		}
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
			UserName: extractStringFromTraits(session.Identity.Traits.(map[string]interface{}), "username", ""),
			Email:    extractStringFromTraits(session.Identity.Traits.(map[string]interface{}), "email", ""),
			Phone:    extractStringFromTraits(session.Identity.Traits.(map[string]interface{}), "phone_number", ""),
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
) (*dto.IdentityUserDTO, *dto.ErrorDTOResponse) {
	// Get session token from context
	sessionTokenVal := ctx.Value(middleware.SessionTokenKey)
	if sessionTokenVal == nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_UNAUTHORIZED",
			Message: "Unauthorized",
			Details: []interface{}{
				map[string]string{"field": "session_token", "error": "Session token not found"},
			},
		}
	}

	sessionToken, ok := sessionTokenVal.(string)
	if !ok || sessionToken == "" {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_UNAUTHORIZED",
			Message: "Unauthorized",
			Details: []interface{}{
				map[string]string{"field": "session_token", "error": "Invalid session token format"},
			},
		}
	}

	// Get session
	session, err := u.kratosService.WhoAmI(ctx, tenantID, sessionToken)
	if err != nil {
		logger.GetLogger().Errorf("Failed to extract user traits: %v", err)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_INVALID_SESSION",
			Message: "Invalid session",
			Details: []interface{}{err.Error()},
		}
	}

	// Extract user traits
	user, err := extractUserFromTraits(session.Identity.Traits, "", "")
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_EXTRACT_USER_FAILED",
			Message: "Failed to extract user traits",
			Details: []interface{}{err.Error()},
		}
	}

	return &user, nil
}

// safeExtractTraits safely converts interface{} to map[string]interface{}
// Returns the map and a boolean indicating success
func safeExtractTraits(traits interface{}) (map[string]interface{}, bool) {
	if traits == nil {
		return make(map[string]interface{}), false
	}

	// Direct type assertion (most common case)
	if traitsMap, ok := traits.(map[string]interface{}); ok {
		return traitsMap, true
	}

	// Fallback: JSON marshal/unmarshal for complex cases
	jsonBytes, err := json.Marshal(traits)
	if err != nil {
		logger.GetLogger().Errorf("Failed to marshal traits: %v", err)
		return make(map[string]interface{}), false
	}

	var traitsMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &traitsMap); err != nil {
		logger.GetLogger().Errorf("Failed to unmarshal traits: %v", err)
		return make(map[string]interface{}), false
	}

	return traitsMap, true
}

// extractUserFromTraits safely extracts user data from traits
func extractUserFromTraits(traits interface{}, fallbackEmail, fallbackPhone string) (dto.IdentityUserDTO, error) {
	traitsMap, ok := safeExtractTraits(traits)
	if !ok {
		return dto.IdentityUserDTO{}, fmt.Errorf("unable to extract traits from interface{}")
	}

	return dto.IdentityUserDTO{
		UserName: extractStringFromTraits(traitsMap, "username", ""),
		Email:    extractStringFromTraits(traitsMap, "email", fallbackEmail),
		Phone:    extractStringFromTraits(traitsMap, "phone_number", fallbackPhone),
		Tenant:   extractStringFromTraits(traitsMap, "tenant", ""),
	}, nil
}

// extractStringFromTraits extracts a string value from traits map
// If the value is a pointer to a string, it dereferences it
// If the value is nil, it returns the default value
// If the value is not a string, it returns the default value
func extractStringFromTraits(traits map[string]interface{}, key, defaultValue string) string {
	if traits == nil {
		return defaultValue
	}

	value, exists := traits[key]
	if !exists {
		return defaultValue
	}

	// Handle different types that might be stored
	switch v := value.(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
		return defaultValue
	case nil:
		return defaultValue
	default:
		// Convert other types to string as fallback
		return fmt.Sprintf("%v", v)
	}
}
