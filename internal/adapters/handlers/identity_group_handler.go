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

type groupHandler struct {
	ucase interfaces.IdentityGroupUseCase
}

func NewIdentityGroupHandler(ucase interfaces.IdentityGroupUseCase) *groupHandler {
	return &groupHandler{
		ucase: ucase,
	}
}

// GetGroups retrieves a list of groups.
// @Summary Retrieve groups
// @Description Get groups
// @Tags groups
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of groups"
// @Failure 400 {object} response.ErrorResponse "Invalid page number or size"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/groups [get]
func (h *groupHandler) GetGroups(ctx *gin.Context) {
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
		logger.GetLogger().Errorf("Failed to get groups: %v", errResponse)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_FAILED_TO_GET_GROUPS",
			"Failed to get groups",
			errResponse,
		)
		return
	}

	// Return the response as a JSON response
	httpresponse.Success(ctx, http.StatusOK, response)
}

// GetDetail retrieves a group by it's ID.
// @Summary Retrieve group by ID
// @Description Get group by ID
// @Tags groups
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param group_id path string true "group ID"
// @Success 200 {object} dto.IdentityGroupDTO "Successful retrieval of group"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "group not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/groups/{group_id} [get]
func (h *groupHandler) GetDetail(ctx *gin.Context) {
	// Extract and parse group_id from query string
	groupId := ctx.Query("group_id")
	if groupId == "" {
		logger.GetLogger().Error("Invalid group ID")
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_GROUP_ID", "Invalid group ID", nil)
		return
	}

	group, err := h.ucase.GetByID(ctx, groupId)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get group: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "MSG_FAILED_TO_GET_GROUP", "Failed to get group", err)
		return
	}

	// Return the response as a JSON response
	httpresponse.Success(ctx, http.StatusOK, group)
}

// CreateGroup creates a new group.
// @Summary Create a new group
// @Description Create a new group
// @Tags groups
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param group body dto.CreateIdentityGroupPayloadDTO true "group payload"
// @Success 201 {object} dto.IdentityGroupDTO "Successful creation of group"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/groups [post]
func (h *groupHandler) CreateGroup(ctx *gin.Context) {
	var reqPayload dto.CreateIdentityGroupPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Failed to create group, invalid payload",
			err,
		)
		return
	}

	// Create the group
	response, err := h.ucase.Create(ctx, reqPayload)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create group: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusInternalServerError,
			"MSG_FAILED_TO_CREATE_GROUP",
			"Failed to create group",
			err,
		)
		return
	}

	// Return the response as a JSON response
	httpresponse.Success(ctx, http.StatusCreated, response)
}

// UpdateGroup updates an existing group.
// @Summary Update an existing group
// @Description Update an existing group
// @Tags groups
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param group_id path string true "group ID"
// @Param group body dto.UpdateIdentityGroupPayloadDTO true "group payload"
// @Success 200 {object} dto.IdentityGroupDTO "Successful update of group"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "group not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/groups/{group_id} [put]
func (h *groupHandler) UpdateGroup(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// DeleteGroup deletes an existing group.
// @Summary Delete an existing group
// @Description Delete an existing group
// @Tags groups
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param group_id path string true "group ID"
// @Success 204 "Successful deletion of group"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "group not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/groups/{group_id} [delete]
func (h *groupHandler) DeleteGroup(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}
