// middleware/logger.go
package middleware

import (
	"net/http"

	httpresponse "github.com/genefriendway/human-network-iam/packages/http/response"
	"github.com/gin-gonic/gin"
)

// XHeaderValidationMiddleware returns a gin middleware for HTTP request logging
func XHeaderValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		organizationId := c.GetHeader("X-Organization-Id")
		if organizationId == "" {
			httpresponse.Error(
				c,
				http.StatusPreconditionRequired,
				"MSG_MISSING_ORGANIZATION_ID_HEADER",
				"Missing X-Organization-Id header",
				gin.H{
					"field": "X-Organization-Id",
					"error": "X-Organization-Id header is required",
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
