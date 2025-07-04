package ucases

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	ucase_interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
)

type userUseCase struct {
	challengeSessionRepo repositories.ChallengeSessionRepository
	kratosService        services.KratosService
}

func NewIdentityUserUseCase(
	challengeSessionRepo repositories.ChallengeSessionRepository,
	kratosService services.KratosService,
) ucase_interfaces.IdentityUserUseCase {
	return &userUseCase{
		challengeSessionRepo: challengeSessionRepo,
		kratosService:        kratosService,
	}
}

// ChallengeWithPhone challenges the user with a phone number
func (u *userUseCase) ChallengeWithPhone(
	ctx context.Context,
	phone string,
) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse) {
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
	flow, err := u.kratosService.InitializeLoginFlow(ctx)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "VERIFICATION_FLOW_FAILED",
			Message: "Failed to initialize verification flow",
			Details: []any{err.Error()},
		}
	}

	// Submit login flow to Kratos
	_, err = u.kratosService.SubmitLoginFlow(ctx, flow, "code", &phone, nil, nil)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "LOGIN_FAILED",
			Message: "Login failed",
			Details: []any{err.Error()},
		}
	}

	// Create challenge session
	err = u.challengeSessionRepo.SaveChallenge(ctx, flow.Id, &entities.ChallengeSession{
		Type:  "phone",
		Phone: phone,
		Flow:  flow.Id,
	}, 5*time.Minute)
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

	// Initialize verification flow with Kratos
	flow, err := u.kratosService.InitializeVerificationFlow(ctx)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "VERIFICATION_FLOW_FAILED",
			Message: "Failed to initialize verification flow",
			Details: []any{err.Error()},
		}
	}

	// Create challenge session
	sessionID := uuid.New().String()
	err = u.challengeSessionRepo.SaveChallenge(ctx, sessionID, &entities.ChallengeSession{
		Type:  "email",
		Email: email,
		Flow:  flow.Id,
	}, 5*time.Minute)
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

// VerifyRegister verifies the registration flow
func (u *userUseCase) VerifyRegister(
	ctx context.Context,
	flowID string,
	code string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	flow, err := u.kratosService.GetRegistrationFlow(ctx, flowID)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_GET_FLOW_FAILED",
			Message: "Failed to get registration flow",
		}
	}

	// Submit registration flow with code
	registrationResult, err := u.kratosService.SubmitRegistrationFlowWithCode(ctx, flow, code)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_REGISTRATION_FAILED",
			Message: "Registration failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Return authentication response
	return &dto.IdentityUserAuthDTO{
		AccessToken:      *registrationResult.SessionToken,
		RefreshToken:     "", // Kratos handles refresh internally
		AccessExpiresAt:  registrationResult.Session.ExpiresAt.Unix(),
		RefreshExpiresAt: 0,
		LastLoginAt:      time.Now().Unix(),
		User: dto.IdentityUserDTO{
			ID:       registrationResult.Identity.Id,
			UserName: extractStringFromTraits(registrationResult.Identity.Traits.(map[string]interface{}), "username", ""),
			Email:    extractStringFromTraits(registrationResult.Identity.Traits.(map[string]interface{}), "email", ""),
			Phone:    extractStringFromTraits(registrationResult.Identity.Traits.(map[string]interface{}), "phone_number", ""),
		},
	}, nil
}

// VerifyLogin verifies the login flow
func (u *userUseCase) VerifyLogin(
	ctx context.Context,
	flowID string,
	code string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	// Get the login flow
	flow, err := u.kratosService.GetLoginFlow(ctx, flowID)
	if err != nil {
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
	loginResult, err := u.kratosService.SubmitLoginFlow(ctx, flow, "code", &sessionValue.Phone, nil, &code)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "LOGIN_FAILED",
			Message: "Login failed",
			Details: []any{err.Error()},
		}
	}

	return &dto.IdentityUserAuthDTO{
		AccessToken:      *loginResult.SessionToken,
		RefreshToken:     "", // Kratos handles refresh internally
		AccessExpiresAt:  loginResult.Session.ExpiresAt.Unix(),
		RefreshExpiresAt: 0,
		LastLoginAt:      time.Now().Unix(),
		User: dto.IdentityUserDTO{
			ID:       loginResult.Session.Identity.Id,
			UserName: extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), "username", ""),
			Email:    extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), "email", ""),
			Phone:    extractStringFromTraits(loginResult.Session.Identity.Traits.(map[string]interface{}), "phone_number", ""),
		},
	}, nil
}

