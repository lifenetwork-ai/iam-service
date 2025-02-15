package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	httpresponse "github.com/genefriendway/human-network-iam/packages/http/response"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

// ValidateBearerToken is a middleware that validates the Bearer token in the Authorization header.
func ValidateBearerToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			logger.GetLogger().Error("Missing Authorization header")
			httpresponse.Error(
				ctx,
				http.StatusUnauthorized,
				"MSG_MISSING_AUTHORIZATION_HEADER",
				"Missing Authorization header",
				nil,
			)
			ctx.Abort()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			logger.GetLogger().Error("Invalid Authorization header format")
			httpresponse.Error(
				ctx,
				http.StatusUnauthorized,
				"MSG_INVALID_AUTHORIZATION_HEADER",
				"Invalid Authorization header format",
				nil,
			)
			ctx.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, bearerPrefix)
		ctx.Set("token", token)
		ctx.Next()
	}
}
