package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/dto"
	httpresponse "github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary Set file validation status
// @Description Updates the validation status of a file (VALID or INVALID)
// @Tags data-access
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param payload body dto.ValidationRequestDTO true "Validation status details"
// @Success 200 {object} object{message=string} "Validation status updated successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload or status"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 403 {object} response.GeneralError "Forbidden - Not a validator"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/data-access/validator/validate [post]
func (h *dataAccessHandler) ValidateFileContent(ctx *gin.Context) {
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

	if accountDTO.Role != constants.Validator.String() {
		httpresponse.Error(ctx, http.StatusForbidden, "Access denied - Not a validator", nil)
		return
	}

	var validationRequestDTO dto.ValidationRequestDTO
	if err := ctx.ShouldBindJSON(&validationRequestDTO); err != nil {
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	requestID := validationRequestDTO.RequestID
	if requestID == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "Request ID is required", nil)
		return
	}

	// get the status from the request
	status := validationRequestDTO.Status
	err = h.dataAccessUCase.ValidatorValidateFileContent(accountDTO.ID, requestID, constants.RequesterRequestStatus(status), validationRequestDTO.Msg)
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate file content: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to validate file content", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "File content validated successfully"})
}

// @Summary Get validator request detail
// @Description Fetches detailed information about a specific validation request
// @Tags data-access
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param requestID path string true "Request ID to get details for"
// @Success 200 {object} domain.DataAccessRequestRequesterTest "Request details retrieved successfully"
// @Failure 400 {object} response.GeneralError "Invalid request ID"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 403 {object} response.GeneralError "Forbidden - Not a validator"
// @Failure 404 {object} response.GeneralError "Request not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/data-access/validator/requests/detail/{requestID} [get]
func (h *dataAccessHandler) GetRequestDetail(ctx *gin.Context) {
	// Get token from context
	token, exists := ctx.Get("token")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Token not found", nil)
		return
	}

	// Validate token and get account details
	accountDTO, err := h.authUCase.ValidateToken(token.(string))
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate token: %v", err)
		httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Ensure user is a validator
	if accountDTO.Role != constants.Validator.String() {
		httpresponse.Error(ctx, http.StatusForbidden, "Access denied - Not a validator", nil)
		return
	}

	// Get request ID from path
	requestID := ctx.Param("requestID")
	if requestID == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "Request ID is required", nil)
		return
	}

	// Get request details
	detail, err := h.dataAccessUCase.ValidatorGetRequestDetail(accountDTO.ID, requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpresponse.Error(ctx, http.StatusNotFound, "Request not found", err)
			return
		}
		logger.GetLogger().Errorf("Failed to get request detail: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to get request detail", err)
		return
	}

	// Call to secure genom to get the capsule
	reencryptedData, err := h.GetDetailFromSecureGenom(ctx.Request.Header, detail.FileInfo.OwnerID, detail.FileInfo.ID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get reencrypted data: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Genom: Failed to get reencrypted data", err)
		return
	}

	res := dto.RequesterRequestDetailDTO{
		ReencryptedDataDTO:  reencryptedData,
		RequesterRequestDTO: detail,
	}
	ctx.JSON(http.StatusOK, res)
}

type RequestDataAccess struct {
	OwnerID string `json:"owner_id"`
	DataID  string `json:"data_id"`
}

func (h *dataAccessHandler) GetDetailFromSecureGenom(headers http.Header, ownerID, dataID string) (dto.ReencryptedDataDTO, error) {
	// Prepare the request payload
	payload := RequestDataAccess{
		OwnerID: ownerID,
		DataID:  dataID,
	}

	// Create request
	url := "http://localhost:8081/api/v1/validator/access-data"
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return dto.ReencryptedDataDTO{}, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return dto.ReencryptedDataDTO{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers from the incoming request
	req.Header = headers
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return dto.ReencryptedDataDTO{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dto.ReencryptedDataDTO{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return dto.ReencryptedDataDTO{}, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response dto.ReencryptedDataDTO
	if err := json.Unmarshal(body, &response); err != nil {
		return dto.ReencryptedDataDTO{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}