func (u *userUseCase) ChallengeVerify(
	ctx context.Context,
	sessionID string,
	code string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
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

	if sessionValue.Flow == "" {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_FLOW",
			Message: "Invalid flow",
			Details: []interface{}{
				map[string]string{"field": "flow", "error": "Missing flow in session"},
			},
		}
	}

	// Get the verification flow from Kratos
	flow, err := u.kratosService.GetVerificationFlow(ctx, sessionValue.Flow)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_VERIFICATION_FLOW_NOT_FOUND",
			Message: "Verification flow not found",
			Details: []interface{}{err.Error()},
		}
	}

	// Submit verification flow to Kratos
	verificationResult, err := u.kratosService.SubmitVerificationFlow(ctx, flow, code)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_VERIFICATION_FAILED",
			Message: "Verification failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Check if verification was successful
	if verificationResult.State != "passed_challenge" {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_VERIFICATION_FAILED",
			Message: "Verification failed",
			Details: []interface{}{"Verification state is not passed"},
		}
	}

	// Get session from verification result
	session, err := u.kratosService.GetSession(ctx, verificationResult.Id)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_GET_SESSION_FAILED",
			Message: "Failed to get Kratos session",
			Details: []interface{}{err.Error()},
		}
	}

	// Extract user information from Kratos session
	var userDTO dto.IdentityUserDTO
	if traits, ok := session.Identity.Traits.(map[string]interface{}); ok {
		userDTO = dto.IdentityUserDTO{
			ID:       session.Identity.Id,
			UserName: extractStringFromTraits(traits, "username", ""),
			Email:    extractStringFromTraits(traits, "email", sessionValue.Email),
			Phone:    extractStringFromTraits(traits, "phone", sessionValue.Phone),
		}
	} else {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_SESSION",
			Message: "Invalid session data from Kratos",
			Details: []interface{}{"Unable to extract user traits"},
		}
	}

	// Return authentication response
	return &dto.IdentityUserAuthDTO{
		AccessToken:      verificationResult.Id,
		RefreshToken:     "", // Kratos handles refresh internally
		AccessExpiresAt:  session.ExpiresAt.Unix(),
		RefreshExpiresAt: 0,
		LastLoginAt:      time.Now().Unix(),
		User:             userDTO,
	}, nil
}

func (u *userUseCase) Register(
	ctx context.Context,
	payload dto.IdentityUserRegisterDTO,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	// Initialize registration flow with Kratos
	flow, err := u.kratosService.InitializeRegistrationFlow(ctx)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "REGISTRATION_FLOW_FAILED",
			Message: "Failed to initialize registration flow",
			Details: []any{err.Error()},
		}
	}

	// Validate phone number if provided
	if payload.Phone != "" && !utils.IsPhoneNumber(payload.Phone) {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_PHONE_NUMBER",
			Message: "Invalid phone number format",
			Details: []any{"Phone number must be in international format (e.g., +1234567890)"},
		}
	}

	// Prepare traits for registration
	traits := make(map[string]any)

	if payload.Phone != "" {
		traits["phone_number"] = payload.Phone
	}
	if payload.Email != "" {
		traits["email"] = payload.Email
	}
	if payload.Tenant != "" {
		traits["tenant"] = payload.Tenant
	}

	// Submit registration flow to Kratos
	_, err = u.kratosService.SubmitRegistrationFlow(ctx, flow, "code", traits)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "REGISTRATION_FAILED",
			Message: "Registration failed",
			Details: []any{err.Error()},
		}
	}

	// Return success with verification flow info
	return &dto.IdentityUserAuthDTO{
		VerificationNeeded: true,
		VerificationFlow: &dto.IdentityUserChallengeDTO{
			FlowID:      flow.Id,
			Receiver:    payload.Phone,
			ChallengeAt: time.Now().Unix(),
		},
	}, nil
}

