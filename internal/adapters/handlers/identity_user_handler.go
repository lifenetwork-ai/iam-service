package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/constants"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
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
// @Param X-Tenant-Id header string true "Tenant ID"
// @Tags users
// @Accept json
// @Produce json
// @Param challenge body dto.IdentityChallengeWithPhoneDTO true "challenge payload"
// @Success 200 {object} response.SuccessResponse "Successful make a challenge with Phone and OTP"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 429 {object} response.ErrorResponse "Too many attempts, rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/challenge-with-phone [post]
func (h *userHandler) ChallengeWithPhone(ctx *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT",
			"Invalid tenant",
			err,
		)
		return
	}

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

	challenge, usecaseErr := h.ucase.ChallengeWithPhone(ctx, tenant.ID, reqPayload.Phone)
	if usecaseErr != nil {
		handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, challenge)
}

// ChallengeWithEmail to login with email and otp.
// @Summary Login with email and otp
// @Description Login with email and otp
// @Param X-Tenant-Id header string true "Tenant ID"
// @Tags users
// @Accept json
// @Produce json
// @Param challenge body dto.IdentityChallengeWithEmailDTO true "challenge payload"
// @Success 200 {object} response.SuccessResponse "Successful make a challenge with Email and OTP"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 429 {object} response.ErrorResponse "Too many attempts, rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/challenge-with-email [post]
func (h *userHandler) ChallengeWithEmail(ctx *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT",
			"Invalid tenant",
			err,
		)
		return
	}

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

	challenge, usecaseErr := h.ucase.ChallengeWithEmail(ctx.Request.Context(), tenant.ID, reqPayload.Email)
	if usecaseErr != nil {
		handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, challenge)
}

// Verify the challenge or registration
// @Summary Verify the challenge or registration
// @Description Verify either a login challenge or registration flow
// @Param X-Tenant-Id header string true "Tenant ID"
// Verify a login or registration challenge
// @Summary Verify login or registration challenge
// @Description Verify a one-time code sent to user for either login or registration challenge.
// @Tags users
// @Accept json
// @Produce json
// @Param challenge body dto.IdentityChallengeVerifyDTO true "Verification payload. `type` must be one of: `register`, `login`"
// @Success 200 {object} response.SuccessResponse "Verification successful"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or code"
// @Failure 429 {object} response.ErrorResponse "Too many attempts, rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/challenge-verify [post]
func (h *userHandler) ChallengeVerify(ctx *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get tenant: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT",
			"Invalid tenant",
			err,
		)
		return
	}

	var reqPayload dto.IdentityChallengeVerifyDTO
	if err = ctx.ShouldBindJSON(&reqPayload); err != nil {
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

	var auth *types.IdentityUserAuthResponse
	var usecaseErr *domainerrors.DomainError

	switch reqPayload.Type {
	case constants.FlowTypeRegister.String():
		auth, usecaseErr = h.ucase.VerifyRegister(ctx.Request.Context(), tenant.ID, reqPayload.FlowID, reqPayload.Code)
	case constants.FlowTypeLogin.String():
		auth, usecaseErr = h.ucase.VerifyLogin(ctx.Request.Context(), tenant.ID, reqPayload.FlowID, reqPayload.Code)
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

	if usecaseErr != nil {
		handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, auth)
}

// Me to get user profile.
// @Summary Get user profile
// @Description Get user profile
// @Param X-Tenant-Id header string true "Tenant ID"
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token (Bearer ory...)" default(Bearer <token>)
// @Success 200 {object} response.SuccessResponse "Successful get user profile"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/me [get]
func (h *userHandler) Me(ctx *gin.Context) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		httpresponse.Error(
			ctx,
			http.StatusUnauthorized,
			"MSG_UNAUTHORIZED",
			"Unauthorized",
			[]interface{}{
				map[string]string{"field": "user", "error": "User not found"},
			},
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, *user)
}

