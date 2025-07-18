package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
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

// handleDomainError is a centralized error handler for domain errors
func (h *userHandler) handleDomainError(ctx *gin.Context, err *domainerrors.DomainError) {
	switch err.Type {
	case domainerrors.ErrorTypeValidation:
		httpresponse.Error(ctx, http.StatusBadRequest, err.Code, err.Message, err.Details)
	case domainerrors.ErrorTypeNotFound:
		httpresponse.Error(ctx, http.StatusNotFound, err.Code, err.Message, err.Details)
	case domainerrors.ErrorTypeUnauthorized:
		httpresponse.Error(ctx, http.StatusUnauthorized, err.Code, err.Message, err.Details)
	case domainerrors.ErrorTypeConflict:
		httpresponse.Error(ctx, http.StatusConflict, err.Code, err.Message, err.Details)
	case domainerrors.ErrorTypeRateLimit:
		httpresponse.Error(ctx, http.StatusTooManyRequests, err.Code, err.Message, err.Details)
	case domainerrors.ErrorTypeInternal:
		// Log internal errors for debugging
		logger.GetLogger().Errorf("Internal error: %v", err.Error())
		httpresponse.Error(ctx, http.StatusInternalServerError, err.Code, err.Message, err.Details)
	default:
		// Fallback for unknown error types
		logger.GetLogger().Errorf("Unknown error type: %v", err.Error())
		httpresponse.Error(ctx, http.StatusInternalServerError, err.Code, err.Message, err.Details)
	}
}

// ChallengeWithPhone to login with phone and otp.
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
		h.handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, challenge)
}

// ChallengeWithEmail to login with email and otp.
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
		h.handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, challenge)
}

// ChallengeVerify verifies the challenge or registration
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

	var auth *dto.IdentityUserAuthDTO
	var usecaseErr *domainerrors.DomainError

	switch reqPayload.Type {
	case "register":
		auth, usecaseErr = h.ucase.VerifyRegister(ctx.Request.Context(), tenant.ID, reqPayload.FlowID, reqPayload.Code)
	case "login":
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
		h.handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, auth)
}

// Me to get user profile.
func (h *userHandler) Me(ctx *gin.Context) {
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
	sessionToken, exists := ctx.Get(string(middleware.SessionTokenKey))
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

	reqCtx := context.WithValue(ctx.Request.Context(), middleware.SessionTokenKey, sessionToken)
	requester, usecaseErr := h.ucase.Profile(reqCtx, tenant.ID)

	if usecaseErr != nil {
		h.handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, *requester)
}

// Logout to de-authenticate user.
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
	sessionToken, exists := ctx.Get(string(middleware.SessionTokenKey))
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

	reqCtx := context.WithValue(ctx.Request.Context(), middleware.SessionTokenKey, sessionToken)
	usecaseErr := h.ucase.LogOut(reqCtx, tenant.ID)
	if usecaseErr != nil {
		h.handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, nil)
}

// Register to register user.
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

	auth, usecaseErr := h.ucase.Register(ctx.Request.Context(), tenant.ID, reqPayload)
	if usecaseErr != nil {
		h.handleDomainError(ctx, usecaseErr)
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
