package utils

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := [32]byte{}
	copy(key[:], []byte("12345678901234567890123456789012"))
	plaintext := "Hello, World!"
	encrypted, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("Decrypted text does not match original plaintext")
	}
}
