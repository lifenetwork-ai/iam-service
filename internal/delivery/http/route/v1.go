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
	userUCase interfaces.IdentityUserUseCase,
) {
	v1 := r.Group("/api/v1")

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
