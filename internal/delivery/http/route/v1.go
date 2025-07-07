package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/handlers"
	middleware "github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	db *gorm.DB,
	organizationUCase interfaces.IdentityOrganizationUseCase,
	userUCase interfaces.IdentityUserUseCase,
	adminUCase interfaces.AdminUseCase,
) {
	v1 := r.Group("/api/v1")

	// SECTION: Admin routes
	adminRouter := v1.Group("admin")
	adminRouter.Use(
		middleware.AdminBasicAuthMiddleware(),
	)

	// Admin Tenant Management subgroup
	adminHandler := handlers.NewAdminHandler(adminUCase)
	tenantRouter := adminRouter.Group("tenants")
	{
		tenantRouter.GET("/", adminHandler.ListTenants)
		tenantRouter.GET("/:id", adminHandler.GetTenant)
		tenantRouter.POST("/", adminHandler.CreateTenant)
		tenantRouter.PUT("/:id", adminHandler.UpdateTenant)
		tenantRouter.DELETE("/:id", adminHandler.DeleteTenant)
		tenantRouter.PUT("/:id/status", adminHandler.UpdateTenantStatus)
	}

	// SECTION: organizations
	organizationRouter := v1.Group("organizations")
	organizationHandler := handlers.NewIdentityOrganizationHandler(organizationUCase)
	organizationRouter.GET(
		"/",
		middleware.RequestAuthenticationMiddleware(),
		middleware.RequestAuthorizationMiddleware("iam:identity_organization:read"),
		organizationHandler.GetOrganizations,
	)
	organizationRouter.GET(
		"/:organization_id",
		middleware.RequestAuthenticationMiddleware(),
		middleware.RequestAuthorizationMiddleware("iam:identity_organization:read"),
		organizationHandler.GetDetail,
	)
	organizationRouter.POST(
		"/",
		middleware.RequestAuthenticationMiddleware(),
		middleware.RequestAuthorizationMiddleware("iam:identity_organization:create"),
		organizationHandler.CreateOrganization,
	)
	organizationRouter.PUT(
		"/:organization_id",
		middleware.RequestAuthenticationMiddleware(),
		middleware.RequestAuthorizationMiddleware("iam:identity_organization:update"),
		organizationHandler.UpdateOrganization,
	)
	organizationRouter.DELETE(
		"/:organization_id",
		middleware.RequestAuthenticationMiddleware(),
		middleware.RequestAuthorizationMiddleware("iam:identity_organization:delete"),
		organizationHandler.DeleteOrganization,
	)

	// SECTION: users
	userRouter := v1.Group("users")
	userHandler := handlers.NewIdentityUserHandler(userUCase)
	userRouter.POST(
		"/challenge-with-phone",
		userHandler.ChallengeWithPhone,
	)

	userRouter.POST(
		"/challenge-with-email",
		userHandler.ChallengeWithEmail,
	)

	userRouter.POST(
		"/challenge-verify",
		userHandler.ChallengeVerify,
	)

	userRouter.POST(
		"/register",
		userHandler.Register,
	)

	userRouter.POST(
		"/logout",
		middleware.RequestAuthenticationMiddleware(),
		userHandler.Logout,
	)
	userRouter.GET(
		"/me",
		middleware.RequestAuthenticationMiddleware(),
		userHandler.Me,
	)
}
