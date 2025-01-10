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
// @Router /api/v1/data-access [post]
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

	// Validate the payload
	if accountDTO.ID == payload.RequestAccountID {
		httpresponse.Error(ctx, http.StatusBadRequest, "Cannot request access to own account", nil)
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

// GetDataAccessRequests retrieves a list of data access requests filtered by status.
// @Summary Get data access requests by status
// @Description Fetches a list of data access requests for the authenticated user filtered by status.
// @Tags data-access
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param status query string false "Request status to filter by (e.g., 'PENDING', 'APPROVED', 'REJECTED')"
// @Success 200 {array} dto.DataAccessRequestDTO "List of data access requests"
// @Failure 400 {object} response.GeneralError "Invalid status"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/data-access [get]
func (h *dataAccessHandler) GetDataAccessRequests(ctx *gin.Context) {
	// Retrieve token and validate the user
	token, exists := ctx.Get("token")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Token not found", nil)
		return
	}

	accountDTO, err := h.authUCase.ValidateToken(token.(string))
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate token: %v", err)
		httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Retrieve the status query parameter
	status := ctx.DefaultQuery("status", "")
	if status != "" && !h.isValidDataAccessRequestStatus(status) {
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid status provided", nil)
		return
	}

	// Fetch requests by status
	requests, err := h.dataAccessUCase.GetRequestsByStatus(accountDTO.ID, constants.DataAccessRequestStatus(status))
	if err != nil {
		logger.GetLogger().Errorf("Failed to fetch data access requests: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to fetch data access requests", err)
		return
	}

	// Respond with the data
	ctx.JSON(http.StatusOK, gin.H{"requests": requests})
}

// isValidDataAccessRequestStatus validates if a given status is valid.
func (h *dataAccessHandler) isValidDataAccessRequestStatus(status string) bool {
	switch constants.DataAccessRequestStatus(status) {
	case constants.DataAccessRequestPending,
		constants.DataAccessRequestApproved,
		constants.DataAccessRequestRejected:
		return true
	default:
		return false
	}
}

// ApproveRequest handles approving a data access request.
// @Summary Approve a data access request
// @Description Approves a pending data access request for the authenticated user.
// @Tags data-access
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param requestID path string true "ID of the request being approved"
// @Success 200 {object} map[string]interface{} "Request approved successfully"
// @Failure 400 {object} response.GeneralError "Bad request"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/data-access/{requestID}/approve [put]
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

	// Get the requestID from the path
	requestID := ctx.Param("requestID")
	if requestID == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "Request ID is required", nil)
		return
	}

	// Approve the request
	err = h.dataAccessUCase.ApproveOrRejectRequestByID(
		accountDTO.ID,
		requestID,
		constants.DataAccessRequestApproved,
		nil, // No rejection reason needed for approval
	)
	if err != nil {
		logger.GetLogger().Errorf("Failed to approve request: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to approve request", err)
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
// @Param requestID path string true "ID of the request being rejected"
// @Param payload body dto.RejectRequestPayloadDTO true "Payload with rejection reason"
// @Success 200 {object} map[string]interface{} "Request rejected successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/data-access/{requestID}/reject [put]
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

	// Get the requestID from the path
	requestID := ctx.Param("requestID")
	if requestID == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "Request ID is required", nil)
		return
	}

	// Parse the rejection payload
	var payload dto.RejectRequestPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	// Reject the request
	err = h.dataAccessUCase.ApproveOrRejectRequestByID(
		accountDTO.ID,
		requestID,
		constants.DataAccessRequestRejected,
		&payload.Reason,
	)
	if err != nil {
		logger.GetLogger().Errorf("Failed to reject request: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to reject request", err)
		return
	}

	// Respond with success
	ctx.JSON(http.StatusOK, gin.H{"message": "Request rejected successfully"})
}

// GetAccessRequest retrieves the data access request for a specific requester, focusing on approved requests.
// @Summary Get a data access request
// @Description Fetches the data access request for a specific requester and authenticated user, prioritizing approved requests.
// @Tags data-access
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param requesterAccountID path string true "ID of the account making the request"
// @Success 200 {object} dto.DataAccessRequestDTO "Data access request details"
// @Failure 400 {object} response.GeneralError "Bad request"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 403 {object} response.GeneralError "Forbidden"
// @Failure 404 {object} response.GeneralError "Request not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/data-access/{requesterAccountID} [get]
func (h *dataAccessHandler) GetAccessRequest(ctx *gin.Context) {
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

	// Check if the authenticated user has the "DATA_OWNER" role
	if accountDTO.Role != string(constants.DataOwner) {
		httpresponse.Error(ctx, http.StatusForbidden, "Access restricted to users only", nil)
		return
	}

	// Get the requesterAccountID from the path
	requesterAccountID := ctx.Param("requesterAccountID")
	if requesterAccountID == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "Requester account ID is required", nil)
		return
	}

	// Fetch the data access request using the use case
	accessRequest, err := h.dataAccessUCase.GetAccessRequest(accountDTO.ID, requesterAccountID)
	if err != nil {
		if err.Error() == "request not found" {
			httpresponse.Error(ctx, http.StatusNotFound, "Request not found", nil)
		} else {
			logger.GetLogger().Errorf("Failed to fetch access request: %v", err)
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to fetch access request", err)
		}
		return
	}

	// Return the access request details
	ctx.JSON(http.StatusOK, gin.H{"access_request": accessRequest})
}
