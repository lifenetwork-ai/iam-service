// middleware/logger.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// XHeaderValidationMiddleware returns a gin middleware for HTTP request logging
func XHeaderValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ignore Swagger requests
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Next() // Skip check headers for Swagger routes
			return
		}
		
		organizationId := c.GetHeader("X-Organization-Id")
		if organizationId == "" {
			c.AbortWithStatusJSON(
				http.StatusPreconditionRequired,
				gin.H{
					"code":    "MSG_MISSING_ORGANIZATION_ID_HEADER",
					"message": "Missing X-Organization-Id header",
					"details": []interface{}{
						map[string]string{
							"field": "X-Organization-Id",
							"error": "X-Organization-Id header is required",
						},
					},
				},
			)
			return
		}

		// Validate organization ID

		// Process request
		c.Set("organizationId", organizationId)
		c.Next()
	}
}
