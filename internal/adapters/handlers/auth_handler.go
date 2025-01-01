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

// Login authenticates the user and returns a token pair (Access + Refresh).
// @Summary Authenticate user
// @Description This endpoint authenticates the user by email and password, and returns an access token and refresh token.
// @Tags authentication
// @Accept json
// @Produce json
// @Param payload body dto.LoginPayloadDTO true "User credentials (email and password)"
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
	tokenPair, err := h.ucase.Login(req.Email, req.Password)
	if err != nil {
		logger.GetLogger().Errorf("Failed to login for email %s: %v", req.Email, err)
		if errors.Is(err, domain.ErrInvalidCredentials) {
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid credentials", err)
		} else {
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to login", err)
		}
		return
	}

	// Respond with success and token pair
	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	})
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

// ValidateAccountRole validates if the role is one of the predefined roles
func (h *authHandler) validateAccountRole(role string) (constants.AccountRole, error) {
	switch constants.AccountRole(role) {
	case constants.User, constants.Partner, constants.Customer, constants.Validator:
		return constants.AccountRole(role), nil
	default:
		return "", errors.New("invalid role provided")
	}
}
