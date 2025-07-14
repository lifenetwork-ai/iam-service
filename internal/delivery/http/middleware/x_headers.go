// middleware/logger.go
package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	cachingTypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/lifenetwork-ai/iam-service/internal/wire/instances"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// XHeaderValidationMiddleware returns a gin middleware for HTTP request checking X-* headers
func XHeaderValidationMiddleware() gin.HandlerFunc {
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

		// Dependency injection
		dbConnection := instances.DBConnectionInstance()
		cacheRepo := instances.CacheRepositoryInstance(c)
		tenantRepo := repositories.NewTenantRepository(dbConnection)

		// Query Redis to find profile with key is tokenMd5
		var tenant *entities.Tenant
		cacheKey := &cachingTypes.Keyer{
			Raw: tenantId,
		}

		err := cacheRepo.RetrieveItem(cacheKey, &tenant)
		if err != nil {
			logger.GetLogger().Errorf("Failed to retrieve tenant from cache: %v", err)
		}
		logger.GetLogger().Infof("tenant: %v", tenant)
		if tenant != nil {
			c.Set(string(TenantIDKey), tenant.ID)
			c.Set(string(TenantKey), tenant)
			c.Next()
			return
		}

		tenant, err = tenantRepo.GetByID(uuid.MustParse(tenantId))
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{
					"code":    "MSG_FAILED_TO_GET_TENANT",
					"message": "Failed to get tenant",
					"details": []interface{}{
						map[string]string{
							"error": err.Error(),
						},
					},
				},
			)
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

		// Cache the user to memory cache
		if err = cacheRepo.SaveItem(cacheKey, *tenant, 30*time.Minute); err != nil {
			logger.GetLogger().Errorf("Failed to cache tenant: %v", err)
		}

		logger.GetLogger().Infof("tenant: %v", tenant)
		c.Set(string(TenantKey), tenant)
		c.Next()
	}
}
