package service

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Nerggg/Audio-Steganography-LSB/backend/models"
)

// Test data for comprehensive testing
var (
	// Sample MP3-like data with proper frame headers
	testMP3Data = createTestMP3Data()

	// Test secret data
	testSecretData = []byte("This is a secret message for testing steganography methods!")
)

func createTestMP3Data() []byte {
	// Create a more realistic MP3 structure with proper frame headers
	data := make([]byte, 10000) // 10KB test file
	pos := 0

	// Create multiple valid frames
	for pos < len(data)-200 { // Leave space for frame
		frameSize := 144 // Simple frame size for testing
		if pos+frameSize > len(data) {
			break
		}

		// Frame header (4 bytes)
		data[pos] = 0xFF   // Sync byte 1
		data[pos+1] = 0xE3 // Sync byte 2 + version/layer (MPEG1 Layer3)
		data[pos+2] = 0x44 // Bitrate index 4 (56 kbps) + sample rate index 0 (44.1kHz)
		data[pos+3] = 0x00 // Padding + private + channel mode

		// Fill rest of frame with test data
		for i := pos + 4; i < pos+frameSize && i < len(data); i++ {
			data[i] = byte((i * 37) % 256) // Pseudo-random pattern
		}

		pos += frameSize
	}

	return data
}

// Integration tests for both steganography methods

// Test capacity calculation for both methods
func TestCapacityCalculationBothMethods(t *testing.T) {
	cryptoSvc := NewCryptographyService()
	audioSvc := NewAudioService()
	stegoSvc := NewStegoService(cryptoSvc, audioSvc)

	capacity, err := stegoSvc.CalculateCapacity(testMP3Data)
	if err != nil {
		t.Fatalf("CalculateCapacity failed: %v", err)
	}

	// Check that all capacity fields are set
	if capacity.OneLSB <= 0 {
		t.Error("OneLSB capacity should be positive")
	}
	if capacity.TwoLSB <= 0 {
		t.Error("TwoLSB capacity should be positive")
	}
	if capacity.ThreeLSB <= 0 {
		t.Error("ThreeLSB capacity should be positive")
	}
	if capacity.FourLSB <= 0 {
		t.Error("FourLSB capacity should be positive")
	}
	if capacity.Parity <= 0 {
		t.Error("Parity capacity should be positive")
	}

	// Check logical relationships
	if capacity.TwoLSB <= capacity.OneLSB {
		t.Error("TwoLSB should have higher capacity than OneLSB")
	}
	if capacity.ThreeLSB <= capacity.TwoLSB {
		t.Error("ThreeLSB should have higher capacity than TwoLSB")
	}
	if capacity.FourLSB <= capacity.ThreeLSB {
		t.Error("FourLSB should have higher capacity than ThreeLSB")
	}

	// Parity capacity should be equal to 1-LSB (both use 1 bit per byte)
	if capacity.Parity != capacity.OneLSB {
		t.Errorf("Parity capacity should equal OneLSB, got Parity: %d, OneLSB: %d", capacity.Parity, capacity.OneLSB)
	}

	t.Logf("Calculated capacities - 1-LSB: %d, 2-LSB: %d, 3-LSB: %d, 4-LSB: %d, Parity: %d",
		capacity.OneLSB, capacity.TwoLSB, capacity.ThreeLSB, capacity.FourLSB, capacity.Parity)
}

// Test LSB embedding and extraction (existing method)
func TestLSBEmbedExtract(t *testing.T) {
	cryptoSvc := NewCryptographyService()
	audioSvc := NewAudioService()
	stegoSvc := NewStegoService(cryptoSvc, audioSvc)

	// Test different LSB values
	lsbValues := []int{1, 2, 3, 4}

	for _, lsb := range lsbValues {
		t.Run(fmt.Sprintf("LSB_%d", lsb), func(t *testing.T) {
			embedReq := &models.EmbedRequest{
				CoverAudio:     testMP3Data,
				SecretFile:     testSecretData,
				SecretFileName: "test.txt",
				Method:         models.MethodLSB,
				NLsb:           lsb,
				StegoKey:       "",
				UseEncryption:  false,
				UseRandomStart: false,
			}

			// Embed the message
			stegoAudio, psnr, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
			if err != nil {
				t.Fatalf("EmbedMessage failed for %d-LSB: %v", lsb, err)
			}

			if psnr <= 0 {
				t.Errorf("PSNR should be positive, got: %f", psnr)
			}

			// Extract the message
			extractReq := &models.ExtractRequest{
				StegoAudio: stegoAudio,
				Method:     models.MethodLSB,
				StegoKey:   "",
			}

			extractedData, filename, err := stegoSvc.ExtractMessage(extractReq, stegoAudio)
			if err != nil {
				t.Fatalf("ExtractMessage failed for %d-LSB: %v", lsb, err)
			}

			// Verify extracted data
			if !bytes.Equal(testSecretData, extractedData) {
				t.Errorf("%d-LSB: Extracted data doesn't match original. Expected length %d, got %d",
					lsb, len(testSecretData), len(extractedData))
			}

			if filename != "test.txt" {
				t.Errorf("%d-LSB: Filename mismatch. Expected 'test.txt', got '%s'", lsb, filename)
			}

			t.Logf("%d-LSB embedding successful. PSNR: %.2f dB", lsb, psnr)
		})
	}
}

