package encryption

import (
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes
	svc, err := New(key)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	tests := []string{
		"hello world",
		"",
		"a",
		"this is a longer test string with special chars: !@#$%^&*()",
		"unicode: 日本語テスト",
	}

	for _, plaintext := range tests {
		encrypted, err := svc.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt(%q) error: %v", plaintext, err)
		}

		decrypted, err := svc.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("Decrypt() error: %v", err)
		}

		if decrypted != plaintext {
			t.Errorf("got %q, want %q", decrypted, plaintext)
		}
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	svc, err := New(key)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	a, _ := svc.Encrypt("same text")
	b, _ := svc.Encrypt("same text")

	if string(a) == string(b) {
		t.Error("two encryptions of same plaintext should produce different ciphertexts (random IV)")
	}
}

func TestNewRejectsInvalidKeySize(t *testing.T) {
	_, err := New([]byte("short"))
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestDecryptRejectsShortCiphertext(t *testing.T) {
	key := []byte("0123456789abcdef")
	svc, _ := New(key)

	_, err := svc.Decrypt([]byte("short"))
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}