// Logout to de-authenticate user.
// @Summary De-authenticate user
// @Description De-authenticate user
// @Param X-Tenant-Id header string true "Tenant ID"
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token (Bearer ory...)" default(Bearer <token>)
// @Param request body object true "Empty request body"
// @Success 200 {object} response.SuccessResponse{data=interface{}} "Successful de-authenticate user"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 403 {object} response.ErrorResponse "Forbidden - Invalid or missing token"
// @Failure 429 {object} response.ErrorResponse "Too many attempts, rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/logout [post]
func (h *userHandler) Logout(ctx *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT",
			"Invalid tenant",
			err,
		)
		return
	}

	// Get session token from gin context and create new context with it
	sessionToken, exists := ctx.Get(string(constants.SessionTokenKey))
	if !exists {
		httpresponse.Error(
			ctx,
			http.StatusUnauthorized,
			"MSG_UNAUTHORIZED",
			"Unauthorized",
			[]interface{}{
				map[string]string{"field": "session_token", "error": "Session token not found"},
			},
		)
		return
	}

	reqCtx := context.WithValue(ctx.Request.Context(), constants.SessionTokenKey, sessionToken)
	usecaseErr := h.ucase.Logout(reqCtx, tenant.ID)
	if usecaseErr != nil {
		handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, nil)
}

// Register to register user.
// @Summary Register a new user
// @Description Register a new user
// @Param X-Tenant-Id header string true "Tenant ID"
// @Tags users
// @Accept json
// @Produce json
// @Param register body dto.IdentityUserRegisterDTO true "Only email or phone must be provided, if both are provided then error will be returned"
// @Success 200 {object} response.SuccessResponse{data=types.IdentityUserAuthResponse} "Successful user registration with verification flow"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 409 {object} response.ErrorResponse "Email or phone number already exists"
// @Failure 429 {object} response.ErrorResponse "Too many attempts, rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/register [post]
func (h *userHandler) Register(ctx *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT",
			"Invalid tenant",
			err,
		)
		return
	}

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

	if errResponse := validateRegisterPayload(reqPayload); errResponse != nil {
		httpresponse.Error(
			ctx,
			errResponse.Status,
			errResponse.Code,
			errResponse.Message,
			errResponse.Details,
		)
		return
	}

	auth, usecaseErr := h.ucase.Register(ctx.Request.Context(), tenant.ID, reqPayload.Email, reqPayload.Phone)
	if usecaseErr != nil {
		handleDomainError(ctx, usecaseErr)
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

	return nil
}

// AddIdentifier to bind additional login identifier to current user.
// @Summary Add new identifier (email or phone)
// @Description Add a verified identifier (email or phone) to current user
// @Tags users
// @Accept json
// @Produce json
// @Param X-Tenant-Id header string true "Tenant ID"
// @Param Authorization header string true "Bearer Token (Bearer ory...)"
// @Param body body dto.IdentityUserAddIdentifierDTO true "Identifier info"
// @Success 200 {object} response.SuccessResponse{data=types.IdentityUserChallengeResponse} "OTP sent for verification"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 409 {object} response.ErrorResponse "Identifier or type already exists"
// @Failure 429 {object} response.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/me/add-identifier [post]
func (h *userHandler) AddIdentifier(ctx *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_TENANT", "Invalid tenant", err)
		return
	}

	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		httpresponse.Error(ctx, http.StatusUnauthorized, "MSG_UNAUTHORIZED", "Unauthorized", nil)
		return
	}

	var req dto.IdentityUserAddIdentifierDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_PAYLOAD", "Invalid payload", err)
		return
	}

	// Validate identifier type
	identifierType, err := utils.GetIdentifierType(req.NewIdentifier)
	if err != nil {
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_IDENTIFIER_TYPE", "Invalid identifier type", err)
		return
	}

	result, usecaseErr := h.ucase.AddNewIdentifier(ctx, tenant.ID, user.GlobalUserID, req.NewIdentifier, identifierType)
	if usecaseErr != nil {
		handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, result)
}
