package identity_organization

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-iam/internal/interfaces"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

type organizationHandler struct {
	ucase interfaces.OrganizationUseCase
}

func NewOrganizationHandler(ucase interfaces.OrganizationUseCase) *organizationHandler {
	return &organizationHandler{
		ucase: ucase,
	}
}

// GetOrganizations retrieves a list of organizations.
// @Summary Retrieve organizations
// @Description Get organizations
// @Tags organization
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of organizations"
// @Failure 400 {object} response.GeneralError "Invalid page number or size"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/organization [get]
func (h *organizationHandler) GetOrganizations(ctx *gin.Context) {
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

	organizations, err := h.ucase.GetOrganizations(ctx, pageInt, sizeInt, keyword)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get organizations: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organizations"})
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, organizations)
}

// GetOrganizationByID retrieves a organization by it's ID.
// @Summary Retrieve organization by ID
// @Description Get organization by ID
// @Tags organization
// @Accept json
// @Produce json
// @Param organization_id path string true "organization ID"
// @Success 200 {object} dto.OrganizationDTO "Successful retrieval of organization"
// @Failure 400 {object} response.GeneralError "Invalid request ID"
// @Failure 404 {object} response.GeneralError "organization not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/organization/{organization_id} [get]
func (h *organizationHandler) GetOrganizationByID(ctx *gin.Context) {
	// Extract and parse organization_id from query string
	organizationId := ctx.Query("organization_id")
	if organizationId == "" {
		logger.GetLogger().Error("Invalid organization ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	organization, err := h.ucase.GetOrganizationByID(ctx, organizationId)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get organization: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization"})
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, organization)
}

// CreateOrganization creates a new organization.
// @Summary Create a new organization
// @Description Create a new organization
// @Tags organization
// @Accept json
// @Produce json
// @Param organization body dto.OrganizationCreatePayloadDTO true "organization payload"
// @Success 201 {object} dto.OrganizationDTO "Successful creation of organization"
// @Failure 400 {object} response.GeneralError "Invalid request payload"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/organization [post]
func (h *organizationHandler) CreateOrganization(ctx *gin.Context) {

}

// UpdateOrganization updates an existing organization.
// @Summary Update an existing organization
// @Description Update an existing organization
// @Tags organization
// @Accept json
// @Produce json
// @Param organization_id path string true "organization ID"
// @Param organization body dto.OrganizationUpdatePayloadDTO true "organization payload"
// @Success 200 {object} dto.OrganizationDTO "Successful update of organization"
// @Failure 400 {object} response.GeneralError "Invalid request payload"
// @Failure 404 {object} response.GeneralError "organization not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/organization/{organization_id} [put]
func (h *organizationHandler) UpdateOrganization(ctx *gin.Context) {

}

// DeleteOrganization deletes an existing organization.
// @Summary Delete an existing organization
// @Description Delete an existing organization
// @Tags organization
// @Accept json
// @Produce json
// @Param organization_id path string true "organization ID"
// @Success 204 "Successful deletion of organization"
// @Failure 400 {object} response.GeneralError "Invalid request ID"
// @Failure 404 {object} response.GeneralError "organization not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/organization/{organization_id} [delete]
func (h *organizationHandler) DeleteOrganization(ctx *gin.Context) {

}
