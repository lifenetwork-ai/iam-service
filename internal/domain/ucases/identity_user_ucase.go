package ucases

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/conf"
	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	ucase_interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
)

type userUseCase struct {
	userRepo             repositories.IdentityUserRepository
	sessionRepo          repositories.AccessSessionRepository
	challengeSessionRepo repositories.ChallengeSessionRepository
	emailService         services.EmailService
	smsService           services.SMSService
	jwtService           services.JWTService
}

func NewIdentityUserUseCase(
	userRepo repositories.IdentityUserRepository,
	sessionRepo repositories.AccessSessionRepository,
	challengeSessionRepo repositories.ChallengeSessionRepository,
	emailService services.EmailService,
	smsService services.SMSService,
	jwtService services.JWTService,
) ucase_interfaces.IdentityUserUseCase {
	return &userUseCase{
		userRepo:             userRepo,
		sessionRepo:          sessionRepo,
		challengeSessionRepo: challengeSessionRepo,
		emailService:         emailService,
		smsService:           smsService,
		jwtService:           jwtService,
	}
}

func (u *userUseCase) ChallengeWithPhone(
	ctx context.Context,
	phone string,
) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse) {
	if utils.IsPhoneNumber(phone) {
		return nil, &dto.ErrorDTOResponse{
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

	return nil, nil
}

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

	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Internal server error",
			Details: []any{err.Error()},
		}
	}

	if user == nil {
		user = &entities.IdentityUser{
			UserName: strings.TrimSpace(email),
			Email:    strings.TrimSpace(email),
		}

		if err := u.userRepo.Create(ctx, user); err != nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Internal server error",
				Details: []interface{}{err.Error()},
			}
		}
	}

	// Send email with OTP
	otp := utils.GenerateOTP()
	if err := u.emailService.SendOTP(ctx, email, otp); err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_SENDING_OTP_FAILED",
			Message: "Sending OTP to email failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Create challenge session
	sessionID := uuid.New().String()
	err = u.challengeSessionRepo.SaveChallenge(ctx, sessionID, &entities.ChallengeSession{
		Type:  "email",
		Email: email,
		OTP:   otp,
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
		SessionID:   sessionID,
		Receiver:    email,
		ChallengeAt: time.Now().Unix(),
	}, nil
}

func (u *userUseCase) ChallengeVerify(
	ctx context.Context,
	sessionID string,
	code string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	// cacheKey := &cachingTypes.Keyer{Raw: sessionID}
	// var sessionValue challengeSession
	// err := u.cacheRepo.RetrieveItem(cacheKey, &sessionValue)
	// if err != nil {
	// 	return nil, &dto.ErrorDTOResponse{
	// 		Status:  http.StatusNotFound,
	// 		Code:    "MSG_SESSION_NOT_FOUND",
	// 		Message: "Session not found",
	// 		Details: []interface{}{err.Error()},
	// 	}
	// }

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

	if sessionValue.OTP == "" {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_CODE",
			Message: "Invalid CODE",
			Details: []interface{}{
				map[string]string{"field": "otp", "error": "Missing OTP in session"},
			},
		}
	}

	if !conf.IsDebugMode() && sessionValue.OTP != code {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_INVALID_CODE",
			Message: "Invalid CODE",
			Details: []interface{}{
				map[string]string{"field": "code", "error": "Invalid OTP code"},
			},
		}
	}

	var challengeUser *entities.IdentityUser
	switch sessionValue.Type {
	case "email":
		if sessionValue.Email == "" {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "MSG_INVALID_CODE",
				Message: "Invalid CODE",
				Details: []interface{}{
					map[string]string{"field": "email", "error": "Missing email in session"},
				},
			}
		}

		user, err := u.userRepo.FindByEmail(ctx, sessionValue.Email)
		if err != nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Database error",
				Details: []interface{}{err.Error()},
			}
		}

		if user == nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusNotFound,
				Code:    "USER_NOT_FOUND",
				Message: "User not found",
				Details: []interface{}{},
			}
		}

		challengeUser = user

	case "phone":
		if sessionValue.Phone == "" {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "MSG_INVALID_CODE",
				Message: "Invalid CODE",
				Details: []interface{}{
					map[string]string{"field": "phone", "error": "Missing phone in session"},
				},
			}
		}

		user, err := u.userRepo.FindByPhone(ctx, sessionValue.Phone)
		if err != nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Database error",
				Details: []interface{}{err.Error()},
			}
		}

		if user == nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusNotFound,
				Code:    "USER_NOT_FOUND",
				Message: "User not found",
				Details: []interface{}{},
			}
		}
		challengeUser = user

	default:
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_TYPE",
			Message: "Invalid challenge type",
			Details: []interface{}{
				map[string]string{"field": "type", "error": "Unsupported challenge type"},
			},
		}
	}

	if challengeUser == nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_CODE",
			Message: "Invalid CODE",
			Details: []interface{}{map[string]string{
				"field": "user", "error": "User not found",
			}},
		}
	}

	// orgIDValue := ctx.Value("organizationId")
	// orgID, ok := orgIDValue.(string)
	// if !ok || orgID == "" {
	// 	return nil, &dto.ErrorDTOResponse{
	// 		Status:  http.StatusInternalServerError,
	// 		Code:    "MSG_ORGANIZATION_NOT_FOUND",
	// 		Message: "Organization not found",
	// 		Details: []interface{}{
	// 			map[string]string{"field": "organizationId", "error": "Organization not found"},
	// 		},
	// 	}
	// }

	// Generate JWT token
	jwtClaims := services.JWTClaims{
		OrganizationId: ctx.Value("organizationId").(string),
		UserId:         challengeUser.ID,
		UserName:       challengeUser.Name,
	}

	jwtToken, err := u.jwtService.GenerateToken(ctx, jwtClaims)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_GENERATE_TOKEN_FAILED",
			Message: "Generate token failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Save JWT token to database
	session := &entities.AccessSession{
		OrganizationId:   jwtClaims.OrganizationId,
		UserId:           jwtClaims.UserId,
		AccessToken:      jwtToken.AccessToken,
		RefreshToken:     jwtToken.RefreshToken,
		AccessExpiredAt:  time.Unix(jwtToken.AccessTokenExpiry, 0),
		RefreshExpiredAt: time.Unix(jwtToken.RefreshTokenExpiry, 0),
		LastRevokedAt:    jwtToken.ClaimAt,
	}
	_, err = u.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_SAVE_SESSION_FAILED",
			Message: "Save session failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Return JWT token
	return &dto.IdentityUserAuthDTO{
		AccessToken:      jwtToken.AccessToken,
		RefreshToken:     jwtToken.RefreshToken,
		AccessExpiresAt:  jwtToken.AccessTokenExpiry,
		RefreshExpiresAt: jwtToken.RefreshTokenExpiry,
		LastLoginAt:      jwtToken.ClaimAt.Unix(),
		User:             challengeUser.ToDTO(),
	}, nil
}

