package utils

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
	"unicode/utf8"
)

func mustKey() [32]byte {
	var k [32]byte
	copy(k[:], []byte("12345678901234567890123456789012"))
	return k
}

func TestEncryptDecrypt_Table(t *testing.T) {
	key := mustKey()

	tests := []struct {
		name      string
		plaintext string
		wantErr   bool // set true if you decided Encrypt("") should error
	}{
		{"ascii", "Hello, World!", false},
		{"unicode", "Xin chào 👋🏼 Việt Nam — 🧪", false},
		{"very_long", strings.Repeat("a", 1<<20), false}, // 1MB
		{"empty", "", true},                              // set to false if you allow empty plaintext
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := Encrypt(key, tt.plaintext)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Encrypt err=%v wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// Must be valid base64 and non-empty
			raw, err := base64.StdEncoding.DecodeString(ct)
			if err != nil || len(raw) == 0 {
				t.Fatalf("ciphertext not valid base64 or empty, err=%v", err)
			}

			pt, err := Decrypt(key, ct)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}
			if pt != tt.plaintext {
				t.Fatalf("roundtrip mismatch")
			}
			// If you expect textual data, ensure UTF-8 stayed intact.
			if !utf8.ValidString(pt) {
				t.Fatalf("plaintext not valid UTF-8 after roundtrip")
			}
		})
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	keyA := mustKey()
	var keyB [32]byte
	copy(keyB[:], []byte("abcdef9876543210abcdef9876543210"))

	ct, err := Encrypt(keyA, "secret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if _, err := Decrypt(keyB, ct); err == nil {
		t.Fatalf("Decrypt with wrong key should fail")
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	key := mustKey()
	_, err := Decrypt(key, "not-base64!!!")
	if err == nil {
		t.Fatalf("expected error for invalid base64 input")
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	key := mustKey()
	ct, err := Encrypt(key, "integrity")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Flip a byte in the base64-decoded ciphertext
	raw, err := base64.StdEncoding.DecodeString(ct)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	raw[len(raw)/2] ^= 0xFF
	tampered := base64.StdEncoding.EncodeToString(raw)

	if _, err := Decrypt(key, tampered); err == nil {
		t.Fatalf("tampering should cause decryption/auth failure")
	}
}

func TestCiphertext_Nondeterministic(t *testing.T) {
	// AEAD with random nonce should produce different ciphertexts for same plaintext
	key := mustKey()
	pt := "same message"

	ct1, err := Encrypt(key, pt)
	if err != nil {
		t.Fatalf("Encrypt1: %v", err)
	}
	ct2, err := Encrypt(key, pt)
	if err != nil {
		t.Fatalf("Encrypt2: %v", err)
	}
	if ct1 == ct2 {
		t.Fatalf("ciphertexts should differ (nonce randomness)")
	}
}

func TestAPI_Contract_BinarySafe(t *testing.T) {
	key := mustKey()
	bin := bytes.Repeat([]byte{0x00, 0xFF, 0x01, 0x02}, 1024)

	ct, err := Encrypt(key, string(bin)) // string can hold arbitrary bytes
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	out, err := Decrypt(key, ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal([]byte(out), bin) {
		t.Fatalf("binary roundtrip mismatch")
	}
}

func FuzzEncryptDecrypt(f *testing.F) {
	key := mustKey()

	// Seed corpus
	seeds := []string{"", "a", "Hello", "Xin chào 👋🏼", strings.Repeat("x", 10_000)}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, msg string) {
		// If you forbid empty plaintext in Encrypt, skip it.
		if msg == "" {
			t.Skip()
		}

		ct, err := Encrypt(key, msg)
		if err != nil {
			t.Fatalf("Encrypt: %v", err)
		}

		pt, err := Decrypt(key, ct)
		if err != nil {
			t.Fatalf("Decrypt: %v", err)
		}

		if pt != msg {
			t.Fatalf("roundtrip mismatch")
		}
	})
}
