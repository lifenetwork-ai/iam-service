// middleware/logger.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// XHeaderValidationMiddleware returns a gin middleware for HTTP request logging
func XHeaderValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		c.Set("organization_id", organizationId)
		c.Next()
	}
}