// Test Parity embedding and extraction (new method)
func TestParityEmbedExtract(t *testing.T) {
	cryptoSvc := NewCryptographyService()
	audioSvc := NewAudioService()
	stegoSvc := NewStegoService(cryptoSvc, audioSvc)

	embedReq := &models.EmbedRequest{
		CoverAudio:     testMP3Data,
		SecretFile:     testSecretData,
		SecretFileName: "test_parity.txt",
		Method:         models.MethodParity,
		NLsb:           1, // Not used for parity but required in struct
		StegoKey:       "",
		UseEncryption:  false,
		UseRandomStart: false,
	}

	// Embed the message using parity method
	stegoAudio, psnr, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
	if err != nil {
		t.Fatalf("Parity EmbedMessage failed: %v", err)
	}

	if psnr <= 0 {
		t.Errorf("PSNR should be positive, got: %f", psnr)
	}

	// Extract the message using parity method
	extractReq := &models.ExtractRequest{
		StegoAudio: stegoAudio,
		Method:     models.MethodParity,
		StegoKey:   "",
	}

	extractedData, filename, err := stegoSvc.ExtractMessage(extractReq, stegoAudio)
	if err != nil {
		t.Fatalf("Parity ExtractMessage failed: %v", err)
	}

	// Verify extracted data
	if !bytes.Equal(testSecretData, extractedData) {
		t.Errorf("Parity: Extracted data doesn't match original. Expected length %d, got %d",
			len(testSecretData), len(extractedData))
		t.Errorf("Original: %s", string(testSecretData))
		t.Errorf("Extracted: %s", string(extractedData))
	}

	if filename != "test_parity.txt" {
		t.Errorf("Parity: Filename mismatch. Expected 'test_parity.txt', got '%s'", filename)
	}

	t.Logf("Parity embedding successful. PSNR: %.2f dB", psnr)
}

// Test auto-detection during extraction
func TestAutoDetectionExtract(t *testing.T) {
	cryptoSvc := NewCryptographyService()
	audioSvc := NewAudioService()
	stegoSvc := NewStegoService(cryptoSvc, audioSvc)

	// Test auto-detection with LSB method
	t.Run("AutoDetect_LSB", func(t *testing.T) {
		embedReq := &models.EmbedRequest{
			CoverAudio:     testMP3Data,
			SecretFile:     testSecretData,
			SecretFileName: "auto_lsb.txt",
			Method:         models.MethodLSB,
			NLsb:           2,
			UseEncryption:  false,
			UseRandomStart: false,
		}

		stegoAudio, _, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
		if err != nil {
			t.Fatalf("Embed failed: %v", err)
		}

		// Extract without specifying method (auto-detect)
		extractReq := &models.ExtractRequest{
			StegoAudio: stegoAudio,
			// Method not specified - should auto-detect
		}

		extractedData, filename, err := stegoSvc.ExtractMessage(extractReq, stegoAudio)
		if err != nil {
			t.Fatalf("Auto-detect extract failed: %v", err)
		}

		if !bytes.Equal(testSecretData, extractedData) {
			t.Error("Auto-detected LSB extraction failed")
		}
		if filename != "auto_lsb.txt" {
			t.Errorf("Filename mismatch: expected 'auto_lsb.txt', got '%s'", filename)
		}
	})

	// Test auto-detection with Parity method
	t.Run("AutoDetect_Parity", func(t *testing.T) {
		embedReq := &models.EmbedRequest{
			CoverAudio:     testMP3Data,
			SecretFile:     testSecretData,
			SecretFileName: "auto_parity.txt",
			Method:         models.MethodParity,
			UseEncryption:  false,
			UseRandomStart: false,
		}

		stegoAudio, _, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
		if err != nil {
			t.Fatalf("Parity embed failed: %v", err)
		}

		// Extract without specifying method (auto-detect)
		extractReq := &models.ExtractRequest{
			StegoAudio: stegoAudio,
			// Method not specified - should auto-detect
		}

		extractedData, filename, err := stegoSvc.ExtractMessage(extractReq, stegoAudio)
		if err != nil {
			t.Fatalf("Auto-detect parity extract failed: %v", err)
		}

		if !bytes.Equal(testSecretData, extractedData) {
			t.Error("Auto-detected Parity extraction failed")
		}
		if filename != "auto_parity.txt" {
			t.Errorf("Filename mismatch: expected 'auto_parity.txt', got '%s'", filename)
		}
	})
}

