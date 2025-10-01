package models

import (
	"errors"
)

// Predefined errors for steganography operations
var (
	ErrInvalidMP3           = errors.New("failed to decode audio data, not a valid MP3 file")
	ErrInsufficientCapacity = errors.New("insufficient audio capacity for the provided data")
	ErrInvalidLSB           = errors.New("LSB value must be between 1 and 4")
	ErrInvalidMethod        = errors.New("invalid steganography method, must be 'lsb' or 'parity'")
	ErrInvalidStegoKey      = errors.New("steganography key cannot be empty when encryption or random start is enabled")
	ErrInvalidSignature     = errors.New("invalid steganography signature - data may not be embedded or corrupted")
	ErrFileTooLarge         = errors.New("file size exceeds maximum allowed limit")
	ErrInvalidFileFormat    = errors.New("invalid file format")
	ErrCorruptedData        = errors.New("embedded data appears to be corrupted")
	ErrExtractionFailed     = errors.New("failed to extract data - wrong key or parameters")
)

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}
