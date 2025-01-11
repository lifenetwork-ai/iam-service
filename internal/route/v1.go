package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/conf"
	"github.com/genefriendway/human-network-auth/constants"
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
	policyUCase interfaces.PolicyUCase,
) {
	v1 := r.Group("/api/v1")

	// SECTION: auth
	appRouterAuth := v1.Group("auth")
	authHandler := handlers.NewAuthHandler(authUCase)
	appRouterAuth.POST("/register", authHandler.Register)
	appRouterAuth.POST("/login", authHandler.Login)
	appRouterAuth.POST("/logout", authHandler.Logout)
	appRouterAuth.POST("/refresh-tokens", authHandler.RefreshTokens)

	// SECTION: account
	appRouterAccount := v1.Group("account")
	accountHandler := handlers.NewAccountHandler(accountUCase, authUCase)
	appRouterAccount.GET(
		"/me",
		middleware.ValidateBearerToken(),
		middleware.RequiredRoles(
			authUCase,
			constants.Admin.String(),
			constants.DataOwner.String(),
			constants.DataUtilizer.String(),
			constants.Validator.String(),
		),
		accountHandler.GetCurrentUser,
	)
	appRouterAccount.PUT(
		"/role",
		middleware.ValidateBearerToken(),
		middleware.RequiredRoles(authUCase, constants.DataOwner.String()),
		accountHandler.UpdateAccountRole,
	)

	// SECTION: validator
	appRouterValidator := v1.Group("validators")
	appRouterValidator.GET(
		"/active",
		middleware.ValidateBearerToken(),
		middleware.RequiredRoles(authUCase, constants.DataOwner.String()),
		accountHandler.GetActiveValidators,
	)

	// SECTION: IAM
	appRouterIAM := v1.Group("iam")
	policyHandler := handlers.NewIAMHandler(policyUCase)
	appRouterIAM.POST(
		"/policies",
		middleware.ValidateBearerToken(),
		middleware.RequiredRoles(authUCase, constants.Admin.String()),
		policyHandler.CreatePolicy,
	)

	// SECTION: data access
	appRouterDataAccess := v1.Group("data-access")
	dataAccessHandler := handlers.NewDataAccessHandler(dataAccessUCase, authUCase)
	appRouterDataAccess.POST("/", middleware.ValidateBearerToken(), dataAccessHandler.CreateDataAccessRequest)
	appRouterDataAccess.GET("/", middleware.ValidateBearerToken(), dataAccessHandler.GetDataAccessRequests)
	appRouterDataAccess.PUT("/:requestID/reject", middleware.ValidateBearerToken(), dataAccessHandler.RejectRequest)
	appRouterDataAccess.PUT("/:requestID/approve", middleware.ValidateBearerToken(), dataAccessHandler.ApproveRequest)
	appRouterDataAccess.GET("/:requesterAccountID", middleware.ValidateBearerToken(), dataAccessHandler.GetAccessRequest)
}
