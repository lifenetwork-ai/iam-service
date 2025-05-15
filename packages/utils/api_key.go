package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateAPIKey generates a random API key
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32) // Generate 256-bit (32 bytes) random key
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error generating random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
