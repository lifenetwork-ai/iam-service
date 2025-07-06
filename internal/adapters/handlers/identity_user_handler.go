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
