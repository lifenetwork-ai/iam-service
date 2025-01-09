package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/internal/interfaces"
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

func RequiredRoles(authUCase interfaces.AuthUCase, allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{})
	for _, role := range allowedRoles {
		roleSet[role] = struct{}{}
	}

	return func(ctx *gin.Context) {
		// Retrieve the token from the context
		token, exists := ctx.Get("token")
		if !exists {
			httpresponse.Error(ctx, http.StatusUnauthorized, "Missing token in request", nil)
			return
		}

		// Validate the token and fetch the account
		accountDTO, err := authUCase.ValidateToken(token.(string))
		if err != nil {
			logger.GetLogger().Errorf("Token validation failed: %v", err)
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid or expired token", nil)
			return
		}

		// Check if the account is missing
		if accountDTO == nil {
			logger.GetLogger().Error("Account information missing or invalid")
			httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing or invalid", nil)
			return
		}

		// Check if the account role matches one of the allowed roles
		if _, allowed := roleSet[accountDTO.Role]; !allowed {
			logger.GetLogger().Warnf("Access denied for account ID: %s, role: %s", accountDTO.ID, accountDTO.Role)
			httpresponse.Error(ctx, http.StatusForbidden, "Access denied: insufficient permissions", nil)
			return
		}

		// Set account in context and proceed
		ctx.Set("account", accountDTO)
		ctx.Next()
	}
}
