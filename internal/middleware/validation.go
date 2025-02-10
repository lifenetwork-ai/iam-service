package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
	httpresponse "github.com/genefriendway/human-network-iam/packages/http/response"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

// ValidateBearerToken is a middleware that validates the Bearer token in the Authorization header.
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

// RequiredRoles is a middleware that checks if the account has the required role.
func RequiredRoles(authUCase interfaces.AuthUCase, allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{})
	for _, role := range allowedRoles {
		roleSet[role] = struct{}{}
	}

	return func(ctx *gin.Context) {
		accountDTO, ok := getAuthenticatedAccount(ctx, authUCase)
		if !ok {
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

// CheckPermission is a middleware that checks if the account has permission to perform an action on a resource.
func CheckPermission(iamUCase interfaces.IAMUCase, authUCase interfaces.AuthUCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accountDTO, ok := getAuthenticatedAccount(ctx, authUCase)
		if !ok {
			return
		}

		// Extract resource and action from the request
		resource := ctx.Param("resource") // Assuming the resource is passed as a path parameter
		action := ctx.Param("action")     // Assuming the action is passed as a path parameter

		if resource == "" || action == "" {
			httpresponse.Error(ctx, http.StatusBadRequest, "Resource or action not specified", nil)
			return
		}

		// Check if the user has the required permission
		hasPermission, err := iamUCase.CheckPermission(accountDTO.ID, resource, action)
		if err != nil {
			logger.GetLogger().Errorf("Permission check failed: %v", err)
			httpresponse.Error(ctx, http.StatusInternalServerError, "Error checking permissions", err)
			return
		}

		if !hasPermission {
			httpresponse.Error(ctx, http.StatusForbidden, "Permission denied", nil)
			return
		}

		// Permission granted, continue to the next handler
		ctx.Next()
	}
}

// getAuthenticatedAccount is a helper function that retrieves the authenticated account from the context.
func getAuthenticatedAccount(ctx *gin.Context, authUCase interfaces.AuthUCase) (*dto.AccountDTO, bool) {
	// Retrieve the token from the context
	token, exists := ctx.Get("token")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Missing token in request", nil)
		return nil, false
	}

	// Validate the token and fetch the account
	accountDTO, err := authUCase.ValidateToken(token.(string))
	if err != nil {
		logger.GetLogger().Errorf("Token validation failed: %v", err)
		httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid or expired token", nil)
		return nil, false
	}

	// Check if the account is missing
	if accountDTO == nil {
		logger.GetLogger().Error("Account information missing or invalid")
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing or invalid", nil)
		return nil, false
	}

	return accountDTO, true
}

func RequiredAPIKey(accountUCase interfaces.AccountUCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Retrieve API key from the header
		apiKey := ctx.GetHeader("X-API-Key")
		if apiKey == "" {
			httpresponse.Error(ctx, http.StatusUnauthorized, "Missing API Key", nil)
			ctx.Abort()
			return
		}

		// Use the account use case to validate the API key
		account, err := accountUCase.FindAccountByAPIKey(apiKey)
		if err != nil {
			logger.GetLogger().Errorf("Error validating API key: %v", err)
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid API Key", nil)
			ctx.Abort()
			return
		}

		if account == nil {
			logger.GetLogger().Warn("API key does not match any account")
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid API Key", nil)
			ctx.Abort()
			return
		}

		// Add the account to the context for downstream handlers
		ctx.Set("account", account)
		ctx.Next()
	}
}

// APIKeyWithPermission is a middleware that validates the API key and checks if the account has the required permission.
func APIKeyWithPermission(
	accountUCase interfaces.AccountUCase,
	iamUCase interfaces.IAMUCase,
	resource string,
	action string,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Retrieve API key from the header
		apiKey := ctx.GetHeader("X-API-Key")
		if apiKey == "" {
			httpresponse.Error(ctx, http.StatusUnauthorized, "Missing API Key", nil)
			ctx.Abort()
			return
		}

		// Validate the API key and fetch the account
		account, err := accountUCase.FindAccountByAPIKey(apiKey)
		if err != nil {
			logger.GetLogger().Errorf("Error validating API key: %v", err)
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid API Key", nil)
			ctx.Abort()
			return
		}

		if account == nil {
			logger.GetLogger().Warn("API key does not match any account")
			httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid API Key", nil)
			ctx.Abort()
			return
		}

		// Check if the account has the required permission
		hasPermission, err := iamUCase.CheckPermission(account.ID, resource, action)
		if err != nil {
			logger.GetLogger().Errorf("Permission check failed: %v", err)
			httpresponse.Error(ctx, http.StatusInternalServerError, "Error checking permissions", err)
			ctx.Abort()
			return
		}

		if !hasPermission {
			httpresponse.Error(ctx, http.StatusForbidden, "Permission denied", nil)
			ctx.Abort()
			return
		}

		// Add account to the context for downstream handlers
		ctx.Set("account", account)
		ctx.Next()
	}
}
