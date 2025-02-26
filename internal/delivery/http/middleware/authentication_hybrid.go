package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	cachingTypes "github.com/genefriendway/human-network-iam/infrastructures/caching/types"
	repositories "github.com/genefriendway/human-network-iam/internal/adapters/repositories"
	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
	httpresponse "github.com/genefriendway/human-network-iam/packages/http/response"
	"github.com/genefriendway/human-network-iam/packages/logger"
	"github.com/genefriendway/human-network-iam/wire/providers"
)

// RequestHybridAuthenticationMiddleware returns a gin middleware for HTTP request authentication with supporting 3rd party service
func RequestHybridAuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Dependency injection
		dbConnection := providers.ProvideDBConnection()
		cacheRepo := providers.ProvideCacheRepository(c)
		lifeAIService := providers.ProvideLifeAIService()

		// Get the Bearer token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Authorization header is required",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Authorization header is required",
				}},
			)
			return
		}

		// Check if the token is a Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Authorization header is required",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid authorization header format",
				}},
			)
			return
		}

		// Query Redis to find profile with key is tokenMd5
		token := tokenParts[1]
		var requester *entities.IdentityUser = nil
		cacheKey := &cachingTypes.Keyer{
			Raw: fmt.Sprintf("middleware_%x", sha256.Sum256([]byte(token))),
		}

		var cacheRequester interface{}
		err := cacheRepo.RetrieveItem(cacheKey, &cacheRequester)
		if err == nil {
			if user, ok := cacheRequester.(entities.IdentityUser); ok {
				requester = &user
			}

			if requester != nil {
				c.Set("requesterId", requester.ID)
				c.Set("requester", requester)
				c.Next()
				return
			}
		}

		// Try to get user from database if not found in cache
		// Check if request is using LifeAI token
		profile, err := lifeAIService.GetProfile(c, authHeader)
		if err != nil {
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid token",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid token",
				}},
			)
			return
		}

		if profile == nil || profile.ID == "" || (profile.Email == "" && profile.Phone == "") {
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid token",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid LifeAI profile data",
				}},
			)
			return
		}

		// Check profile is exist or not -> if not exist, create new user
		userRepo := repositories.NewIdentityUserRepository(dbConnection, cacheRepo)
		userByID, err := userRepo.FindByLifeAIID(c, profile.ID)
		if err != nil {
			httpresponse.Error(
				c,
				http.StatusInternalServerError,
				"INTERNAL_SERVER_ERROR",
				"Failed to get user with LifeAI profile by ID",
				[]map[string]string{{
					"field": "Authorization",
					"error": err.Error(),
				}},
			)
			return
		}

		requester = userByID
		if requester == nil && profile.Phone != "" {
			userByPhone, err := userRepo.FindByPhone(c, profile.Phone)
			if err != nil {
				httpresponse.Error(
					c,
					http.StatusInternalServerError,
					"INTERNAL_SERVER_ERROR",
					"Failed to get user with LifeAI profile by phone",
					[]map[string]string{{
						"field": "Authorization",
						"error": err.Error(),
					}},
				)
				return
			}
			requester = userByPhone
		}

		if requester == nil && profile.Email != "" {
			userByEmail, err := userRepo.FindByEmail(c, profile.Email)
			if err != nil {
				httpresponse.Error(
					c,
					http.StatusInternalServerError,
					"INTERNAL_SERVER_ERROR",
					"Failed to get user with LifeAI profile by email",
					[]map[string]string{{
						"field": "Authorization",
						"error": err.Error(),
					}},
				)
				return
			}
			requester = userByEmail
		}

		if requester == nil {
			// Create new user
			username := func() string {
				if profile.Phone != "" {
					return profile.Phone
				}
				return profile.Email
			}()

			newIdentityUser := &entities.IdentityUser{
				UserName: username,
				Email:    profile.Email,
				Phone:    profile.Phone,
				LifeAIID: profile.ID,
			}

			if err := userRepo.Create(c, newIdentityUser); err != nil {
				httpresponse.Error(
					c,
					http.StatusInternalServerError,
					"INTERNAL_SERVER_ERROR",
					"Failed to create user with LifeAI profile",
					[]map[string]string{{
						"field": "Authorization",
						"error": err.Error(),
					}},
				)
				return
			}

			requester = newIdentityUser
		}

		if requester == nil {
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid token",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid token",
				}},
			)
			return
		}

		// Cache the user to memory cache
		if err = cacheRepo.SaveItem(cacheKey, *requester, 30*time.Minute); err != nil {
			logger.GetLogger().Errorf("Failed to cache user: %v", err)
		}

		// Set the user in the context
		c.Set("requesterId", requester.ID)
		c.Set("requester", requester)
		c.Next()
	}
}
