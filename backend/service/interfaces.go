package service

import (
	"github.com/Nerggg/Audio-Steganography-LSB/backend/models"
)

// SteganographyService defines the interface for steganography operations
type SteganographyService interface {
	// CalculateCapacity calculates the embedding capacity for different LSB methods
	CalculateCapacity(audioData []byte) (*models.CapacityResult, error)

	// EmbedMessage embeds a secret message into audio data
	EmbedMessage(req *models.EmbedRequest, secretData []byte, metadata []byte) ([]byte, float64, error)

	// ExtractMessage extracts a secret message from audio data
	ExtractMessage(req *models.ExtractRequest, audioData []byte) ([]byte, string, error)

	// CreateMetadata creates metadata for embedding
	CreateMetadata(filename string, fileSize int, useEncryption, useRandomStart bool, nLsb int) []byte
}

// CryptographyService defines the interface for cryptographic operations
type CryptographyService interface {
	// VigenereCipher performs Vigenère cipher encryption/decryption
	VigenereCipher(data []byte, key string, encrypt bool) []byte
}

// AudioService defines the interface for audio processing operations
type AudioService interface {
	// CalculatePSNR calculates Peak Signal-to-Noise Ratio between original and modified audio
	CalculatePSNR(original, modified []byte) float64
}

// AudioEncoder defines the interface for audio encoding operations
type AudioEncoder interface {
	// EncodeToWAV encodes PCM data to WAV format
	EncodeToWAV(pcmData []byte, sampleRate int) ([]byte, error)

	// EncodeToMP3 encodes PCM data to MP3 format using quantization noise steganography
	EncodeToMP3(pcmData []byte, sampleRate int) ([]byte, error)

	// extractFromSteganographicMP3 extracts PCM data from steganographic MP3 format
	extractFromSteganographicMP3(mp3Data []byte) ([]byte, error)
}
