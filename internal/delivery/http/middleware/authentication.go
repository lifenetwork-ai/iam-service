package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/constants"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// AuthMiddleware handles the complete authentication flow
type AuthMiddleware struct {
	identityUseCase interfaces.IdentityUserUseCase
}

// NewAuthMiddleware creates a new auth middleware with dependencies injected
func NewAuthMiddleware(identityUseCase interfaces.IdentityUserUseCase) *AuthMiddleware {
	return &AuthMiddleware{
		identityUseCase: identityUseCase,
	}
}

// validateAuthorizationHeader extracts and validates the authorization header format
func (am *AuthMiddleware) validateAuthorizationHeader(ctx *gin.Context) (bool, string) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		httpresponse.Error(
			ctx,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Authorization header is required",
			[]map[string]string{{
				"field": "Authorization",
				"error": "Authorization header is required",
			}},
		)
		return false, ""
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || (tokenParts[0] != "Bearer" && tokenParts[0] != "Token") {
		httpresponse.Error(
			ctx,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Invalid authorization header format",
			[]map[string]string{{
				"field": "Authorization",
				"error": "Invalid authorization header format. Expected 'Bearer <token>' or 'Token <token>'",
			}},
		)
		return false, ""
	}

	return true, strings.TrimSpace(tokenParts[1])
}

// RequireAuth is a complete authentication middleware
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Extract token from header
		isValid, token := am.validateAuthorizationHeader(ctx)
		if !isValid {
			ctx.Abort()
			return
		}

		// Get tenant from context (assuming this is set by another middleware)
		tenant, err := GetTenantFromContext(ctx)
		if err != nil {
			logger.GetLogger().Errorf("Failed to get tenant: %v", err)
			httpresponse.Error(
				ctx,
				http.StatusBadRequest,
				"MSG_INVALID_TENANT",
				"Invalid tenant",
				err,
			)
			ctx.Abort()
			return
		}

		// Set session token into request.Context()
		ctxWithToken := context.WithValue(ctx.Request.Context(), constants.SessionTokenKey, token)
		ctx.Request = ctx.Request.WithContext(ctxWithToken)

		// Validate token and get user
		user, ucaseErr := am.identityUseCase.Profile(ctx.Request.Context(), tenant.ID)
		if ucaseErr != nil {
			logger.GetLogger().Errorf("Token validation failed: %v", ucaseErr)
			httpresponse.Error(
				ctx,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid or expired token",
				ucaseErr,
			)
			ctx.Abort()
			return
		}

		// Set user and token in context for downstream handlers
		ctx.Set(string(constants.SessionTokenKey), token)
		ctx.Set(string(constants.UserContextKey), user)

		ctx.Next()
	}
}

// OptionalAuth middleware that doesn't require authentication but sets user context if token is present
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.Next()
			return
		}

		isValid, token := am.validateAuthorizationHeader(ctx)
		if !isValid {
			ctx.Next()
			return
		}

		tenant, err := GetTenantFromContext(ctx)
		if err != nil {
			ctx.Next()
			return
		}

		user, ucaseErr := am.identityUseCase.Profile(ctx, tenant.ID)
		if ucaseErr != nil {
			// Log but don't fail for optional auth
			logger.GetLogger().Warnf("Optional auth failed: %v", ucaseErr)
			ctx.Next()
			return
		}

		ctx.Set(string(constants.SessionTokenKey), token)
		ctx.Set(string(constants.UserContextKey), user)

		ctx.Next()
	}
}

// Helper functions to extract data from context
func GetUserFromContext(ctx *gin.Context) (*domain.UserIdentity, error) {
	user, exists := ctx.Get(string(constants.UserContextKey))
	if !exists {
		return nil, errors.New("user not found in context")
	}

	userObj, ok := user.(*domain.UserIdentity)
	if !ok {
		return nil, errors.New("invalid user type in context")
	}

	return userObj, nil
}

func GetTokenFromContext(ctx *gin.Context) (string, error) {
	token, exists := ctx.Get(string(constants.SessionTokenKey))
	if !exists {
		return "", errors.New("token not found in context")
	}

	tokenStr, ok := token.(string)
	if !ok {
		return "", errors.New("invalid token type in context")
	}

	return tokenStr, nil
}
