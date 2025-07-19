// middleware/logger.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	repotypes "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
)

type XHeaderValidationMiddleware struct {
	tenantRepo repotypes.TenantRepository
}

func NewXHeaderValidationMiddleware(tenantRepo repotypes.TenantRepository) *XHeaderValidationMiddleware {
	return &XHeaderValidationMiddleware{
		tenantRepo: tenantRepo,
	}
}

// XHeaderValidationMiddleware returns a gin middleware for HTTP request checking X-* headers
func (m *XHeaderValidationMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ignore Swagger requests
		// if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
		// 	c.Next() // Skip check headers for Swagger routes
		// 	return
		// }

		tenantId := c.GetHeader("X-Tenant-Id")
		if tenantId == "" {
			c.AbortWithStatusJSON(
				http.StatusPreconditionRequired,
				gin.H{
					"code":    "MSG_MISSING_TENANT_ID_HEADER",
					"message": "Missing X-Tenant-Id header",
					"details": []interface{}{
						map[string]string{
							"field": "X-Tenant-Id",
							"error": "X-Tenant-Id header is required",
						},
					},
				},
			)
			return
		}

		// Query Redis to find profile with key is tokenMd5
		tenantIdUUID, err := uuid.Parse(tenantId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":    "MSG_INVALID_TENANT_ID",
				"message": "Invalid tenant ID",
			})
			return
		}

		tenant, err := m.tenantRepo.GetByID(tenantIdUUID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"code":    "MSG_TENANT_NOT_FOUND",
				"message": "Tenant not found",
			})
			return
		}

		if tenant == nil {
			c.AbortWithStatusJSON(
				http.StatusNotFound,
				gin.H{
					"code":    "MSG_TENANT_NOT_FOUND",
					"message": "Tenant not found",
					"details": []interface{}{
						map[string]string{
							"field": "tenant_id",
							"error": "Tenant not found",
						},
					},
				},
			)
			return
		}

		c.Set(string(constants.TenantKey), tenant)
		c.Next()
	}
}
