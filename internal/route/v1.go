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
	userUCase interfaces.IdentityUserUseCase,
) {
	v1 := r.Group("/api/v1")

	// SECTION: organizations
	organizationRouter := v1.Group("organizations")
	organizationHandler := handlers.NewIdentityOrganizationHandler(organizationUCase)
	organizationRouter.GET("/", organizationHandler.GetOrganizations)
	organizationRouter.GET("/:organization_id", organizationHandler.GetDetail)
	organizationRouter.POST("/", organizationHandler.CreateOrganization)
	organizationRouter.PUT("/:organization_id", organizationHandler.UpdateOrganization)
	organizationRouter.DELETE("/:organization_id", organizationHandler.DeleteOrganization)

	// SECTION: organizations
	userRouter := v1.Group("users")
	userHandler := handlers.NewIdentityUserHandler(userUCase)
	userRouter.POST("/challenge-with-phone", userHandler.ChallengeWithPhone)
	userRouter.POST("/challenge-with-email", userHandler.ChallengeWithEmail)
	userRouter.POST("/challenge-verify", userHandler.ChallengeVerify)
	userRouter.POST("/login-with-google", userHandler.LoginWithGoogle)
	userRouter.POST("/login-with-facebook", userHandler.LoginWithFacebook)
	userRouter.POST("/login-with-apple", userHandler.LoginWithApple)
	// userRouter.POST("/register", userHandler.Register)
	userRouter.POST("/login", userHandler.Login)
	userRouter.GET("/me", userHandler.Me)
	userRouter.POST("/refresh-token", userHandler.RefreshToken)
	userRouter.POST("/logout", userHandler.Logout)
}
