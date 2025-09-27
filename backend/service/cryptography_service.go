package service

import (
	"log"
)

// cryptographyService implements the CryptographyService interface
type cryptographyService struct{}

// NewCryptographyService creates a new cryptography service instance
func NewCryptographyService() CryptographyService {
	return &cryptographyService{}
}

// VigenereCipher performs XOR-based encryption/decryption using a repeating key
// This is a modern variant of the Vigenère cipher optimized for binary data
// Note: XOR is symmetric, so encryption and decryption are the same operation
func (c *cryptographyService) VigenereCipher(data []byte, key string, encrypt bool) []byte {
	if len(key) == 0 {
		log.Printf("[WARN] VigenereCipher: Empty key provided, returning data unchanged")
		return data
	}

	operation := "encrypting"
	if !encrypt {
		operation = "decrypting"
	}

	log.Printf("[DEBUG] VigenereCipher: %s %d bytes with key length %d", operation, len(data), len(key))

	result := make([]byte, len(data))
	keyBytes := []byte(key)

	// XOR each byte of data with repeating key
	for i, b := range data {
		keyByte := keyBytes[i%len(keyBytes)]
		result[i] = b ^ keyByte
	}

	log.Printf("[DEBUG] VigenereCipher: Successfully processed %d bytes", len(result))
	return result
}
