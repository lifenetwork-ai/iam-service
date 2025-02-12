package identity_group

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
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
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of groups"
// @Failure 400 {object} response.GeneralError "Invalid page number or size"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/groups [get]
func (h *groupHandler) GetGroups(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")
	keyword := ctx.DefaultQuery("keyword", "")

	// Parse page and size into integers
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		logger.GetLogger().Errorf("Invalid page number: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		logger.GetLogger().Errorf("Invalid size: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid size"})
		return
	}

	response, errResponse := h.ucase.List(ctx, pageInt, sizeInt, keyword)
	if errResponse != nil {
		logger.GetLogger().Errorf("Failed to get groups: %v", errResponse)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to get groups", errResponse)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}

// GetDetail retrieves a group by it's ID.
// @Summary Retrieve group by ID
// @Description Get group by ID
// @Tags groups
// @Accept json
// @Produce json
// @Param group_id path string true "group ID"
// @Success 200 {object} dto.groupDTO "Successful retrieval of group"
// @Failure 400 {object} response.GeneralError "Invalid request ID"
// @Failure 404 {object} response.GeneralError "group not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/groups/{group_id} [get]
func (h *groupHandler) GetDetail(ctx *gin.Context) {
	// Extract and parse group_id from query string
	groupId := ctx.Query("group_id")
	if groupId == "" {
		logger.GetLogger().Error("Invalid group ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	group, err := h.ucase.GetByID(ctx, groupId)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get group: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get group"})
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, group)
}

// CreateGroup creates a new group.
// @Summary Create a new group
// @Description Create a new group
// @Tags groups
// @Accept json
// @Produce json
// @Param group body dto.groupCreatePayloadDTO true "group payload"
// @Success 201 {object} dto.groupDTO "Successful creation of group"
// @Failure 400 {object} response.GeneralError "Invalid request payload"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/groups [post]
func (h *groupHandler) CreateGroup(ctx *gin.Context) {
	var reqPayload dto.CreateIdentityGroupPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to create payment orders, invalid payload", err)
		return
	}

	// Create the group
	response, err := h.ucase.Create(ctx, reqPayload)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create group: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to create group", err)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusCreated, response)
}

// UpdateGroup updates an existing group.
// @Summary Update an existing group
// @Description Update an existing group
// @Tags groups
// @Accept json
// @Produce json
// @Param group_id path string true "group ID"
// @Param group body dto.groupUpdatePayloadDTO true "group payload"
// @Success 200 {object} dto.groupDTO "Successful update of group"
// @Failure 400 {object} response.GeneralError "Invalid request payload"
// @Failure 404 {object} response.GeneralError "group not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/groups/{group_id} [put]
func (h *groupHandler) UpdateGroup(ctx *gin.Context) {

}

// DeleteGroup deletes an existing group.
// @Summary Delete an existing group
// @Description Delete an existing group
// @Tags groups
// @Accept json
// @Produce json
// @Param group_id path string true "group ID"
// @Success 204 "Successful deletion of group"
// @Failure 400 {object} response.GeneralError "Invalid request ID"
// @Failure 404 {object} response.GeneralError "group not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/groups/{group_id} [delete]
func (h *groupHandler) DeleteGroup(ctx *gin.Context) {

}
