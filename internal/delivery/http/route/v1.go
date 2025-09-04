package route

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/handlers"
	middleware "github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	"github.com/lifenetwork-ai/iam-service/internal/wire"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	ucases *wire.UseCases,
	repos *wire.Repos,
) {
	authMiddleware := middleware.NewAuthMiddleware(ucases.IdentityUserUCase)
	v1 := r.Group("/api/v1")

	// SECTION: Admin routes
	adminRouter := v1.Group("admin")

	// Admin Tenant Management subgroup
	adminHandler := handlers.NewAdminHandler(ucases.AdminUCase, ucases.IdentityUserUCase)
	accountRouter := adminRouter.Group("accounts")
	{
		accountRouter.Use(
			middleware.RootAuthMiddleware(),
		)
		accountRouter.POST("/", adminHandler.CreateAdminAccount)
	}

	// Admin Identifier Management subgroup
	identifierGroup := adminRouter.Group("identifiers")
	{
		identifierGroup.Use(
			middleware.AdminAuthMiddleware(repos.AdminAccountRepo),
		)
		identifierGroup.Use(
			middleware.NewXHeaderValidationMiddleware(repos.TenantRepo).Middleware(),
		)
		identifierGroup.POST("/check", adminHandler.CheckIdentifier)
	}

	// Admin Tenant Management subgroup
	tenantRouter := adminRouter.Group("tenants")
	{
		tenantRouter.Use(middleware.AdminAuthMiddleware(repos.AdminAccountRepo))
		tenantRouter.GET("/", adminHandler.ListTenants)
		tenantRouter.GET("/:id", adminHandler.GetTenant)
		tenantRouter.POST("/", adminHandler.CreateTenant)
		tenantRouter.PUT("/:id", adminHandler.UpdateTenant)
		tenantRouter.DELETE("/:id", adminHandler.DeleteTenant)
	}

	// SECTION: Permission routes
	permissionHandler := handlers.NewPermissionHandler(ucases.PermissionUCase)
	permissionRouter := v1.Group("permissions")
	permissionRouter.Use(middleware.NewXHeaderValidationMiddleware(repos.TenantRepo).Middleware())
	{
		permissionRouter.POST("/self-check", authMiddleware.RequireAuth(), permissionHandler.SelfCheckPermission)
		permissionRouter.POST("/check", permissionHandler.CheckPermission)
		permissionRouter.POST("/relation-tuples", authMiddleware.RequireAuth(), permissionHandler.CreateRelationTuple)
		permissionRouter.POST("/delegate", authMiddleware.RequireAuth(), permissionHandler.DelegateAccess)
	}

	// SECTION: users
	userRouter := v1.Group("users")
	userRouter.Use(
		middleware.NewXHeaderValidationMiddleware(repos.TenantRepo).Middleware(),
	)
	userHandler := handlers.NewIdentityUserHandler(ucases.IdentityUserUCase)
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
		authMiddleware.RequireAuth(),
		userHandler.Logout,
	)

	userRouter.GET(
		"/me",
		authMiddleware.RequireAuth(),
		userHandler.Me,
	)

	userRouter.POST(
		"/me/add-identifier",
		authMiddleware.RequireAuth(),
		userHandler.AddIdentifier,
	)

	userRouter.POST(
		"/me/update-identifier",
		authMiddleware.RequireAuth(),
		userHandler.UpdateIdentifier,
	)

	// SECTION: Courier (OTP delivery) routes
	courierHandler := handlers.NewCourierHandler(ucases.CourierUCase)
	courierRouter := v1.Group("courier")

	courierRouter.POST("/messages", courierHandler.ReceiveCourierMessageHandler)

	courierRouter.GET(
		"/available-channels",
		middleware.NewXHeaderValidationMiddleware(repos.TenantRepo).Middleware(),
		courierHandler.GetAvailableChannelsHandler,
	)

	courierRouter.POST(
		"/choose-channel",
		middleware.NewXHeaderValidationMiddleware(repos.TenantRepo).Middleware(),
		courierHandler.ChooseChannelHandler,
	)
}
