package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-iam/conf"
	"github.com/genefriendway/human-network-iam/internal/adapters/handlers"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	db *gorm.DB,
	organizationUCase interfaces.IdentityOrganizationUseCase,
) {
	v1 := r.Group("/api/v1")

	// SECTION: organization
	organizationRouter := v1.Group("organizations")
	organizationHandler := handlers.NewIdentityOrganizationHandler(organizationUCase)
	organizationRouter.GET("/", organizationHandler.GetOrganizations)
	organizationRouter.GET("/:organization_id", organizationHandler.GetDetail)
	organizationRouter.POST("/", organizationHandler.CreateOrganization)
	organizationRouter.PUT("/:organization_id", organizationHandler.UpdateOrganization)
	organizationRouter.DELETE("/:organization_id", organizationHandler.DeleteOrganization)

	// SECTION: identity
	// identityRouter := v1.Group("auth")
}
