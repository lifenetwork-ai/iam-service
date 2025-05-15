package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type serviceHandler struct {
	ucase interfaces.IdentityServiceUseCase
}

func NewIdentityServiceHandler(ucase interfaces.IdentityServiceUseCase) *serviceHandler {
	return &serviceHandler{
		ucase: ucase,
	}
}

// GetServices retrieves a list of services.
// @Summary Retrieve services
// @Description Get services
// @Tags services
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of services"
// @Failure 400 {object} response.ErrorResponse "Invalid page number or size"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/services [get]
func (h *serviceHandler) GetServices(ctx *gin.Context) {
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
		logger.GetLogger().Errorf("Failed to get services: %v", errResponse)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_SIZE",
			"Invalid size",
			err,
		)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}

// GetDetail retrieves a service by it's ID.
// @Summary Retrieve service by ID
// @Description Get service by ID
// @Tags services
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param service_id path string true "service ID"
// @Success 200 {object} dto.IdentityServiceDTO "Successful retrieval of service"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "service not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/services/{service_id} [get]
func (h *serviceHandler) GetDetail(ctx *gin.Context) {
	// Extract and parse service_id from query string
	serviceId := ctx.Query("service_id")
	if serviceId == "" {
		logger.GetLogger().Error("Invalid service ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	service, err := h.ucase.GetByID(ctx, serviceId)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get service: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get service"})
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, service)
}

// CreateService creates a new service.
// @Summary Create a new service
// @Description Create a new service
// @Tags services
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param service body dto.CreateIdentityServicePayloadDTO true "service payload"
// @Success 201 {object} dto.IdentityServiceDTO "Successful creation of service"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/services [post]
func (h *serviceHandler) CreateService(ctx *gin.Context) {
	var reqPayload dto.CreateIdentityServicePayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Failed to create service, invalid payload",
			err)
		return
	}

	// Create the service
	response, err := h.ucase.Create(ctx, reqPayload)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create service: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusInternalServerError,
			"MSG_FAILED_TO_CREATE_SERVICE",
			"Failed to create service",
			err,
		)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusCreated, response)
}

// UpdateService updates an existing service.
// @Summary Update an existing service
// @Description Update an existing service
// @Tags services
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param service_id path string true "Service ID"
// @Param service body dto.UpdateIdentityServicePayloadDTO true "service payload"
// @Success 200 {object} dto.IdentityServiceDTO "Successful update of service"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "service not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/services/{service_id} [put]
func (h *serviceHandler) UpdateService(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// DeleteService deletes an existing service.
// @Summary Delete an existing service
// @Description Delete an existing service
// @Tags services
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param service_id path string true "Service ID"
// @Success 204 "Successful deletion of service"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "service not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/services/{service_id} [delete]
func (h *serviceHandler) DeleteService(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}
