package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	httpresponse "github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
)

func ValidateBearerToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			logger.GetLogger().Error("Missing Authorization header")
			httpresponse.Error(ctx, http.StatusUnauthorized, "Missing Authorization header", nil)
			ctx.Abort()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			logger.GetLogger().Error("Invalid Authorization header format")
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid Authorization header format", nil)
			ctx.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, bearerPrefix)
		ctx.Set("token", token)
		ctx.Next()
	}
}
