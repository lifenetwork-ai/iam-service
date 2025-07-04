package handlers

import (
	"net/http"

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

// Verify the challenge or registration
// @Summary Verify the challenge or registration
// @Description Verify either a login challenge or registration flow
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param challenge body dto.IdentityChallengeVerifyDTO true "verification payload, type can be registration or login"
// @Success 200 {object} response.SuccessResponse "Successful verification"
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

	var auth *dto.IdentityUserAuthDTO
	var err *dto.ErrorDTOResponse

	switch reqPayload.Type {
	case "challenge":
		auth, err = h.ucase.ChallengeVerify(ctx, reqPayload.FlowID, reqPayload.Code)
	case "registration":
		auth, err = h.ucase.VerifyRegister(ctx, reqPayload.FlowID, reqPayload.Code)
	case "login":
		auth, err = h.ucase.VerifyLogin(ctx, reqPayload.FlowID, reqPayload.Code)
	default:
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_VERIFICATION_TYPE",
			"Invalid verification type",
			nil,
		)
		return
	}

	if err != nil {
		httpresponse.Error(
			ctx,
			err.Status,
			err.Code,
			err.Message,
			err.Details,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, auth)
}

// // Login to authenticate user.
// // @Summary Authenticate user
// // @Description Authenticate user
// // @Tags users
// // @Accept json
// // @Produce json
// // @Param X-Organization-Id header string true "Organization ID"
// // @Param login body dto.IdentityUserLoginDTO true "login payload"
// // @Success 200 {object} response.SuccessResponse "Successful authenticate user"
// // @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// // @Failure 500 {object} response.ErrorResponse "Internal server error"
// // @Router /api/v1/users/login [post]
// func (h *userHandler) Login(ctx *gin.Context) {
// 	httpresponse.Error(
// 		ctx,
// 		http.StatusNotImplemented,
// 		"MSG_NOT_IMPLEMENTED",
// 		"Not implemented",
// 		nil,
// 	)
// }

// // Login with Google to authenticate user.
// // @Summary Authenticate user with Google
// // @Description Authenticate user with Google
// // @Tags users
// // @Accept json
// // @Produce json
// // @Param X-Organization-Id header string true "Organization ID"
// // @Success 200 {object} response.SuccessResponse "Successful authenticate user with Google"
// // @Failure 500 {object} response.ErrorResponse "Internal server error"
// // @Router /api/v1/users/login-with-google [post]
// func (h *userHandler) LoginWithGoogle(ctx *gin.Context) {
// 	httpresponse.Error(
// 		ctx,
// 		http.StatusNotImplemented,
// 		"MSG_NOT_IMPLEMENTED",
// 		"Not implemented",
// 		nil,
// 	)
// }

// // Login with Facebook to authenticate user.
// // @Summary Authenticate user with Facebook
// // @Description Authenticate user with Facebook
// // @Tags users
// // @Accept json
// // @Produce json
// // @Param X-Organization-Id header string true "Organization ID"
// // @Success 200 {object} response.SuccessResponse "Successful authenticate user with Facebook"
// // @Failure 500 {object} response.ErrorResponse "Internal server error"
// // @Router /api/v1/users/login-with-facebook [post]
// func (h *userHandler) LoginWithFacebook(ctx *gin.Context) {
// 	httpresponse.Error(
// 		ctx,
// 		http.StatusNotImplemented,
// 		"MSG_NOT_IMPLEMENTED",
// 		"Not implemented",
// 		nil,
// 	)
// }

// // Login with Apple to authenticate user.
// // @Summary Authenticate user with Apple
// // @Description Authenticate user with Apple
// // @Tags users
// // @Accept json
// // @Produce json
// // @Param X-Organization-Id header string true "Organization ID"
// // @Success 200 {object} response.SuccessResponse "Successful authenticate user with Apple"
// // @Failure 500 {object} response.ErrorResponse "Internal server error"
// // @Router /api/v1/users/login-with-apple [post]
// func (h *userHandler) LoginWithApple(ctx *gin.Context) {
// 	httpresponse.Error(
// 		ctx,
// 		http.StatusNotImplemented,
// 		"MSG_NOT_IMPLEMENTED",
// 		"Not implemented",
// 		nil,
// 	)
// }

