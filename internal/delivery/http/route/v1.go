package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/handlers"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	middleware "github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	db *gorm.DB,
	userUCase interfaces.IdentityUserUseCase,
	adminUCase interfaces.AdminUseCase,
) {
	v1 := r.Group("/api/v1")

	// SECTION: Admin routes
	adminRepo := repositories.NewAdminAccountRepository(db)
	adminRouter := v1.Group("admin")

	// Admin Tenant Management subgroup
	adminHandler := handlers.NewAdminHandler(adminUCase)
	accountRouter := adminRouter.Group("accounts")
	{
		accountRouter.Use(
			middleware.RootAuthMiddleware(),
		)
		accountRouter.POST("/", adminHandler.CreateAdminAccount)
	}

	tenantRouter := adminRouter.Group("tenants")
	{
		tenantRouter.Use(middleware.AdminAuthMiddleware(adminRepo))
		tenantRouter.GET("/", adminHandler.ListTenants)
		tenantRouter.GET("/:id", adminHandler.GetTenant)
		tenantRouter.POST("/", adminHandler.CreateTenant)
		tenantRouter.PUT("/:id", adminHandler.UpdateTenant)
		tenantRouter.DELETE("/:id", adminHandler.DeleteTenant)
		tenantRouter.PUT("/:id/status", adminHandler.UpdateTenantStatus)
	}

	// SECTION: users
	userRouter := v1.Group("users")
	userRouter.Use(
		middleware.XHeaderValidationMiddleware(),
	)
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
