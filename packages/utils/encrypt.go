package utils

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/gtank/cryptopasta"
)

// Encrypt encrypts a plaintext string using the key and returns a base64 string.
func Encrypt(key [32]byte, plaintext string) (string, error) {
	if plaintext == "" {
		return "", errors.New("plaintext is empty")
	}

	encrypted, err := cryptopasta.Encrypt([]byte(plaintext), &key)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt plaintext: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt decrypts a base64-encoded ciphertext string using the key.
func Decrypt(key [32]byte, ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", errors.New("ciphertext is empty")
	}

	encrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}
	if len(encrypted) == 0 {
		return "", errors.New("decoded ciphertext is empty")
	}

	decrypted, err := cryptopasta.Decrypt(encrypted, &key)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}

	return string(decrypted), nil
}
