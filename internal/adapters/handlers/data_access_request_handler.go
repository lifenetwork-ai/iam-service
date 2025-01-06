package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/constants"
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

// ApproveRequest handles approving a data access request.
// @Summary Approve a data access request
// @Description Approves a pending data access request for the authenticated user.
// @Tags data-access
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param requesterAccountID path string true "ID of the account making the request"
// @Success 200 {object} map[string]interface{} "Request approved successfully"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 404 {object} response.GeneralError "Request not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/data-access/{requesterAccountID}/approve [put]
func (h *dataAccessHandler) ApproveRequest(ctx *gin.Context) {
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

	// Get the requesterAccountID from the path
	requesterAccountID := ctx.Param("requesterAccountID")
	if requesterAccountID == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "Requester account ID is required", nil)
		return
	}

	// Approve the request
	err = h.dataAccessUCase.ApproveOrRejectRequest(
		accountDTO.ID,
		requesterAccountID,
		constants.DataAccessRequestApproved,
		nil, // No rejection reason needed for approval
	)
	if err != nil {
		if err.Error() == "request not found" {
			httpresponse.Error(ctx, http.StatusNotFound, "Request not found", nil)
		} else {
			logger.GetLogger().Errorf("Failed to approve request: %v", err)
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to approve request", err)
		}
		return
	}

	// Respond with success
	ctx.JSON(http.StatusOK, gin.H{"message": "Request approved successfully"})
}

// RejectRequest handles rejecting a data access request.
// @Summary Reject a data access request
// @Description Rejects a pending data access request for the authenticated user.
// @Tags data-access
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param requesterAccountID path string true "ID of the account making the request"
// @Param payload body map[string]string true "Payload containing the reason for rejection"
// @Success 200 {object} map[string]interface{} "Request rejected successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 404 {object} response.GeneralError "Request not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/data-access/{requesterAccountID}/reject [put]
func (h *dataAccessHandler) RejectRequest(ctx *gin.Context) {
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

	// Get the requesterAccountID from the path
	requesterAccountID := ctx.Param("requesterAccountID")
	if requesterAccountID == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "Requester account ID is required", nil)
		return
	}

	// Parse the rejection reason from the request body
	var payload map[string]string
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	reasonForRejection, exists := payload["reason"]
	if !exists || reasonForRejection == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "Reason for rejection is required", nil)
		return
	}

	// Reject the request
	err = h.dataAccessUCase.ApproveOrRejectRequest(
		accountDTO.ID,
		requesterAccountID,
		constants.DataAccessRequestRejected,
		&reasonForRejection,
	)
	if err != nil {
		if err.Error() == "request not found" {
			httpresponse.Error(ctx, http.StatusNotFound, "Request not found", nil)
		} else {
			logger.GetLogger().Errorf("Failed to reject request: %v", err)
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to reject request", err)
		}
		return
	}

	// Respond with success
	ctx.JSON(http.StatusOK, gin.H{"message": "Request rejected successfully"})
}
