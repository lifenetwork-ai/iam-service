package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/conf"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

const (
	RoleAdmin = "admin"
	RoleRoot  = "root"
)

var (
	// Context keys
	ContextKeyIsRoot        = "isRoot"
	ContextKeyRootUsername  = "rootUsername"
	ContextKeyRole          = "role"
	ContextKeyIsAdmin       = "isAdmin"
	ContextKeyAdminUsername = "adminUsername"
	ContextKeyAdminID       = "adminID"
)

// validateBasicAuth validates and extracts credentials from Basic auth header
func validateBasicAuth(authHeader string) (username, password string, err error) {
	if authHeader == "" {
		return "", "", fmt.Errorf("authorization header is required")
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		return "", "", fmt.Errorf("invalid authorization header format")
	}

	payload, err := base64.StdEncoding.DecodeString(authHeader[6:])
	if err != nil {
		return "", "", fmt.Errorf("invalid authorization header format")
	}

	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return "", "", fmt.Errorf("invalid authorization header format")
	}

	return pair[0], pair[1], nil
}

// isRootUser checks if the provided credentials match the root user
func isRootUser(username, password string) bool {
	config := conf.GetConfiguration()
	rootUsername := config.RootAccount.RootUsername
	rootPassword := config.RootAccount.RootPassword

	return username == rootUsername && password == rootPassword
}

// setRootContext sets root user context variables
func setRootContext(c *gin.Context, username string) {
	c.Set(ContextKeyIsRoot, true)
	c.Set(ContextKeyRootUsername, username)
	c.Set(ContextKeyRole, RoleRoot)
}

// setAdminContext sets admin user context variables
func setAdminContext(c *gin.Context, account *domain.AdminAccount) {
	c.Set(ContextKeyIsAdmin, true)
	c.Set(ContextKeyAdminUsername, account.Username)
	c.Set(ContextKeyAdminID, account.ID.String())
	c.Set(ContextKeyRole, account.Role)
}

// sendAuthError sends authentication error response
func sendAuthError(c *gin.Context, realm, message string, statusCode int) {
	c.Header("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))

	errorCode := "UNAUTHORIZED"
	if statusCode == http.StatusForbidden {
		errorCode = "FORBIDDEN"
	}

	httpresponse.Error(
		c,
		statusCode,
		errorCode,
		message,
		[]map[string]string{{
			"field": "Authorization",
			"error": message,
		}},
	)
	c.Abort()
}

// RootAuthMiddleware returns a gin middleware for root authentication
func RootAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for Swagger UI
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Next()
			return
		}

		username, password, err := validateBasicAuth(c.GetHeader("Authorization"))
		if err != nil {
			sendAuthError(c, "Root Area", err.Error(), http.StatusUnauthorized)
			return
		}

		if !isRootUser(username, password) {
			sendAuthError(c, "Root Area", "Invalid credentials", http.StatusUnauthorized)
			return
		}

		setRootContext(c, username)
		c.Next()
	}
}

// AdminAuthMiddleware returns a gin middleware for admin authentication
func AdminAuthMiddleware(adminRepo interfaces.AdminAccountRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for Swagger UI
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Next()
			return
		}

		username, password, err := validateBasicAuth(c.GetHeader("Authorization"))
		if err != nil {
			sendAuthError(c, "Admin Area", err.Error(), http.StatusUnauthorized)
			return
		}
		logger.GetLogger().Infof("AdminAuthMiddleware: username: %s, password: %s", username, password)

		// Check if this is a root user first
		if isRootUser(username, password) {
			setRootContext(c, username)
			c.Next()
			return
		}

		// Get admin account from database
		account, err := adminRepo.GetByUsername(username)
		if err != nil || account == nil {
			sendAuthError(c, "Admin Area", "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
			sendAuthError(c, "Admin Area", "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Check if account has admin role
		if account.Role != RoleAdmin && account.Role != RoleRoot {
			sendAuthError(c, "Admin Area", "Insufficient privileges", http.StatusForbidden)
			return
		}

		setAdminContext(c, account)
		c.Next()
	}
}
