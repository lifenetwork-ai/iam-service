package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	dto "github.com/genefriendway/human-network-iam/internal/delivery/dto"
	interfaces "github.com/genefriendway/human-network-iam/internal/domain/ucases/types"
	httpresponse "github.com/genefriendway/human-network-iam/packages/http/response"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

type policyHandler struct {
	ucase interfaces.AccessPolicyUseCase
}

func NewAccessPolicyHandler(ucase interfaces.AccessPolicyUseCase) *policyHandler {
	return &policyHandler{
		ucase: ucase,
	}
}

// GetPolicies retrieves a list of policies.
// @Summary Retrieve policies
// @Description Get policies
// @Tags policies
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of policies"
// @Failure 400 {object} response.ErrorResponse "Invalid page number or size"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/policies [get]
func (h *policyHandler) GetPolicies(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")
	keyword := ctx.DefaultQuery("keyword", "")

	// Parse page and size into integers
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		logger.GetLogger().Errorf("Invalid page number: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAGE_NUMBER",
			"Invalid page number",
			err,
		)
		return
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		logger.GetLogger().Errorf("Invalid size: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_SIZE",
			"Invalid size",
			err,
		)
		return
	}

	response, errResponse := h.ucase.List(ctx, pageInt, sizeInt, keyword)
	if errResponse != nil {
		logger.GetLogger().Errorf("Failed to get policies: %v", errResponse)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_FAILED_TO_GET_POLICIES",
			"Failed to get policies",
			errResponse,
		)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}

// GetDetail retrieves a policy by it's ID.
// @Summary Retrieve policy by ID
// @Description Get policy by ID
// @Tags policies
// @Accept json
// @Produce json
// @Param policy_id path string true "policy ID"
// @Success 200 {object} dto.AccessPolicyDTO "Successful retrieval of policy"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "policy not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/policies/{policy_id} [get]
func (h *policyHandler) GetDetail(ctx *gin.Context) {
	// Extract and parse policy_id from query string
	policyId := ctx.Query("policy_id")
	if policyId == "" {
		logger.GetLogger().Error("Invalid policy ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	policy, err := h.ucase.GetByID(ctx, policyId)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get policy: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get policy"})
		return
	}

	// Return the response as a JSON response
	httpresponse.Success(ctx, http.StatusOK, policy)
}

// CreatePolicy creates a new policy.
// @Summary Create a new policy
// @Description Create a new policy
// @Tags policies
// @Accept json
// @Produce json
// @Param policy body dto.CreateAccessPolicyPayloadDTO true "policy payload"
// @Success 201 {object} dto.AccessPolicyDTO "Successful creation of policy"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/policies [post]
func (h *policyHandler) CreatePolicy(ctx *gin.Context) {
	var reqPayload dto.CreateAccessPolicyPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Failed to create policy, invalid payload",
			err,
		)
		return
	}

	// Create the policy
	response, err := h.ucase.Create(ctx, reqPayload)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create policy: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusInternalServerError,
			"MSG_FAILED_TO_CREATE_POLICY",
			"Failed to create policy",
			err,
		)
		return
	}

	// Return the response as a JSON response
	httpresponse.Success(ctx, http.StatusCreated, response)
}

// UpdatePolicy updates an existing policy.
// @Summary Update an existing policy
// @Description Update an existing policy
// @Tags policies
// @Accept json
// @Produce json
// @Param policy_id path string true "policy ID"
// @Param policy body dto.UpdateAccessPolicyPayloadDTO true "policy payload"
// @Success 200 {object} dto.AccessPolicyDTO "Successful update of policy"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "policy not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/policies/{policy_id} [put]
func (h *policyHandler) UpdatePolicy(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// DeletePolicy deletes an existing policy.
// @Summary Delete an existing policy
// @Description Delete an existing policy
// @Tags policies
// @Accept json
// @Produce json
// @Param policy_id path string true "policy ID"
// @Success 204 "Successful deletion of policy"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "policy not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/policies/{policy_id} [delete]
func (h *policyHandler) DeletePolicy(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}
