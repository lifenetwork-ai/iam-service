package middleware

import (
	"github.com/gin-gonic/gin"
)

// RequestAuthorizationMiddleware returns a gin middleware for HTTP request authorization
func RequestAuthorizationMiddleware(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()
	}
}
