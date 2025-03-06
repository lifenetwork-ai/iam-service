package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
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
	organizationId := ctx.Value("organizationId").(string)
	var requester *entities.IdentityUser = nil
	cacheKey := &cachingtypes.Keyer{
		Raw: fmt.Sprintf("%s_%x", organizationId, sha256.Sum256([]byte(token))),
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
	organizationId := ctx.Value("organizationId").(string)
	cacheKey := &cachingtypes.Keyer{
		Raw: fmt.Sprintf("%s_%x", organizationId, sha256.Sum256([]byte(token))),
	}

	if err := cacheRepo.SaveItem(cacheKey, *requester, 10*time.Minute); err != nil {
		logger.GetLogger().Errorf("Failed to cache user: %v", err)
	}
}

func getOrganizationProfile(
	ctx context.Context,
	organization *entities.IdentityOrganization,
	authHeader string,
) (map[string]interface{}, error) {
	if (organization == nil) || (organization.AuthenticateUrl == "") {
		return nil, fmt.Errorf("invalid organization or authenticate URL")
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, organization.AuthenticateUrl, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Authorization header
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Version-Management", "1.0.20|web")

	// Make the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, error: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response body
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return response, nil
}

func selfAuthentication(
	ctx *gin.Context,
	organization *entities.IdentityOrganization,
	authHeader string,
) *entities.IdentityUser {
	// Check if request is using self organization token
	// Dependency injection
	dbConnection := providers.ProvideDBConnection()
	cacheRepo := providers.ProvideCacheRepository(ctx)

	profile, err := getOrganizationProfile(ctx, organization, authHeader)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get profile from organization: %v", err)
		return nil
	}

	if profile == nil {
		logger.GetLogger().Errorf("Invalid profile from organization: %v", profile)
		return nil
	}

	profileId, ok := profile["id"]
	if !ok {
		logger.GetLogger().Errorf("Profile ID not found in organization response: %v", profile)
		return nil
	}

	email, ok := profile["email"]
	if !ok {
		logger.GetLogger().Debugf("Email not found in organization response: %v", profile)
		email = ""
	}

	phone, ok := profile["phone"]
	if !ok {
		logger.GetLogger().Debugf("Phone not found in organization response: %v", profile)
		phone = ""
	}

	selfID := strings.TrimSpace(fmt.Sprintf("%v", profileId))
	Email := strings.TrimSpace(fmt.Sprintf("%v", email))
	Phone := strings.TrimSpace(fmt.Sprintf("%v", phone))

	// Try to fix invalid email or phone
	EmailUpper := strings.ToUpper(Email)
	if EmailUpper == "NIL" || EmailUpper == "<NIL>" || EmailUpper == "NULL" || EmailUpper == "<NULL>" {
		Email = ""
	}

	// Try to fix invalid email or phone
	PhoneUpper := strings.ToUpper(Phone)
	if PhoneUpper == "NIL" || PhoneUpper == "<NIL>" || PhoneUpper == "NULL" || PhoneUpper == "<NULL>" {
		Phone = ""
	}

	if selfID == "" || (Email == "" && Phone == "") {
		logger.GetLogger().Errorf("Invalid profile from LifeAI: %v", profile)
		return nil
	}

	// Check profile is exist or not -> if not exist, create new user
	userRepo := repositories.NewIdentityUserRepository(dbConnection, cacheRepo)
	userByID, err := userRepo.FindBySelfAuthenticateID(ctx, selfID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get user with LifeAI profile by ID: %v", err)
		return nil
	}

	requester := userByID
	if requester == nil && Phone != "" {
		userByPhone, err := userRepo.FindByPhone(ctx, Phone)
		if err != nil {
			logger.GetLogger().Errorf("Failed to get user with LifeAI profile by phone: %v", err)
			return nil
		}
		requester = userByPhone
	}

	if requester == nil && Email != "" {
		userByEmail, err := userRepo.FindByEmail(ctx, Email)
		if err != nil {
			logger.GetLogger().Errorf("Failed to get user with LifeAI profile by email: %v", err)
			return nil
		}
		requester = userByEmail
	}

	if requester == nil {
		// Create new user
		username := func() string {
			if Phone != "" {
				return Phone
			}
			return Email
		}()

		newIdentityUser := &entities.IdentityUser{
			UserName:           username,
			Email:              Email,
			Phone:              Phone,
			SelfAuthenticateID: selfID,
		}

		if err := userRepo.Create(ctx, newIdentityUser); err != nil {
			logger.GetLogger().Errorf("Failed to create new user: %v", err)
			return nil
		}
		requester = newIdentityUser
	}

	return requester
}

func jwtAuthentication(
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
		logger.GetLogger().Errorf("Invalid session by access token")
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
			requester = jwtAuthentication(ctx, token)
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
		isOK, authHeader, token := validateAuthorizationHeader(ctx)
		if !isOK {
			return
		}

		// Query Redis to find profile with key is tokenMd5
		requester := cacheAuthentication(ctx, token)
		if requester == nil {
			requester = jwtAuthentication(ctx, token)
			if requester == nil {
				organization := ctx.Value("organization").(*entities.IdentityOrganization)
				if organization != nil && organization.SelfAuthenticate {
					requester = selfAuthentication(ctx, organization, authHeader)
				}
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
