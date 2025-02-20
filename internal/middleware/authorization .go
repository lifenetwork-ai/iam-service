package middleware

import (
	"github.com/gin-gonic/gin"
)

// RequestAuthorizationMiddleware returns a gin middleware for HTTP request logging
func RequestAuthorizationMiddleware(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()
	}
}
