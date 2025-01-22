package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

var (
	// encryptionKey will be of 32 bytes for AES-256 encryption
	encryptionKey = sha256.Sum256([]byte(os.Getenv("ENCRYPTION_KEY")))
)

// getSecretKeyFromEncryptionKey will convert that encryption key to byte slice
func getSecretKeyFromEncryptionKey() []byte {
	return encryptionKey[:]
}

// Encrypt encrypts the input string and returns base64 encoded string
func Encrypt(plaintext string) (string, error) {
	// Create new cipher block
	block, err := aes.NewCipher(getSecretKeyFromEncryptionKey())
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return as base64 encoded string
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the base64 encoded string
func Decrypt(encryptedString string) (string, error) {
	// Decode base64 string
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedString)
	if err != nil {
		return "", err
	}

	// Create new cipher block
	block, err := aes.NewCipher(getSecretKeyFromEncryptionKey())
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Get nonce size
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
