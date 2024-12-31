package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

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
