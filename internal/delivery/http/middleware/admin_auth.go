package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/conf"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
)

// AdminBasicAuthMiddleware returns a gin middleware for admin basic authentication
func AdminBasicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for Swagger UI
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Next()
			return
		}

		// Get the Basic Auth header
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

		// Check if it's a Basic Auth header
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

		// Decode the Base64 string
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

		// Split username and password
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

		// Get admin credentials from config
		config := conf.GetConfiguration()
		adminEmail := config.AdminAccount.AdminEmail
		adminPassword := config.AdminAccount.AdminPassword

		// Validate credentials
		if pair[0] != adminEmail || pair[1] != adminPassword {
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

		// Set admin info in context
		c.Set("isAdmin", true)
		c.Set("adminEmail", adminEmail)

		c.Next()
	}
}
