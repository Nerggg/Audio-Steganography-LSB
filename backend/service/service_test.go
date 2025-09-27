package service

import (
	"bytes"
	"testing"
)

func TestVigenereCipher(t *testing.T) {
	cryptoSvc := NewCryptographyService()

	testData := []byte("Hello, World!")
	key := "secret"

	// Encrypt the data
	encrypted := cryptoSvc.VigenereCipher(testData, key, true)

	// Decrypt the data
	decrypted := cryptoSvc.VigenereCipher(encrypted, key, false)

	// Verify the result matches original
	if !bytes.Equal(testData, decrypted) {
		t.Errorf("VigenereCipher failed: expected %s, got %s", string(testData), string(decrypted))
	}

	// Verify that encrypted data is different from original
	if bytes.Equal(testData, encrypted) {
		t.Error("VigenereCipher failed: encrypted data is same as original")
	}
}

func TestBytesToBits(t *testing.T) {
	testData := []byte{0xFF, 0x00, 0xAA} // 11111111 00000000 10101010
	expectedBits := []int{1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 1, 0, 1, 0}

	bits := bytesToBits(testData)

	if len(bits) != len(expectedBits) {
		t.Errorf("bytesToBits failed: expected length %d, got %d", len(expectedBits), len(bits))
		return
	}

	for i, bit := range bits {
		if bit != expectedBits[i] {
			t.Errorf("bytesToBits failed at index %d: expected %d, got %d", i, expectedBits[i], bit)
		}
	}
}

func TestBitsToBytes(t *testing.T) {
	testBits := []int{1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 1, 0, 1, 0}
	expectedData := []byte{0xFF, 0x00, 0xAA}

	data := bitsToBytes(testBits)

	if !bytes.Equal(data, expectedData) {
		t.Errorf("bitsToBytes failed: expected %v, got %v", expectedData, data)
	}
}

func TestCapacityCalculation(t *testing.T) {
	// Create mock service
	stegoSvc := NewSteganographyService()

	// Test with invalid data
	_, err := stegoSvc.CalculateCapacity([]byte("invalid audio data"))
	if err == nil {
		t.Error("CalculateCapacity should fail with invalid audio data")
	}
}

func TestMetadataCreation(t *testing.T) {
	stegoSvc := NewSteganographyService()

	filename := "test.txt"
	fileSize := 1024
	useEncryption := true
	useRandomStart := false
	nLsb := 2

	metadata := stegoSvc.CreateMetadata(filename, fileSize, useEncryption, useRandomStart, nLsb)

	// Check metadata length
	if len(metadata) < 38 {
		t.Error("Metadata should be at least 38 bytes")
	}

	// Check flags
	flags := metadata[36]
	if (flags & 0x01) == 0 { // Should have encryption flag set
		t.Error("Encryption flag should be set")
	}

	if (flags & 0x02) != 0 { // Should not have random start flag set
		t.Error("Random start flag should not be set")
	}
}

func TestWAVEncoding(t *testing.T) {
	encoder := NewAudioEncoder()

	// Create test PCM data (16-bit stereo samples)
	pcmData := make([]byte, 1024) // 512 samples for stereo
	for i := 0; i < len(pcmData); i += 2 {
		// Create a simple sine wave pattern
		pcmData[i] = byte(i % 256)
		pcmData[i+1] = byte((i + 1) % 256)
	}

	sampleRate := 44100

	wavData, err := encoder.EncodeToWAV(pcmData, sampleRate)
	if err != nil {
		t.Errorf("WAV encoding failed: %v", err)
		return
	}

	// Check WAV header structure
	if len(wavData) < 44 {
		t.Error("WAV data too short to contain header")
		return
	}

	// Check RIFF signature
	if string(wavData[:4]) != "RIFF" {
		t.Error("WAV should start with RIFF signature")
	}

	// Check WAVE format
	if string(wavData[8:12]) != "WAVE" {
		t.Error("WAV should contain WAVE format identifier")
	}

	// Check fmt chunk
	if string(wavData[12:16]) != "fmt " {
		t.Error("WAV should contain fmt chunk")
	}

	// Check data chunk
	dataChunkPos := 36
	if string(wavData[dataChunkPos:dataChunkPos+4]) != "data" {
		t.Error("WAV should contain data chunk")
	}

	// Verify the data size matches
	expectedSize := len(pcmData) + 44 // PCM data + header
	if len(wavData) != expectedSize {
		t.Errorf("WAV size mismatch: expected %d, got %d", expectedSize, len(wavData))
	}
}