// // Refresh to refresh token.
// // @Summary Refresh token
// // @Description Refresh token
// // @Tags users
// // @Accept json
// // @Produce json
// // @Param X-Organization-Id header string true "Organization ID"
// // @Param Authorization header string true "Bearer Token"
// // @Param refresh_token body dto.IdentityRefreshTokenDTO true "refresh token payload"
// // @Success 200 {object} response.SuccessResponse "Successful refresh token"
// // @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// // @Failure 500 {object} response.ErrorResponse "Internal server error"
// // @Router /api/v1/users/refresh-token [post]
// func (h *userHandler) RefreshToken(ctx *gin.Context) {
// 	// Get the Bearer token from the Authorization header
// 	authHeader := ctx.GetHeader("Authorization")
// 	if authHeader == "" {
// 		httpresponse.Error(
// 			ctx,
// 			http.StatusUnauthorized,
// 			"UNAUTHORIZED",
// 			"Authorization header is required",
// 			[]map[string]string{{
// 				"field": "Authorization",
// 				"error": "Authorization header is required",
// 			}},
// 		)
// 		return
// 	}

// 	// Check if the token is a Bearer token
// 	tokenParts := strings.Split(authHeader, " ")
// 	tokenPrefix := tokenParts[0]
// 	if len(tokenParts) != 2 || (tokenPrefix != "Bearer" && tokenPrefix != "Token") {
// 		httpresponse.Error(
// 			ctx,
// 			http.StatusUnauthorized,
// 			"UNAUTHORIZED",
// 			"Authorization header is required",
// 			[]map[string]string{{
// 				"field": "Authorization",
// 				"error": "Invalid authorization header format",
// 			}},
// 		)
// 		return
// 	}

// 	payload := dto.IdentityRefreshTokenDTO{}
// 	if err := ctx.ShouldBindJSON(&payload); err != nil {
// 		httpresponse.Error(
// 			ctx,
// 			http.StatusBadRequest,
// 			"MSG_INVALID_PAYLOAD",
// 			"Invalid payload",
// 			err,
// 		)
// 		return
// 	}

// 	if payload.RefreshToken == "" {
// 		httpresponse.Error(
// 			ctx,
// 			http.StatusBadRequest,
// 			"MSG_REFRESH_TOKEN_IS_REQUIRED",
// 			"Refresh token is required",
// 			nil,
// 		)
// 		return
// 	}

// 	auth, err := h.ucase.RefreshToken(ctx, tokenParts[1], payload.RefreshToken)
// 	if err != nil {
// 		httpresponse.Error(
// 			ctx,
// 			http.StatusInternalServerError,
// 			"MSG_FAILED_TO_REFRESH_TOKEN",
// 			"Failed to refresh token",
// 			err,
// 		)
// 		return
// 	}

// 	httpresponse.Success(ctx, http.StatusOK, auth)
// }

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
// @Param request body object true "Empty request body"
// @Success 200 {object} response.SuccessResponse{data=interface{}} "Successful de-authenticate user"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/logout [post]
func (h *userHandler) Logout(ctx *gin.Context) {
	if err := h.ucase.LogOut(ctx); err != nil {
		httpresponse.Error(
			ctx,
			err.Status,
			err.Code,
			err.Message,
			err.Details,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, nil)
}

// Register to register user.
// @Summary Register a new user
// @Description Register a new user
// @Tags users
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param register body dto.IdentityUserRegisterDTO true "Only email or phone must be provided, tenant is required, if both are provided, email will be used"
// @Success 200 {object} response.SuccessResponse{data=dto.IdentityUserAuthDTO} "Successful user registration with verification flow"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/register [post]
func (h *userHandler) Register(ctx *gin.Context) {
	var reqPayload dto.IdentityUserRegisterDTO
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

	// Validate that at least one contact method is provided
	if reqPayload.Email == "" && reqPayload.Phone == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_CONTACT_METHOD_REQUIRED",
			"Either email or phone must be provided",
			[]map[string]string{
				{
					"field": "email,phone",
					"error": "At least one contact method (email or phone) must be provided",
				},
			},
		)
		return
	}

	// Validate tenant is provided
	if reqPayload.Tenant == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_TENANT_REQUIRED",
			"Tenant is required",
			[]map[string]string{
				{
					"field": "tenant",
					"error": "Tenant must be provided",
				},
			},
		)
		return
	}

	// Call the use case to handle registration
	auth, err := h.ucase.Register(ctx, reqPayload)
	if err != nil {
		httpresponse.Error(
			ctx,
			err.Status,
			err.Code,
			err.Message,
			err.Details,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, auth)
}