// Login with username and password
// Deprecated: This method is not used anymore
func (u *userUseCase) LogIn(
	ctx context.Context,
	username string,
	password string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	// Initialize login flow with Kratos
	flow, err := u.kratosService.InitializeLoginFlow(ctx)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "LOGIN_FLOW_FAILED",
			Message: "Failed to initialize login flow",
			Details: []any{err.Error()},
		}
	}

	// Submit login flow to Kratos
	session, err := u.kratosService.SubmitLoginFlow(ctx, flow, "password", &username, &password, nil)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "LOGIN_FAILED",
			Message: "Login failed",
			Details: []any{err.Error()},
		}
	}

	// Extract user information from Kratos session
	var userDTO dto.IdentityUserDTO
	userDTO, err = extractUserFromTraits(session.Session.Identity.Traits, "", "")
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_SESSION",
			Message: "Invalid session data",
			Details: []interface{}{err.Error()},
		}
	}
	userDTO.ID = session.Session.Identity.Id

	// Return authentication response
	return &dto.IdentityUserAuthDTO{
		AccessToken:      *session.SessionToken,
		RefreshToken:     "", // Kratos handles refresh internally
		AccessExpiresAt:  session.Session.ExpiresAt.Unix(),
		RefreshExpiresAt: 0,
		LastLoginAt:      time.Now().Unix(),
		User:             userDTO,
	}, nil
}

func (u *userUseCase) LogOut(
	ctx context.Context,
) *dto.ErrorDTOResponse {
	// Extract session token from context or headers
	sessionToken, exists := ctx.Value("sessionToken").(string)
	if !exists {
		return &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_SESSION_TOKEN_MISSING",
			Message: "Session token missing",
			Details: []interface{}{"No session token provided"},
		}
	}

	// Revoke session in Kratos
	err := u.kratosService.Logout(ctx, sessionToken)
	if err != nil {
		return &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "LOGOUT_FAILED",
			Message: "Failed to logout",
			Details: []interface{}{err.Error()},
		}
	}
	return nil
}

func (u *userUseCase) RefreshToken(
	ctx context.Context,
	accessToken string,
	refreshToken string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	// With Kratos, session refresh is handled automatically
	// We just need to validate the current session
	session, err := u.kratosService.GetSession(ctx, accessToken)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_INVALID_SESSION",
			Message: "Invalid session",
			Details: []interface{}{err.Error()},
		}
	}

	// Check if session is still valid
	if session.ExpiresAt.Before(time.Now()) {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_SESSION_EXPIRED",
			Message: "Session expired",
			Details: []interface{}{"Session has expired"},
		}
	}

	// Extract user information from Kratos session
	var userDTO dto.IdentityUserDTO
	if traits, ok := session.Identity.Traits.(map[string]interface{}); ok {
		userDTO = dto.IdentityUserDTO{
			ID:       session.Identity.Id,
			UserName: extractStringFromTraits(traits, "username", ""),
			Email:    extractStringFromTraits(traits, "email", ""),
			Phone:    extractStringFromTraits(traits, "phone", ""),
		}
	} else {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_SESSION",
			Message: "Invalid session data from Kratos",
			Details: []interface{}{"Unable to extract user traits"},
		}
	}

	// Return current session information
	return &dto.IdentityUserAuthDTO{
		AccessToken:      session.Id,
		RefreshToken:     "", // Kratos handles refresh internally
		AccessExpiresAt:  session.ExpiresAt.Unix(),
		RefreshExpiresAt: 0,
		LastLoginAt:      session.IssuedAt.Unix(),
		User:             userDTO,
	}, nil
}

func (u *userUseCase) Profile(
	ctx context.Context,
) (*dto.IdentityUserDTO, *dto.ErrorDTOResponse) {
	// Extract session token from context
	sessionToken, exists := ctx.Value("sessionToken").(string)
	if !exists {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_SESSION_TOKEN_MISSING",
			Message: "Session token missing",
			Details: []interface{}{
				map[string]string{
					"field": "session",
					"error": "Session token not found in context",
				},
			},
		}
	}

	// Get session using whoami endpoint
	session, err := u.kratosService.WhoAmI(ctx, sessionToken)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusUnauthorized,
			Code:    "MSG_INVALID_SESSION",
			Message: "Invalid session",
			Details: []interface{}{err.Error()},
		}
	}

	// Extract user information from Kratos session
	userDTO, err := extractUserFromTraits(session.Identity.Traits, "", "")
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_SESSION",
			Message: "Invalid session data",
			Details: []interface{}{err.Error()},
		}
	}
	userDTO.ID = session.Identity.Id
	return &userDTO, nil
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
		log.Printf("Failed to marshal traits: %v", err)
		return make(map[string]interface{}), false
	}

	var traitsMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &traitsMap); err != nil {
		log.Printf("Failed to unmarshal traits: %v", err)
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
