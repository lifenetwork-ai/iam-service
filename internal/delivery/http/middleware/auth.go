package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/conf"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"golang.org/x/crypto/bcrypt"
)

// RootAuthMiddleware returns a gin middleware for root authentication
func RootAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for Swagger UI
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.Header("WWW-Authenticate", `Basic realm="Root Area"`)
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
			c.Abort()
			return
		}

		if !strings.HasPrefix(auth, "Basic ") {
			c.Header("WWW-Authenticate", `Basic realm="Root Area"`)
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid authorization header format",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid authorization header format",
				}},
			)
			c.Abort()
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			c.Header("WWW-Authenticate", `Basic realm="Root Area"`)
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid authorization header format",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid authorization header format",
				}},
			)
			c.Abort()
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 {
			c.Header("WWW-Authenticate", `Basic realm="Root Area"`)
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid authorization header format",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid authorization header format",
				}},
			)
			c.Abort()
			return
		}

		// Get root credentials from config
		config := conf.GetConfiguration()
		rootEmail := config.AdminAccount.AdminEmail
		rootPassword := config.AdminAccount.AdminPassword

		// Validate root credentials
		if pair[0] != rootEmail || pair[1] != rootPassword {
			c.Header("WWW-Authenticate", `Basic realm="Root Area"`)
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid credentials",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid credentials",
				}},
			)
			c.Abort()
			return
		}

		// Set root info in context
		c.Set("isRoot", true)
		c.Set("rootEmail", rootEmail)
		c.Set("role", "ROOT")

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

		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
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
			c.Abort()
			return
		}

		if !strings.HasPrefix(auth, "Basic ") {
			c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid authorization header format",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid authorization header format",
				}},
			)
			c.Abort()
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid authorization header format",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid authorization header format",
				}},
			)
			c.Abort()
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 {
			c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid authorization header format",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid authorization header format",
				}},
			)
			c.Abort()
			return
		}

		email := pair[0]
		password := pair[1]

		// Get admin account from database
		account, err := adminRepo.GetByEmail(email)
		if err != nil || account == nil {
			c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid credentials",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid credentials",
				}},
			)
			c.Abort()
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
			c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid credentials",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Invalid credentials",
				}},
			)
			c.Abort()
			return
		}

		// Check if account has admin role
		if account.Role != "ADMIN" && account.Role != "ROOT" {
			c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
			httpresponse.Error(
				c,
				http.StatusForbidden,
				"FORBIDDEN",
				"Insufficient privileges",
				[]map[string]string{{
					"field": "Authorization",
					"error": "Insufficient privileges",
				}},
			)
			c.Abort()
			return
		}

		// Set admin info in context
		c.Set("isAdmin", true)
		c.Set("adminEmail", account.Email)
		c.Set("adminID", account.ID.String())
		c.Set("role", account.Role)

		c.Next()
	}
}
