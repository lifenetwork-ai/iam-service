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
	iamUCase interfaces.IAMUCase,
	fileInfoUCase interfaces.FileInfoUCase,
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
	appRouterAccount.Use(middleware.ValidateBearerToken())
	accountHandler := handlers.NewAccountHandler(iamUCase, accountUCase, authUCase)
	appRouterAccount.GET(
		"/me",
		middleware.RequiredRoles(
			authUCase,
			constants.Admin.String(),
			constants.User.String(),
			constants.DataOwner.String(),
			constants.DataUtilizer.String(),
			constants.Validator.String(),
		),
		accountHandler.GetCurrentUser,
	)
	appRouterAccount.PUT(
		"/role",
		middleware.RequiredRoles(
			authUCase,
			constants.User.String(),
			constants.DataOwner.String(),
		),
		accountHandler.UpdateAccountRole,
	)
	appRouterAccount.PUT(
		"/api-key",
		middleware.RequiredRoles(authUCase, constants.User.String()),
		accountHandler.UpdateAPIKey,
	)

	// SECTION: validator
	appRouterValidator := v1.Group("validators")
	appRouterValidator.Use(middleware.ValidateBearerToken())
	appRouterValidator.GET(
		"/active",
		middleware.RequiredRoles(authUCase, constants.DataOwner.String()),
		accountHandler.GetActiveValidators,
	)

	// SECTION: IAM
	appRouterIAM := v1.Group("iam")
	appRouterIAM.Use(
		middleware.ValidateBearerToken(),
		middleware.RequiredRoles(authUCase, constants.Admin.String()),
	)
	iamHandler := handlers.NewIAMHandler(iamUCase, authUCase)
	appRouterIAM.POST(
		"/policies",
		iamHandler.CreatePolicy,
	)
	appRouterIAM.GET(
		"/policies",
		iamHandler.GetPoliciesWithPermissions,
	)
	appRouterIAM.POST(
		"/policies/permissions",
		iamHandler.AssignPermissionToPolicy,
	)
	appRouterIAM.POST(
		"/accounts/:accountID/policies",
		iamHandler.AssignPolicyToAccount,
	)

	// SECTION: data access
	appRouterDataAccess := v1.Group("data-access")
	dataAccessHandler := handlers.NewDataAccessHandler(dataAccessUCase, authUCase, accountUCase)
	appRouterDataAccess.GET("/", middleware.ValidateBearerToken(), dataAccessHandler.GetDataAccessRequests)
	appRouterDataAccess.PUT("/:requestID/reject", middleware.ValidateBearerToken(), dataAccessHandler.RejectRequest)
	appRouterDataAccess.PUT("/:requestID/approve", middleware.ValidateBearerToken(), dataAccessHandler.ApproveRequest)
	appRouterDataAccess.GET("/:requesterAccountID", middleware.ValidateBearerToken(), dataAccessHandler.GetAccessRequest)

	// SECTION: notifications
	appRouterNotifications := v1.Group("notifications")
	appRouterNotifications.Use(middleware.ValidateBearerToken())
	notificationHandler := handlers.NewNotificationHandler(authUCase, accountUCase, dataAccessUCase, fileInfoUCase)
	appRouterNotifications.POST(
		"/data-upload",
		middleware.RequiredRoles(authUCase, constants.DataOwner.String()),
		notificationHandler.DataUploadWebhook,
	)
}
