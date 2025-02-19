package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-iam/conf"
)

func CallLifeAIProfileAPI(authHeader string) (map[string]interface{}, error) {
	lifeAIConfig := conf.GetLifeAIConfiguration()

	// Assuming you have a package config that contains the AuthServiceURL
	url := lifeAIConfig.BackendURL + "/api/v1/user-profile/"

	// Create a new request using http package
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set the Authorization header
	req.Header.Set("Authorization", authHeader)

	// Send the request using http.DefaultClient
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the response status is not 200 OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status: %s", resp.Status)
	}

	// Parse the response body to get the requester_id
	type LifeAIProfile struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	var meResult struct {
		Profile LifeAIProfile `json:"profile"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&meResult); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":       meResult.Profile.ID,
		"username": meResult.Profile.Username,
		"email":    meResult.Profile.Email,
	}

	return result, nil
}

// RequestAuthenticationMiddleware returns a gin middleware for HTTP request logging
func RequestAuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Bearer token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header is required",
					"details": []interface{}{
						map[string]string{
							"field": "Authorization",
							"error": "Authorization header is required",
						},
					},
				},
			)
			return
		}

		// tokenMd5 := fmt.Sprintf("%x", md5.Sum([]byte(authHeader)))
		// Query Redis to find profile with key is tokenMd5

		// Check if the token is a Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header is required",
					"details": []interface{}{
						map[string]string{
							"field": "Authorization",
							"error": "Invalid authorization header format",
						},
					},
				},
			)
			return
		}

		// token := tokenParts[1]
		// You can now use the token for further processing

		// Check if request is using LifeAI token
		profile, err := CallLifeAIProfileAPI(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid token",
					"details": []interface{}{
						map[string]string{
							"field": "Authorization",
							"error": "Invalid token",
						},
					},
				},
			)
			return
		}

		// Check profile is exist or not -> if not exist, create new user

		// Set the profile in the context
		c.Set("profile", profile)

		// Process request
		c.Next()
	}
}
