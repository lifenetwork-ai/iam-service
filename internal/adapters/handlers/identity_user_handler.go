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
// @Param challenge body dto.IdentityChallengeWithPhoneDTO true "challenge payload"
// @Success 200 {object} response.SuccessResponse "Successful make a challenge with Phone and OTP"
// @Failure 400 {object} response.ErrorResponse "Invalid phone number format or missing phone"
// @Failure 401 {object} response.ErrorResponse "Login failed - invalid credentials"
// @Failure 404 {object} response.ErrorResponse "User not found - no user registered with this phone number"
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

	challenge, err := h.ucase.ChallengeWithPhone(ctx, reqPayload.Phone)
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

	httpresponse.Success(ctx, http.StatusOK, challenge)
}

// ChallengeWithEmail to login with email and otp.
// @Summary Login with email and otp
// @Description Login with email and otp
// @Tags users
// @Accept json
// @Produce json
// @Param challenge body dto.IdentityChallengeWithEmailDTO true "challenge payload"
// @Success 200 {object} response.SuccessResponse "Successful make a challenge with Email and OTP"
// @Failure 400 {object} response.ErrorResponse "Invalid email format or missing email"
// @Failure 401 {object} response.ErrorResponse "Login failed - invalid credentials"
// @Failure 404 {object} response.ErrorResponse "User not found - no user registered with this email address"
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

	challenge, err := h.ucase.ChallengeWithEmail(ctx, reqPayload.Email)
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

	httpresponse.Success(ctx, http.StatusOK, challenge)
}

// Verify the challenge or registration
// @Summary Verify the challenge or registration
// @Description Verify either a login challenge or registration flow
// @Tags users
// @Accept json
// @Produce json
// @Param challenge body dto.IdentityChallengeVerifyDTO true "verification payload, type can be registration or login"
// @Success 200 {object} response.SuccessResponse "Successful verification"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or verification code"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - verification failed"
// @Failure 404 {object} response.ErrorResponse "Challenge session not found"
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
	case "register":
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

// Me to get user profile.
// @Summary Get user profile
// @Description Get user profile
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token(Bearer ory...)" default(Bearer <token>)
// @Success 200 {object} response.SuccessResponse "Successful get user profile"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/me [get]
func (h *userHandler) Me(ctx *gin.Context) {
	requester, err := h.ucase.Profile(ctx)
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

	httpresponse.Success(ctx, http.StatusOK, *requester)
}

// Logout to de-authenticate user.
// @Summary De-authenticate user
// @Description De-authenticate user
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token(Bearer ory...)" default(Bearer <token>)
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
// @Param register body dto.IdentityUserRegisterDTO true "Only email or phone must be provided, if both are provided then error will be returned. Tenant field is required(available value: `genetica`,`life_ai`)"
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

	err := validateRegisterPayload(reqPayload)
	if err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			err.Status,
			err.Code,
			err.Message,
			err.Details,
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

func validateRegisterPayload(reqPayload dto.IdentityUserRegisterDTO) *dto.ErrorDTOResponse {
	if reqPayload.Email == "" && reqPayload.Phone == "" {
		return &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_CONTACT_METHOD_REQUIRED",
			Message: "Either email or phone must be provided",
		}
	}

	if reqPayload.Email != "" && reqPayload.Phone != "" {
		return &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_ONLY_EMAIL_OR_PHONE_MUST_BE_PROVIDED",
			Message: "Only email or phone must be provided",
		}
	}

	if reqPayload.Tenant == "" {
		return &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "MSG_TENANT_REQUIRED",
			Message: "Tenant is required",
		}
	}

	return nil
}
