package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	httpresponse "github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
)

type authHandler struct {
	ucase interfaces.AuthUCase
}

func NewAuthHandler(
	ucase interfaces.AuthUCase,
) *authHandler {
	return &authHandler{
		ucase: ucase,
	}
}

// Register creates a new user account and role-specific details.
// @Summary Register a new account
// @Description This endpoint registers a new account and its associated role-specific details.
// @Tags authentication
// @Accept json
// @Produce json
// @Param payload body dto.RegisterPayloadDTO true "User registration details"
// @Success 201 {object} map[string]interface{} "Registration successful: {\"success\": true}"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 409 {object} response.GeneralError "Account already exists"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *authHandler) Register(ctx *gin.Context) {
	var req dto.RegisterPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid registration payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to register, invalid payload", err)
		return
	}

	// Validate the role
	role, err := h.validateAccountRole(req.Role)
	if err != nil {
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid role provided", err)
		return
	}

	// Register the user using the use case
	err = h.ucase.Register(&req, role)
	if err != nil {
		logger.GetLogger().Errorf("Failed to register user: %v", err)
		if errors.Is(err, domain.ErrAccountAlreadyExists) {
			httpresponse.Error(ctx, http.StatusConflict, "Account already exists", err)
		} else {
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to register", err)
		}
		return
	}

	// Respond with success
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// Login authenticates the user and returns a token pair (Access + Refresh).
// @Summary Authenticate user
// @Description This endpoint authenticates the user by email, username, or phone number, and returns an access token and refresh token.
// @Tags authentication
// @Accept json
// @Produce json
// @Param payload body dto.LoginPayloadDTO true "User credentials (identifier, password, and identifierType)"
// @Success 200 {object} dto.TokenPairDTO "Login successful: {\"access_token\": \"...\", \"refresh_token\": \"...\"}"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 401 {object} response.GeneralError "Invalid credentials"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *authHandler) Login(ctx *gin.Context) {
	var req dto.LoginPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid login payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to login, invalid payload", err)
		return
	}

	// Authenticate the user using the use case
	tokenPair, err := h.ucase.Login(req.Identifier, req.Password, req.IdentifierType)
	if err != nil {
		logger.GetLogger().Errorf("Failed to login for identifier %s (type: %s): %v", req.Identifier, req.IdentifierType, err)
		if errors.Is(err, domain.ErrInvalidCredentials) {
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid credentials", err)
		} else {
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to login", err)
		}
		return
	}

	// Respond with success and token pair
	ctx.JSON(http.StatusOK, dto.TokenPairDTO{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

// ValidateAccountRole validates if the role is one of the predefined roles
func (h *authHandler) validateAccountRole(role string) (constants.AccountRole, error) {
	switch constants.AccountRole(role) {
	case constants.User, constants.Partner, constants.Customer, constants.Validator:
		return constants.AccountRole(role), nil
	default:
		return "", errors.New("invalid role provided")
	}
}

// Logout invalidates the refresh token.
// @Summary Logout user
// @Description This endpoint invalidates the refresh token, effectively logging the user out.
// @Tags authentication
// @Accept json
// @Produce json
// @Param payload body dto.LogoutPayloadDTO true "Logout payload containing the refresh token"
// @Success 200 {object} map[string]interface{} "Logout successful: {\"success\": true}"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 401 {object} response.GeneralError "Invalid or expired refresh token"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/auth/logout [post]
func (h *authHandler) Logout(ctx *gin.Context) {
	var req dto.LogoutPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid logout payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to logout, invalid payload", err)
		return
	}

	// Invalidate the refresh token using the use case
	err := h.ucase.Logout(req.RefreshToken)
	if err != nil {
		logger.GetLogger().Errorf("Failed to logout: %v", err)
		if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrExpiredToken) {
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid or expired refresh token", err)
		} else {
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to logout", err)
		}
		return
	}

	// Respond with success
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// RefreshTokens generates new access and refresh tokens using a valid refresh token.
// @Summary Refresh tokens
// @Description This endpoint generates a new pair of access and refresh tokens using a valid refresh token.
// @Tags authentication
// @Accept json
// @Produce json
// @Param payload body dto.RefreshTokenPayloadDTO true "Refresh token payload"
// @Success 200 {object} dto.TokenPairDTO "Tokens refreshed successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 401 {object} response.GeneralError "Invalid or expired refresh token"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/auth/refresh-tokens [post]
func (h *authHandler) RefreshTokens(ctx *gin.Context) {
	var req dto.RefreshTokenPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid refresh token payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to refresh tokens, invalid payload", err)
		return
	}

	// Refresh tokens using the use case
	tokenPair, err := h.ucase.RefreshTokens(req.RefreshToken)
	if err != nil {
		logger.GetLogger().Errorf("Failed to refresh tokens: %v", err)
		if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrExpiredToken) {
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid or expired refresh token", err)
		} else {
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to refresh tokens", err)
		}
		return
	}

	// Respond with the new token pair
	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	})
}

// @Summary Validate access token
// @Description This endpoint validates an access token and retrieves the corresponding user's details.
// @Tags authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Success 200 {object} dto.AccountDTO "Token is valid: User details"
// @Failure 401 {object} response.GeneralError "Invalid or expired token"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/auth/validate-token [get]
func (h *authHandler) ValidateToken(ctx *gin.Context) {
	// Retrieve the token from the context
	token, exists := ctx.Get("token")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Token not found", nil)
		return
	}

	// Validate the token using the use case
	account, err := h.ucase.ValidateToken(token.(string))
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate token: %v", err)
		if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrExpiredToken) {
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid or expired token", err)
		} else {
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to validate token", err)
		}
		return
	}

	// Respond with user details
	ctx.JSON(http.StatusOK, account)
}
