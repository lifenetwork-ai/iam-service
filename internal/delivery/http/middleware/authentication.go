package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	cachingtypes "github.com/genefriendway/human-network-iam/infrastructures/caching/types"
	repositories "github.com/genefriendway/human-network-iam/internal/adapters/repositories"
	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
	httpresponse "github.com/genefriendway/human-network-iam/packages/http/response"
	"github.com/genefriendway/human-network-iam/packages/logger"
	"github.com/genefriendway/human-network-iam/wire/providers"
)

func validateAuthorizationHeader(
	ctx *gin.Context,
) (bool, string, string) {
	// Get the Bearer token from the Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		httpresponse.Error(
			ctx,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Authorization header is required",
			[]map[string]string{{
				"field": "Authorization",
				"error": "Authorization header is required",
			}},
		)
		return false, "", ""
	}

	// Check if the token is a Bearer token
	tokenParts := strings.Split(authHeader, " ")
	tokenPrefix := tokenParts[0]
	if len(tokenParts) != 2 || (tokenPrefix != "Bearer" && tokenPrefix != "Token") {
		httpresponse.Error(
			ctx,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Authorization header is required",
			[]map[string]string{{
				"field": "Authorization",
				"error": "Invalid authorization header format",
			}},
		)
		return false, "", ""
	}

	return true, authHeader, strings.TrimSpace(tokenParts[1])
}

func cacheAuthentication(
	ctx *gin.Context,
	token string,
) *entities.IdentityUser {
	// Dependency injection
	cacheRepo := providers.ProvideCacheRepository(ctx)

	// Query Redis to find profile with key is tokenMd5
	var requester *entities.IdentityUser = nil
	cacheKey := &cachingtypes.Keyer{
		Raw: fmt.Sprintf("middleware_%x", sha256.Sum256([]byte(token))),
	}

	var cacheRequester interface{}
	err := cacheRepo.RetrieveItem(cacheKey, &cacheRequester)
	if err == nil {
		if user, ok := cacheRequester.(entities.IdentityUser); ok {
			requester = &user
		}
	}

	return requester
}

func saveToCache(
	ctx *gin.Context,
	requester *entities.IdentityUser,
	token string,
) {
	if requester == nil {
		return
	}

	// Dependency injection
	cacheRepo := providers.ProvideCacheRepository(ctx)

	// Cache the user to memory cache
	cacheKey := &cachingtypes.Keyer{
		Raw: fmt.Sprintf("middleware_%x", sha256.Sum256([]byte(token))),
	}

	if err := cacheRepo.SaveItem(cacheKey, *requester, 30*time.Minute); err != nil {
		logger.GetLogger().Errorf("Failed to cache user: %v", err)
	}
}

func lifeAIAuthentication(
	ctx *gin.Context,
	authHeader string,
) *entities.IdentityUser {
	// Check if request is using LifeAI token
	// Dependency injection
	dbConnection := providers.ProvideDBConnection()
	cacheRepo := providers.ProvideCacheRepository(ctx)
	lifeAIService := providers.ProvideLifeAIService()

	// Try to get user from database if not found in cache
	profile, err := lifeAIService.GetProfile(ctx, authHeader)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get profile from LifeAI: %v", err)
		return nil
	}

	if profile == nil || profile.ID == "" || (profile.Email == "" && profile.Phone == "") {
		logger.GetLogger().Errorf("Invalid profile from LifeAI: %v", profile)
		return nil
	}

	// Check profile is exist or not -> if not exist, create new user
	userRepo := repositories.NewIdentityUserRepository(dbConnection, cacheRepo)
	userByID, err := userRepo.FindByLifeAIID(ctx, profile.ID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get user with LifeAI profile by ID: %v", err)
		return nil
	}

	requester := userByID
	if requester == nil && profile.Phone != "" {
		userByPhone, err := userRepo.FindByPhone(ctx, profile.Phone)
		if err != nil {
			logger.GetLogger().Errorf("Failed to get user with LifeAI profile by phone: %v", err)
			return nil
		}
		requester = userByPhone
	}

	if requester == nil && profile.Email != "" {
		userByEmail, err := userRepo.FindByEmail(ctx, profile.Email)
		if err != nil {
			logger.GetLogger().Errorf("Failed to get user with LifeAI profile by email: %v", err)
			return nil
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

		if err := userRepo.Create(ctx, newIdentityUser); err != nil {
			logger.GetLogger().Errorf("Failed to create new user: %v", err)
			return nil
		}
		requester = newIdentityUser
	}

	return requester
}

func selfAuthentication(
	ctx *gin.Context,
	token string,
) *entities.IdentityUser {
	// Check if request is using self token
	// Dependency injection
	dbConnection := providers.ProvideDBConnection()
	cacheRepo := providers.ProvideCacheRepository(ctx)
	jwtService := providers.ProvideJWTService()

	// Try to get user from database if not found in cache
	sessionRepo := repositories.NewAccessSessionRepository(dbConnection, cacheRepo)
	session, err := sessionRepo.FindByAccessToken(ctx, token)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get session by access token: %v", err)
		return nil
	}

	if session == nil {
		logger.GetLogger().Errorf("Invalid session by access token: %v", session)
		return nil
	}

	// Validate token
	jwtClaims, err := jwtService.ValidateToken(ctx, token)
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate token: %v", err)
		return nil
	}

	if jwtClaims == nil {
		logger.GetLogger().Errorf("Invalid token claims: %v", jwtClaims)
		return nil
	}

	tokenOrganizationId := jwtClaims.OrganizationId
	requestOrganizationId := ctx.Value("organizationId").(string)
	if tokenOrganizationId != requestOrganizationId {
		logger.GetLogger().Errorf("Organization ID in token is not match with request: %v", jwtClaims)
		return nil
	}

	// Get user from database
	userRepo := repositories.NewIdentityUserRepository(dbConnection, cacheRepo)
	userByID, err := userRepo.FindByID(ctx, jwtClaims.UserId)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get user by ID: %v", err)
		return nil
	}

	if userByID == nil {
		logger.GetLogger().Errorf("User not found by ID: %v", jwtClaims.UserId)
		return nil
	}

	return userByID
}

// RequestAuthenticationMiddleware returns a gin middleware for HTTP request authentication
func RequestAuthenticationMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the token from the header
		isOK, _, token := validateAuthorizationHeader(ctx)
		if !isOK {
			return
		}

		// Query Redis to find profile with key is tokenMd5
		requester := cacheAuthentication(ctx, token)
		if requester == nil {
			requester = selfAuthentication(ctx, token)

			saveToCache(ctx, requester, token)
		}

		if requester == nil {
			httpresponse.Error(
				ctx,
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

		// Set the user in the context
		ctx.Set("requesterId", requester.ID)
		ctx.Set("requester", requester)
		ctx.Next()
	}
}

// RequestHybridAuthenticationMiddleware returns a gin middleware for HTTP request authentication with supporting 3rd party service
func RequestHybridAuthenticationMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the token from the header
		isOK, _, token := validateAuthorizationHeader(ctx)
		if !isOK {
			return
		}

		// Query Redis to find profile with key is tokenMd5
		requester := cacheAuthentication(ctx, token)
		if requester == nil {
			requester = selfAuthentication(ctx, token)
			if requester == nil {
				requester = lifeAIAuthentication(ctx, token)
			}

			saveToCache(ctx, requester, token)
		}

		if requester == nil {
			httpresponse.Error(
				ctx,
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

		// Set the user in the context
		ctx.Set("requesterId", requester.ID)
		ctx.Set("requester", requester)
		ctx.Next()
	}
}
