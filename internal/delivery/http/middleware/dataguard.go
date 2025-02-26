package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	httpresponse "github.com/genefriendway/human-network-iam/packages/http/response"
)

// IsSQLInjection checks if the input contains common SQL injection patterns
func IsSQLInjection(input string) bool {
	lowerInput := strings.ToLower(input)

	// List of suspicious SQL keywords and characters
	sqlPatterns := []string{
		"select ", "insert ", "update ", "delete ", "drop ", "alter ", "truncate ",
		"union ", "exec ", "or ", "and ", "like ", "benchmark(", "sleep(", "load_file(",
		"outfile ", "--", ";", "#", "/*", "xp_", "declare ", "cast(", "convert(",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	// Regular expression to detect suspicious SQL syntax patterns
	sqlRegex := regexp.MustCompile(`(?i)(\b(select|insert|update|delete|drop|alter|truncate|union|exec|or|and)\b|(--|;|#|/\*|\*/|xp_|declare|benchmark\(|sleep\(|load_file\(|outfile))`)
	return sqlRegex.MatchString(input)
}

// ValidateJSONPayload checks all string fields in JSON request for SQL injection attempts
func ValidateJSONPayload(data map[string]interface{}) bool {
	for _, value := range data {
		switch v := value.(type) {
		case string:
			if IsSQLInjection(v) {
				return true
			}
		case map[string]interface{}:
			if ValidateJSONPayload(v) { // Recursively check nested JSON objects
				return true
			}
		case []interface{}:
			for _, item := range v {
				if str, ok := item.(string); ok && IsSQLInjection(str) {
					return true
				}
			}
		}
	}
	return false
}

// RequestDataGuardMiddleware returns a gin middleware for HTTP request authorization
func RequestDataGuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
			c.Next()
			return
		}

		var bodyBytes []byte
		var err error

		// Read request body
		if c.Request.Body != nil {
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				httpresponse.Error(
					c,
					http.StatusUnauthorized,
					"FAILED_TO_READ_REQUEST_BODY",
					"Failed to read request body",
					[]map[string]string{{
						"field": "DataGuard",
						"error": "Failed to read request body",
					}},
				)
				return
			}
		}

		// Restore the request body for next handler
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse JSON
		var jsonData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"INVALID_JSON_FORMAT",
				"Invalid JSON format",
				[]map[string]string{{
					"field": "DataGuard",
					"error": "Invalid JSON format",
				}},
			)
			return
		}

		// Validate for SQL injection
		if ValidateJSONPayload(jsonData) {
			httpresponse.Error(
				c,
				http.StatusUnauthorized,
				"SQL_INJECTION_ATTEMPT",
				"SQL Injection attempt detected!",
				[]map[string]string{{
					"field": "DataGuard",
					"error": "SQL Injection attempt detected!",
				}},
			)
			return
		}

		c.Next()
	}
}
