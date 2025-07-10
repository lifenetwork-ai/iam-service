package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
)

func validateAuthorizationHeader(
	ctx *gin.Context,
) (bool, string, string) {
	// Get the Bearer token from the Authorization header
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
		return false, "", ""
	}

	// Check if the token is a Bearer token
	tokenParts := strings.Split(authHeader, " ")
	tokenPrefix := tokenParts[0]
	if len(tokenParts) != 2 || (tokenPrefix != "Bearer" && tokenPrefix != "Token") {
		httpresponse.Error(
			ctx,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Authorization header is required",
			[]map[string]string{{
				"field": "Authorization",
				"error": "Invalid authorization header format",
			}},
		)
		return false, "", ""
	}

	return true, authHeader, strings.TrimSpace(tokenParts[1])
}

// RequestAuthenticationMiddleware returns a gin middleware for HTTP request authentication
func RequestAuthenticationMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the token from the header
		isOK, _, token := validateAuthorizationHeader(ctx)
		if !isOK {
			return
		}

		// // Query Redis to find profile with key is tokenMd5
		// requester := cacheAuthentication(ctx, token)
		// if requester == nil {
		// 	requester = jwtAuthentication(ctx, token)
		// 	saveToCache(ctx, requester, token)
		// }

		// if requester == nil {
		// 	httpresponse.Error(
		// 		ctx,
		// 		http.StatusUnauthorized,
		// 		"UNAUTHORIZED",
		// 		"Invalid token",
		// 		[]map[string]string{{
		// 			"field": "Authorization",
		// 			"error": "Invalid token",
		// 		}},
		// 	)
		// 	return
		// }

		// Set the user in the context
		// ctx.Set("requesterId", requester.ID)
		// ctx.Set("requester", requester)
		ctx.Set("sessionToken", token)
		ctx.Next()
	}
}
