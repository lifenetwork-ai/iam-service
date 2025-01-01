package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/conf"
	"github.com/genefriendway/human-network-auth/internal/adapters/handlers"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	db *gorm.DB,
	authUCase interfaces.AuthUCase,
) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: auth
	authHandler := handlers.NewAuthHandler(authUCase)
	appRouter.POST("/auth/register", authHandler.Register)
	appRouter.POST("/auth/login", authHandler.Login)
	appRouter.POST("/auth/logout", authHandler.Logout)
	appRouter.POST("/auth/refresh-tokens", authHandler.RefreshTokens)
	appRouter.GET("/validate-token", authHandler.ValidateToken)
}
