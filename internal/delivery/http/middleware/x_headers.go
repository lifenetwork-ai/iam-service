// middleware/logger.go
package middleware

import (
	"net/http"
	"strings"
	"time"

	cachingTypes "github.com/genefriendway/human-network-iam/infrastructures/caching/types"
	repositories "github.com/genefriendway/human-network-iam/internal/adapters/repositories"
	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
	"github.com/genefriendway/human-network-iam/packages/logger"
	"github.com/genefriendway/human-network-iam/wire/providers"
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

		// Dependency injection
		dbConnection := providers.ProvideDBConnection()
		cacheRepo := providers.ProvideCacheRepository(c)
		organizationRepo := repositories.NewIdentityOrganizationRepository(dbConnection, cacheRepo)

		// Query Redis to find profile with key is tokenMd5
		var organization *entities.IdentityOrganization = nil
		cacheKey := &cachingTypes.Keyer{
			Raw: organizationId,
		}

		var cacheRequester interface{}
		err := cacheRepo.RetrieveItem(cacheKey, &cacheRequester)
		if err == nil {
			if org, ok := cacheRequester.(entities.IdentityOrganization); ok {
				organization = &org
			}

			if organization != nil {
				c.Set("requesterId", organization.ID)
				c.Set("organization", organization)
				c.Next()
				return
			}
		}

		organization, err = organizationRepo.GetByID(c, organizationId)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{
					"code":    "MSG_FAILED_TO_GET_ORGANIZATION",
					"message": "Failed to get organization",
					"details": []interface{}{
						map[string]string{
							"error": err.Error(),
						},
					},
				},
			)
			return
		}

		if organization == nil {
			c.AbortWithStatusJSON(
				http.StatusNotFound,
				gin.H{
					"code":    "MSG_ORGANIZATION_NOT_FOUND",
					"message": "Organization not found",
					"details": []interface{}{
						map[string]string{
							"field": "organization_id",
							"error": "Organization not found",
						},
					},
				},
			)
			return
		}

		// Cache the user to memory cache
		if err = cacheRepo.SaveItem(cacheKey, *organization, 30*time.Minute); err != nil {
			logger.GetLogger().Errorf("Failed to cache organization: %v", err)
		}

		c.Set("requesterId", organization.ID)
		c.Set("organization", organization)
		c.Next()
	}
}
