package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/conf"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	httpresponse "github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
)

type fileObjectHandler struct {
	ucase  interfaces.FileObjectUCase
	config *conf.Configuration
}

func NewFileObjectHandler(
	ucase interfaces.FileObjectUCase,
	config *conf.Configuration,
) *fileObjectHandler {
	return &fileObjectHandler{
		ucase:  ucase,
		config: config,
	}
}

// GetNetworksMetadata retrieves all networks metadata.
// @Summary Retrieves all networks metadata.
// @Description Retrieves all networks metadata.
// @Tags metadata
// @Accept json
// @Produce json
// @Success 200 {array} dto.NetworkMetadataDTO
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/metadata/networks [get]
func (h *fileObjectHandler) UploadFile(ctx *gin.Context) {
	metadata, err := h.ucase.UploadFile(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve networks metadata: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve networks metadata", err)
		return
	}
	ctx.JSON(http.StatusOK, metadata)
}

func (h *fileObjectHandler) GetDetail(ctx *gin.Context) {
	objectID := ctx.Param("object_id")
	metadata, err := h.ucase.GetDetail(ctx, objectID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve networks metadata: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve networks metadata", err)
		return
	}
	ctx.JSON(http.StatusOK, metadata)
}
