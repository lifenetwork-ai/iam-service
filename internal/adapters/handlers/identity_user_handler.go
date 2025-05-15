package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type userHandler struct {
	ucase interfaces.IdentityUserUseCase
}

func NewIdentityUserHandler(ucase interfaces.IdentityUserUseCase) *userHandler {
	return &userHandler{
		ucase: ucase,
	}
}

// ChallengeWithPhone to login with phone and otp.
// @Summary Login with phone and otp
// @Description Login with phone and otp
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param challenge body dto.IdentityChallengeWithPhoneDTO true "challenge payload"
// @Success 200 {object} response.SuccessResponse "Successful make a challenge with Phone and OTP"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/challenge-with-phone [post]
func (h *userHandler) ChallengeWithPhone(ctx *gin.Context) {
	reqPayload := dto.IdentityChallengeWithPhoneDTO{}
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid payload",
			err,
		)
		return
	}

	if reqPayload.Phone == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_PHONE_NUMBER_IS_REQUIRED",
			"Phone number is required",
			nil,
		)
		return
	}

	session, err := h.ucase.ChallengeWithPhone(ctx, reqPayload.Phone)
	if err != nil {
		httpresponse.Error(
			ctx,
			http.StatusInternalServerError,
			"MSG_FAILED_TO_MAKE_CHALLENGE",
			"Failed to make a challenge",
			err,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, session)
}

// ChallengeWithEmail to login with email and otp.
// @Summary Login with email and otp
// @Description Login with email and otp
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param challenge body dto.IdentityChallengeWithEmailDTO true "challenge payload"
// @Success 200 {object} response.SuccessResponse "Successful make a challenge with Email and OTP"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/challenge-with-email [post]
func (h *userHandler) ChallengeWithEmail(ctx *gin.Context) {
	var reqPayload dto.IdentityChallengeWithEmailDTO
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid payload",
			err,
		)
		return
	}

	if reqPayload.Email == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_EMAIL_IS_REQUIRED",
			"Email is required",
			nil,
		)
		return
	}

	session, err := h.ucase.ChallengeWithEmail(ctx, reqPayload.Email)
	if err != nil {
		httpresponse.Error(
			ctx,
			http.StatusInternalServerError,
			"MSG_FAILED_TO_MAKE_CHALLENGE",
			"Failed to make a challenge",
			err,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, session)
}

// Verify the challenge
// @Summary Verify the challenge
// @Description Verify the challenge
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param challenge body dto.IdentityChallengeVerifyDTO true "challenge payload"
// @Success 200 {object} response.SuccessResponse "Successful verify the challenge"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/challenge-verify [post]
func (h *userHandler) ChallengeVerify(ctx *gin.Context) {
	var reqPayload dto.IdentityChallengeVerifyDTO
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid payload",
			err,
		)
		return
	}

	if reqPayload.SessionID == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_SESSION_ID_IS_REQUIRED",
			"Session ID is required",
			nil,
		)
		return
	}

	if reqPayload.Code == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_CHALLENGE_CODE_IS_REQUIRED",
			"Challenge code is required",
			nil,
		)
		return
	}

	auth, err := h.ucase.ChallengeVerify(ctx, reqPayload.SessionID, reqPayload.Code)
	if err != nil {
		httpresponse.Error(
			ctx,
			http.StatusInternalServerError,
			"MSG_FAILED_TO_VERIFY_CHALLENGE",
			"Failed to verify the challenge",
			err,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, auth)
}

// Login to authenticate user.
// @Summary Authenticate user
// @Description Authenticate user
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param login body dto.IdentityUserLoginDTO true "login payload"
// @Success 200 {object} response.SuccessResponse "Successful authenticate user"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/login [post]
func (h *userHandler) Login(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Login with Google to authenticate user.
// @Summary Authenticate user with Google
// @Description Authenticate user with Google
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Success 200 {object} response.SuccessResponse "Successful authenticate user with Google"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/login-with-google [post]
func (h *userHandler) LoginWithGoogle(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Login with Facebook to authenticate user.
// @Summary Authenticate user with Facebook
// @Description Authenticate user with Facebook
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Success 200 {object} response.SuccessResponse "Successful authenticate user with Facebook"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/login-with-facebook [post]
func (h *userHandler) LoginWithFacebook(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Login with Apple to authenticate user.
// @Summary Authenticate user with Apple
// @Description Authenticate user with Apple
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Success 200 {object} response.SuccessResponse "Successful authenticate user with Apple"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/login-with-apple [post]
func (h *userHandler) LoginWithApple(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Refresh to refresh token.
// @Summary Refresh token
// @Description Refresh token
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param refresh_token body dto.IdentityRefreshTokenDTO true "refresh token payload"
// @Success 200 {object} response.SuccessResponse "Successful refresh token"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/refresh-token [post]
func (h *userHandler) RefreshToken(ctx *gin.Context) {
	// Get the Bearer token from the Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		httpresponse.Error(
			ctx,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Authorization header is required",
			[]map[string]string{{
				"field": "Authorization",
				"error": "Authorization header is required",
			}},
		)
		return
	}

	// Check if the token is a Bearer token
	tokenParts := strings.Split(authHeader, " ")
	tokenPrefix := tokenParts[0]
	if len(tokenParts) != 2 || (tokenPrefix != "Bearer" && tokenPrefix != "Token") {
		httpresponse.Error(
			ctx,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Authorization header is required",
			[]map[string]string{{
				"field": "Authorization",
				"error": "Invalid authorization header format",
			}},
		)
		return
	}

	payload := dto.IdentityRefreshTokenDTO{}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid payload",
			err,
		)
		return
	}

	if payload.RefreshToken == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_REFRESH_TOKEN_IS_REQUIRED",
			"Refresh token is required",
			nil,
		)
		return
	}

	auth, err := h.ucase.RefreshToken(ctx, tokenParts[1], payload.RefreshToken)
	if err != nil {
		httpresponse.Error(
			ctx,
			http.StatusInternalServerError,
			"MSG_FAILED_TO_REFRESH_TOKEN",
			"Failed to refresh token",
			err,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, auth)
}

// Me to get user profile.
// @Summary Get user profile
// @Description Get user profile
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} response.SuccessResponse "Successful get user profile"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/me [get]
func (h *userHandler) Me(ctx *gin.Context) {
	requester, err := h.ucase.Profile(ctx)
	if err != nil {
		httpresponse.Error(
			ctx,
			http.StatusInternalServerError,
			"MSG_FAILED_TO_GET_USER_PROFILE",
			"Failed to get user profile",
			err,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, *requester)
}

// Logout to de-authenticate user.
// @Summary De-authenticate user
// @Description De-authenticate user
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} response.SuccessResponse "Successful de-authenticate user"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/logout [post]
func (h *userHandler) Logout(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}
