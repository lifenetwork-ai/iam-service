package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	httpresponse "github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
)

type dataAccessHandler struct {
	dataAccessUCase interfaces.DataAccessUCase
	authUCase       interfaces.AuthUCase
}

func NewDataAccessHandler(
	dataAccessUCase interfaces.DataAccessUCase,
	authUCase interfaces.AuthUCase,
) *dataAccessHandler {
	return &dataAccessHandler{
		dataAccessUCase: dataAccessUCase,
		authUCase:       authUCase,
	}
}

// CreateDataAccessRequest handles the logic to create a new data access request.
// @Summary Create a data access request
// @Description Allows a requester to create a new data access request for a specific user.
// @Tags data-access
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param payload body dto.DataAccessRequestPayloadDTO true "Payload containing user ID and reason for request"
// @Success 201 {object} map[string]interface{} "Data access request created successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 404 {object} response.GeneralError "Requested user not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/data-requests [post]
func (h *dataAccessHandler) CreateDataAccessRequest(ctx *gin.Context) {
	// Retrieve the token from the context
	token, exists := ctx.Get("token")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Token not found", nil)
		return
	}

	// Validate the token and fetch requester details
	accountDTO, err := h.authUCase.ValidateToken(token.(string))
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate token: %v", err)
		httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Parse the request payload
	var payload dto.DataAccessRequestPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	// Create the data access request
	err = h.dataAccessUCase.CreateRequest(
		payload,
		accountDTO.ID,
		accountDTO.Role,
	)
	if err != nil {
		if err.Error() == "requested account not found" {
			httpresponse.Error(ctx, http.StatusNotFound, "Requested user not found", nil)
		} else {
			logger.GetLogger().Errorf("Failed to create data access request: %v", err)
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to create data access request", err)
		}
		return
	}

	// Return success response
	ctx.JSON(http.StatusCreated, gin.H{"message": "Data access request created successfully"})
}

// GetPendingDataAccessRequests retrieves a list of pending data access requests.
// @Summary Get pending data access requests
// @Description Fetches a list of pending data access requests for the authenticated user.
// @Tags data-access
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Success 200 {array} dto.DataAccessRequestDTO "List of pending data access requests"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/data-requests/pending [get]
func (h *dataAccessHandler) GetPendingDataAccessRequests(ctx *gin.Context) {
	// Retrieve the token from the context
	token, exists := ctx.Get("token")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Token not found", nil)
		return
	}

	// Validate the token and fetch user details
	accountDTO, err := h.authUCase.ValidateToken(token.(string))
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate token: %v", err)
		httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Fetch the pending requests
	requests, err := h.dataAccessUCase.GetPendingRequests(accountDTO.ID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to fetch pending data access requests: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to fetch pending data access requests", err)
		return
	}

	// Return the list of pending requests
	ctx.JSON(http.StatusOK, gin.H{"requests": requests})
}
