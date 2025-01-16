package handlers

import (
	"net/http"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/dto"
	httpresponse "github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
	"github.com/gin-gonic/gin"
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
