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
	organizationUCase interfaces.OrganizationUseCase,
) {
	v1 := r.Group("/api/v1")

	// SECTION: organization
	organizationRouter := v1.Group("organizations")
	organizationHandler := handlers.NewOrganizationHandler(organizationUCase)
	organizationRouter.GET("/", organizationHandler.GetOrganizations)
	organizationRouter.GET("/:organizationID", organizationHandler.GetOrganizationByID)
	organizationRouter.POST("/", organizationHandler.CreateOrganization)
	organizationRouter.PUT("/:organizationID", organizationHandler.UpdateOrganization)
	organizationRouter.DELETE("/:organizationID", organizationHandler.DeleteOrganization)
	organizationRouter.GET("/:organizationID/members", organizationHandler.GetMembers)
	organizationRouter.POST("/:organizationID/members/:memberID", organizationHandler.AddMember)
	organizationRouter.DELETE("/:organizationID/members/:memberID", organizationHandler.RemoveMember)

	// SECTION: identity
	// identityRouter := v1.Group("auth")
}
