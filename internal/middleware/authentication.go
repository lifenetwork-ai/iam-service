package middleware

import (
	"github.com/gin-gonic/gin"
)

// RequestAuthenticationMiddleware returns a gin middleware for HTTP request logging
func RequestAuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()
	}
}
