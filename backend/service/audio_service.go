package service

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
)

// audioService implements the AudioService interface
type audioService struct{}

// audioEncoder implements the AudioEncoder interface
type audioEncoder struct{}

// NewAudioService creates a new audio service instance
func NewAudioService() AudioService {
	return &audioService{}
}

// NewAudioEncoder creates a new audio encoder instance
func NewAudioEncoder() AudioEncoder {
	return &audioEncoder{}
}

// CalculatePSNR calculates Peak Signal-to-Noise Ratio between original and modified audio
// According to specification: PSNR = 10 * log10(MAX²/MSE)
// Minimum PSNR threshold: 30 dB (values below indicate significant audio degradation)
func (a *audioService) CalculatePSNR(original, modified []byte) float64 {
	if len(original) != len(modified) {
		log.Printf("[WARN] CalculatePSNR: Length mismatch - original: %d, modified: %d", len(original), len(modified))
		return 0.0
	}

	var mse float64
	sampleCount := len(original) / 2 // 16-bit samples

	for i := 0; i < len(original)-1; i += 2 {
		// Convert bytes to 16-bit signed integers (little-endian)
		originalSample := int16(binary.LittleEndian.Uint16(original[i : i+2]))
		modifiedSample := int16(binary.LittleEndian.Uint16(modified[i : i+2]))

		diff := float64(originalSample - modifiedSample)
		mse += diff * diff
	}

	if sampleCount == 0 {
		return 0.0
	}

	mse /= float64(sampleCount)

	// Avoid division by zero
	if mse == 0 {
		return math.Inf(1) // Perfect match
	}

	// Calculate PSNR using specification formula: PSNR = 10 * log10(MAX²/MSE)
	// MAX = 32767 for 16-bit PCM (as per specification)
	maxValue := 32767.0 // Maximum value for 16-bit signed integer
	psnr := 10 * math.Log10((maxValue*maxValue)/mse)

	log.Printf("[DEBUG] CalculatePSNR: MSE=%.6f, PSNR=%.2f dB (samples: %d)", mse, psnr, sampleCount)
	return psnr
}

// EncodeToWAV encodes PCM data to WAV format
func (e *audioEncoder) EncodeToWAV(pcmData []byte, sampleRate int) ([]byte, error) {
	var wav bytes.Buffer

	// WAV header structure
	dataSize := len(pcmData)
	fileSize := 36 + dataSize

	// RIFF header
	wav.Write([]byte("RIFF"))
	binary.Write(&wav, binary.LittleEndian, uint32(fileSize))
	wav.Write([]byte("WAVE"))

	// fmt chunk
	wav.Write([]byte("fmt "))
	binary.Write(&wav, binary.LittleEndian, uint32(16)) // fmt chunk size
	binary.Write(&wav, binary.LittleEndian, uint16(1))  // PCM format
	binary.Write(&wav, binary.LittleEndian, uint16(2))  // stereo channels
	binary.Write(&wav, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&wav, binary.LittleEndian, uint32(sampleRate*2*2)) // byte rate
	binary.Write(&wav, binary.LittleEndian, uint16(4))              // block align
	binary.Write(&wav, binary.LittleEndian, uint16(16))             // bits per sample

	// data chunk
	wav.Write([]byte("data"))
	binary.Write(&wav, binary.LittleEndian, uint32(dataSize))
	wav.Write(pcmData)

	return wav.Bytes(), nil
}