func (u *userUseCase) LogInWithGoogle(
	ctx context.Context,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithFacebook(
	ctx context.Context,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithApple(
	ctx context.Context,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) Register(
	ctx context.Context,
	payload dto.IdentityUserRegisterDTO,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogIn(
	ctx context.Context,
	username string,
	password string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogOut(
	ctx context.Context,
) *dto.ErrorDTOResponse {
	return nil
}

func (u *userUseCase) RefreshToken(
	ctx context.Context,
	accessToken string,
	refreshToken string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	session, err := u.sessionRepo.FindByAccessToken(ctx, accessToken)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_INVALID_ACCESS_TOKEN",
			Message: "Invalid access token",
			Details: []interface{}{err.Error()},
		}
	}

	if session == nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "MSG_SESSION_NOT_FOUND",
			Message: "Session not found",
			Details: []interface{}{map[string]string{
				"field": "session", "error": "Session not found",
			}},
		}
	}

	refreshTokenHash := utils.HashToken(refreshToken)
	if session.RefreshToken != refreshTokenHash {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_REFRESH_TOKEN_MISSMATCH",
			Message: "Refresh token missmatch",
			Details: []interface{}{map[string]string{
				"field": "refresh_token", "error": "Refresh token missmatch",
			}},
		}
	}

	// Generate new JWT token
	jwtClaims, err := u.jwtService.ValidateToken(ctx, accessToken)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_INVALID_ACCESS_TOKEN",
			Message: "Invalid access token",
			Details: []interface{}{err.Error()},
		}
	}

	jwtToken, err := u.jwtService.GenerateToken(ctx, *jwtClaims)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_GENERATE_TOKEN_FAILED",
			Message: "Generate token failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Save JWT token to database
	session.AccessToken = jwtToken.AccessToken
	session.RefreshToken = jwtToken.RefreshToken
	session.AccessExpiredAt = time.Unix(jwtToken.AccessTokenExpiry, 0)
	session.RefreshExpiredAt = time.Unix(jwtToken.RefreshTokenExpiry, 0)
	session.LastRevokedAt = jwtToken.ClaimAt
	_, err = u.sessionRepo.Update(ctx, session)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_SAVE_SESSION_FAILED",
			Message: "Save session failed",
			Details: []interface{}{err.Error()},
		}
	}

	requester, err := u.userRepo.FindByID(ctx, jwtClaims.UserId)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Internal server error",
			Details: []interface{}{err.Error()},
		}
	}

	// Return JWT token
	return &dto.IdentityUserAuthDTO{
		AccessToken:      jwtToken.AccessToken,
		RefreshToken:     jwtToken.RefreshToken,
		AccessExpiresAt:  jwtToken.AccessTokenExpiry,
		RefreshExpiresAt: jwtToken.RefreshTokenExpiry,
		LastLoginAt:      jwtToken.ClaimAt.Unix(),
		User:             requester.ToDTO(),
	}, nil
}

func (u *userUseCase) Profile(
	ctx context.Context,
) (*dto.IdentityUserDTO, *dto.ErrorDTOResponse) {
	requesterId, exists := ctx.Value("requesterId").(string)
	if !exists {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_REQUESTER_NOT_FOUND",
			Message: "Requester not found",
			Details: []interface{}{
				map[string]string{
					"field": "requester",
					"error": "Requester not found",
				},
			},
		}
	}

	requester, err := u.userRepo.FindByID(ctx, requesterId)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Internal server error",
			Details: []interface{}{err.Error()},
		}
	}

	dto := requester.ToDTO()
	return &dto, nil
}
