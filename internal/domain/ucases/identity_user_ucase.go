package ucases

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/genefriendway/human-network-iam/conf"
	cachingTypes "github.com/genefriendway/human-network-iam/infrastructures/caching/types"
	infra_interfaces "github.com/genefriendway/human-network-iam/infrastructures/interfaces"
	repositories "github.com/genefriendway/human-network-iam/internal/adapters/repositories/types"
	"github.com/genefriendway/human-network-iam/internal/adapters/services"
	dto "github.com/genefriendway/human-network-iam/internal/delivery/dto"
	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
	ucase_interfaces "github.com/genefriendway/human-network-iam/internal/domain/ucases/types"
	"github.com/genefriendway/human-network-iam/packages/utils"
)

type userUseCase struct {
	userRepo     repositories.IdentityUserRepository
	sessionRepo  repositories.AccessSessionRepository
	cacheRepo    infra_interfaces.CacheRepository
	emailService services.EmailService
	smsService   services.SMSService
	jwtService   services.JWTService
}

func NewIdentityUserUseCase(
	userRepo repositories.IdentityUserRepository,
	sessionRepo repositories.AccessSessionRepository,
	cacheRepo infra_interfaces.CacheRepository,
	emailService services.EmailService,
	smsService services.SMSService,
	jwtService services.JWTService,
) ucase_interfaces.IdentityUserUseCase {
	return &userUseCase{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		cacheRepo:    cacheRepo,
		emailService: emailService,
		smsService:   smsService,
		jwtService:   jwtService,
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
			Details: []interface{}{
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
			Details: []interface{}{err.Error()},
		}
	}

	if user == nil {
		// Create user with email
		user = &entities.IdentityUser{
			UserName: strings.TrimSpace(email),
			Email:    strings.TrimSpace(email),
		}

		err = u.userRepo.Create(ctx, user)
		if err != nil {
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
	err = u.emailService.SendOTP(ctx, email, otp)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_SENDING_OTP_FAILED",
			Message: "Sending OTP to email failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Create challenge session
	session := uuid.New().String()

	// Save challenge session to cache for 5 minutes
	cacheKey := &cachingTypes.Keyer{Raw: session}
	cacheValue := map[string]string{
		"type":  "email",
		"email": email,
		"otp":   otp,
	}
	err = u.cacheRepo.SaveItem(cacheKey, cacheValue, 5*time.Minute)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_CACHING_FAILED",
			Message: "Caching failed",
			Details: []interface{}{err.Error()},
		}
	}

	// Return challenge session
	return &dto.IdentityUserChallengeDTO{
		SessionID:   session,
		Receiver:    email,
		ChallengeAt: time.Now().Unix(),
	}, nil
}

func (u *userUseCase) ChallengeVerify(
	ctx context.Context,
	sessionID string,
	code string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	cacheKey := &cachingTypes.Keyer{Raw: sessionID}
	var cacheValue interface{}
	err := u.cacheRepo.RetrieveItem(cacheKey, &cacheValue)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "MSG_SESSION_NOT_FOUND",
			Message: "Session not found",
			Details: []interface{}{err.Error()},
		}
	}

	sessionValue, ok := cacheValue.(map[string]string)
	if !ok {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_CODE",
			Message: "Invalid CODE",
			Details: []interface{}{map[string]string{
				"field": "session", "error": "Invalid session",
			}},
		}
	}

	otp, exists := sessionValue["otp"]
	if !exists {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_CODE",
			Message: "Invalid CODE",
			Details: []interface{}{map[string]string{
				"field": "otp", "error": "Invalid OTP",
			}},
		}
	}

	// Ignore OTP check in debug mode
	if !conf.IsDebugMode() {
		if otp != code {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Code:    "MSG_INVALID_CODE",
				Message: "Invalid CODE",
				Details: []interface{}{map[string]string{
					"field": "code", "error": "Invalid code",
				}},
			}
		}
	}

	challenge_type, exists := sessionValue["type"]
	if !exists || (challenge_type != "email" && challenge_type != "phone") {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_INVALID_CODE",
			Message: "Invalid CODE",
			Details: []interface{}{map[string]string{
				"field": "type", "error": "Invalid type",
			}},
		}
	}

	var challengeUser *entities.IdentityUser
	if challenge_type == "email" {
		email, exists := sessionValue["email"]
		if !exists {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "MSG_INVALID_CODE",
				Message: "Invalid CODE",
				Details: []interface{}{map[string]string{
					"field": "email", "error": "Not found email in session",
				}},
			}
		}

		user, err := u.userRepo.FindByEmail(ctx, email)
		if err != nil || user == nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Internal server error",
				Details: []interface{}{err.Error()},
			}
		}

		challengeUser = user
	}

	if challenge_type == "phone" {
		phone, exists := sessionValue["phone"]
		if !exists {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "MSG_INVALID_CODE",
				Message: "Invalid CODE",
				Details: []interface{}{map[string]string{
					"field": "phone", "error": "Not found phone in session",
				}},
			}
		}

		user, err := u.userRepo.FindByPhone(ctx, phone)
		if err != nil || user == nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Internal server error",
				Details: []interface{}{err.Error()},
			}
		}

		challengeUser = user
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
