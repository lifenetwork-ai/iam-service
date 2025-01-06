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
	authUCase interfaces.AuthUCase,
	accountUCase interfaces.AccountUCase,
	dataAccessUCase interfaces.DataAccessUCase,
) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: auth
	authHandler := handlers.NewAuthHandler(authUCase)
	appRouter.POST("/auth/register", authHandler.Register)
	appRouter.POST("/auth/login", authHandler.Login)
	appRouter.POST("/auth/logout", authHandler.Logout)
	appRouter.POST("/auth/refresh-tokens", authHandler.RefreshTokens)

	// SECTION: account
	accountHandler := handlers.NewAccountHandler(accountUCase, authUCase)
	appRouter.GET("/account/me", middleware.ValidateBearerToken(), accountHandler.GetCurrentUser)
	appRouter.PUT("/account/role", middleware.ValidateBearerToken(), accountHandler.UpdateAccountRole)

	// SECTION: validator
	appRouter.GET("validators/active", middleware.ValidateBearerToken(), accountHandler.GetActiveValidators)

	// SECTION: data access
	dataAccessHandler := handlers.NewDataAccessHandler(dataAccessUCase, authUCase)
	appRouter.POST("/data-access", middleware.ValidateBearerToken(), dataAccessHandler.CreateDataAccessRequest)
	appRouter.GET("/data-access/pending", middleware.ValidateBearerToken(), dataAccessHandler.GetPendingDataAccessRequests)
	appRouter.PUT("/data-access/:requesterAccountID/reject", middleware.ValidateBearerToken(), dataAccessHandler.RejectRequest)
	appRouter.PUT("/data-access/:requesterAccountID/approve", middleware.ValidateBearerToken(), dataAccessHandler.ApproveRequest)
}
