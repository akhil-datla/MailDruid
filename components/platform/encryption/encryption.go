package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"main/confidential"
	"os"

	"github.com/pterm/pterm"
)

func Decrypt(ciphertext []byte) []byte {

	// Key
	key := confidential.EncryptionKey

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		pterm.Error.Println("Error creating AES cipher: ", err)
		os.Exit(1)
	}

	// Before even testing the decryption,
	// if the text is too small, then it is incorrect
	if len(ciphertext) < aes.BlockSize {
		pterm.Error.Println("Ciphertext too short")
		os.Exit(1)
	}

	// Get the 16 byte IV
	iv := ciphertext[:aes.BlockSize]

	// Remove the IV from the ciphertext
	ciphertext = ciphertext[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from ciphertext
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext
}

func Encrypt(plainString string) []byte {

	plaintext := []byte(plainString)

	// Key
	key := confidential.EncryptionKey

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		pterm.Error.Println("Error creating AES cipher: ", err)
		os.Exit(1)
	}

	// Empty array of 16 + plaintext length
	// Include the IV at the beginning
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	// Slice of first 16 bytes
	iv := ciphertext[:aes.BlockSize]

	// Write 16 rand bytes to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		pterm.Error.Println("Error creating IV: ", err)
		os.Exit(1)
	}

	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt bytes from plaintext to ciphertext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext
}
