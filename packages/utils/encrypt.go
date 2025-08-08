package utils

import (
	"github.com/gtank/cryptopasta"
)

// Encrypt encrypts a plaintext string using the key
func Encrypt(key [32]byte, plaintext string) (string, error) {
	encrypted, err := cryptopasta.Encrypt([]byte(plaintext), &key)
	if err != nil {
		return "", err
	}
	return string(encrypted), nil
}

// Decrypt decrypts a ciphertext string using the key
func Decrypt(key [32]byte, ciphertext string) (string, error) {
	decrypted, err := cryptopasta.Decrypt([]byte(ciphertext), &key)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