// Test encryption with both methods
func TestEncryptionBothMethods(t *testing.T) {
	cryptoSvc := NewCryptographyService()
	audioSvc := NewAudioService()
	stegoSvc := NewStegoService(cryptoSvc, audioSvc)

	testKey := "secretkey123"

	methods := []struct {
		method models.SteganographyMethod
		nLsb   int
		name   string
	}{
		{models.MethodLSB, 2, "LSB_Encrypted"},
		{models.MethodParity, 1, "Parity_Encrypted"},
	}

	for _, methodTest := range methods {
		t.Run(methodTest.name, func(t *testing.T) {
			embedReq := &models.EmbedRequest{
				CoverAudio:     testMP3Data,
				SecretFile:     testSecretData,
				SecretFileName: "encrypted.txt",
				Method:         methodTest.method,
				NLsb:           methodTest.nLsb,
				StegoKey:       testKey,
				UseEncryption:  true,
				UseRandomStart: false,
			}

			stegoAudio, psnr, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
			if err != nil {
				t.Fatalf("%s embed with encryption failed: %v", methodTest.name, err)
			}

			// Extract with correct key
			extractReq := &models.ExtractRequest{
				StegoAudio: stegoAudio,
				Method:     methodTest.method,
				StegoKey:   testKey,
			}

			extractedData, _, err := stegoSvc.ExtractMessage(extractReq, stegoAudio)
			if err != nil {
				t.Fatalf("%s extract with encryption failed: %v", methodTest.name, err)
			}

			if !bytes.Equal(testSecretData, extractedData) {
				t.Errorf("%s: Encrypted data extraction failed", methodTest.name)
			}

			// Test extraction with wrong key (should fail)
			extractReqWrongKey := &models.ExtractRequest{
				StegoAudio: stegoAudio,
				Method:     methodTest.method,
				StegoKey:   "wrongkey",
			}

			_, _, err = stegoSvc.ExtractMessage(extractReqWrongKey, stegoAudio)
			if err == nil {
				t.Errorf("%s: Extraction with wrong key should fail", methodTest.name)
			}

			t.Logf("%s with encryption successful. PSNR: %.2f dB", methodTest.name, psnr)
		})
	}
}

// Test random start with both methods
func TestRandomStartBothMethods(t *testing.T) {
	cryptoSvc := NewCryptographyService()
	audioSvc := NewAudioService()
	stegoSvc := NewStegoService(cryptoSvc, audioSvc)

	testKey := "randomkey456"

	methods := []struct {
		method models.SteganographyMethod
		nLsb   int
		name   string
	}{
		{models.MethodLSB, 3, "LSB_RandomStart"},
		{models.MethodParity, 1, "Parity_RandomStart"},
	}

	for _, methodTest := range methods {
		t.Run(methodTest.name, func(t *testing.T) {
			embedReq := &models.EmbedRequest{
				CoverAudio:     testMP3Data,
				SecretFile:     testSecretData,
				SecretFileName: "random.txt",
				Method:         methodTest.method,
				NLsb:           methodTest.nLsb,
				StegoKey:       testKey,
				UseEncryption:  false,
				UseRandomStart: true,
			}

			stegoAudio, psnr, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
			if err != nil {
				t.Fatalf("%s embed with random start failed: %v", methodTest.name, err)
			}

			// Extract with correct key
			extractReq := &models.ExtractRequest{
				StegoAudio: stegoAudio,
				Method:     methodTest.method,
				StegoKey:   testKey,
			}

			extractedData, _, err := stegoSvc.ExtractMessage(extractReq, stegoAudio)
			if err != nil {
				t.Fatalf("%s extract with random start failed: %v", methodTest.name, err)
			}

			if !bytes.Equal(testSecretData, extractedData) {
				t.Errorf("%s: Random start data extraction failed", methodTest.name)
			}

			t.Logf("%s with random start successful. PSNR: %.2f dB", methodTest.name, psnr)
		})
	}
}

