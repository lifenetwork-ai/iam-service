package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/conf"
	"github.com/genefriendway/human-network-auth/internal/adapters/handlers"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	"github.com/genefriendway/human-network-auth/internal/middleware"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	db *gorm.DB,
	fileObjectUCase interfaces.FileObjectUCase,
) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: file object
	fileObjectHandler := handlers.NewFileObjectHandler(fileObjectUCase, config)
	appRouter.POST("/file-object/upload", middleware.ValidateVendorID(), fileObjectHandler.UploadFile)
	appRouter.GET("/file-object/:object_id", middleware.ValidateVendorID(), fileObjectHandler.GetDetail)
}
