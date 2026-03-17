package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Service provides AES-256-CFB encryption and decryption.
type Service struct {
	key []byte
}

// New creates an encryption service with the given key.
// Key must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256.
func New(key []byte) (*Service, error) {
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, fmt.Errorf("invalid key size %d: must be 16, 24, or 32 bytes", keyLen)
	}
	keyCopy := make([]byte, keyLen)
	copy(keyCopy, key)
	return &Service{key: keyCopy}, nil
}

// Encrypt encrypts plaintext using AES-CFB with a random IV.
func (s *Service) Encrypt(plaintext string) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	plain := []byte(plaintext)
	ciphertext := make([]byte, aes.BlockSize+len(plain))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generating IV: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plain)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext produced by Encrypt.
func (s *Service) Decrypt(ciphertext []byte) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short: %d bytes", len(ciphertext))
	}

	iv := ciphertext[:aes.BlockSize]
	encrypted := make([]byte, len(ciphertext)-aes.BlockSize)
	copy(encrypted, ciphertext[aes.BlockSize:])

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encrypted, encrypted)

	return string(encrypted), nil
}