// Test error cases
func TestErrorCases(t *testing.T) {
	cryptoSvc := NewCryptographyService()
	audioSvc := NewAudioService()
	stegoSvc := NewStegoService(cryptoSvc, audioSvc)

	// Test invalid method
	t.Run("InvalidMethod", func(t *testing.T) {
		embedReq := &models.EmbedRequest{
			CoverAudio:     testMP3Data,
			SecretFile:     testSecretData,
			SecretFileName: "test.txt",
			Method:         "invalid", // Invalid method
			NLsb:           1,
		}

		_, _, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
		if err == nil {
			t.Error("EmbedMessage should fail with invalid method")
		}
	})

	// Test invalid LSB for LSB method
	t.Run("InvalidLSB", func(t *testing.T) {
		embedReq := &models.EmbedRequest{
			CoverAudio:     testMP3Data,
			SecretFile:     testSecretData,
			SecretFileName: "test.txt",
			Method:         models.MethodLSB,
			NLsb:           5, // Invalid LSB value
		}

		_, _, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
		if err == nil {
			t.Error("EmbedMessage should fail with invalid LSB value")
		}
	})

	// Test insufficient capacity
	t.Run("InsufficientCapacity", func(t *testing.T) {
		// Create very large secret data
		largeSecret := make([]byte, 50000) // 50KB
		for i := range largeSecret {
			largeSecret[i] = byte(i % 256)
		}

		embedReq := &models.EmbedRequest{
			CoverAudio:     testMP3Data, // Only 10KB test data
			SecretFile:     largeSecret,
			SecretFileName: "large.txt",
			Method:         models.MethodParity, // Lowest capacity method
		}

		_, _, err := stegoSvc.EmbedMessage(embedReq, largeSecret, nil)
		if err == nil {
			t.Error("EmbedMessage should fail with insufficient capacity")
		}
	})

	// Test missing key when required
	t.Run("MissingKeyForEncryption", func(t *testing.T) {
		embedReq := &models.EmbedRequest{
			CoverAudio:     testMP3Data,
			SecretFile:     testSecretData,
			SecretFileName: "test.txt",
			Method:         models.MethodLSB,
			NLsb:           1,
			UseEncryption:  true,
			StegoKey:       "", // Missing key
		}

		_, _, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
		if err == nil {
			t.Error("EmbedMessage should fail when encryption is enabled but no key provided")
		}
	})

	// Test invalid audio data
	t.Run("InvalidAudioData", func(t *testing.T) {
		embedReq := &models.EmbedRequest{
			CoverAudio:     []byte("not audio"), // Invalid audio
			SecretFile:     testSecretData,
			SecretFileName: "test.txt",
			Method:         models.MethodLSB,
			NLsb:           1,
		}

		_, _, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
		if err == nil {
			t.Error("EmbedMessage should fail with invalid audio data")
		}
	})
}

// Test method validation functions
func TestMethodValidation(t *testing.T) {
	// Test valid methods
	validMethods := []models.SteganographyMethod{
		models.MethodLSB,
		models.MethodParity,
	}

	for _, method := range validMethods {
		if !method.IsValid() {
			t.Errorf("Method %s should be valid", method)
		}
	}

	// Test invalid methods
	invalidMethods := []models.SteganographyMethod{
		"invalid",
		"",
		"LSB",    // Wrong case
		"PARITY", // Wrong case
		"unknown",
	}

	for _, method := range invalidMethods {
		if method.IsValid() {
			t.Errorf("Method %s should be invalid", method)
		}
	}
}

// Benchmark tests for performance comparison
func BenchmarkLSBEmbed(b *testing.B) {
	cryptoSvc := NewCryptographyService()
	audioSvc := NewAudioService()
	stegoSvc := NewStegoService(cryptoSvc, audioSvc)

	embedReq := &models.EmbedRequest{
		CoverAudio:     testMP3Data,
		SecretFile:     testSecretData,
		SecretFileName: "bench.txt",
		Method:         models.MethodLSB,
		NLsb:           2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkParityEmbed(b *testing.B) {
	cryptoSvc := NewCryptographyService()
	audioSvc := NewAudioService()
	stegoSvc := NewStegoService(cryptoSvc, audioSvc)

	embedReq := &models.EmbedRequest{
		CoverAudio:     testMP3Data,
		SecretFile:     testSecretData,
		SecretFileName: "bench.txt",
		Method:         models.MethodParity,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := stegoSvc.EmbedMessage(embedReq, testSecretData, nil)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
