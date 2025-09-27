package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Nerggg/Audio-Steganography-LSB/backend/models"
	"github.com/hajimehoshi/go-mp3"
)

// steganographyService implements the SteganographyService interface
type steganographyService struct{}

// NewSteganographyService creates a new steganography service instance
func NewSteganographyService() SteganographyService {
	return &steganographyService{}
}

// CalculateCapacity calculates the embedding capacity for different LSB methods
func (s *steganographyService) CalculateCapacity(audioData []byte) (*models.CapacityResult, error) {
	start := time.Now()
	log.Printf("[DEBUG] CalculateCapacity: Starting capacity calculation for audio data (size: %d bytes)", len(audioData))

	var pcmData []byte
	var err error

	// Try MP3 decoder first
	audioReader := bytes.NewReader(audioData)
	decoder, err := mp3.NewDecoder(audioReader)
	if err == nil {
		// Successfully created MP3 decoder
		pcmData, err = io.ReadAll(decoder)
		if err != nil {
			log.Printf("[ERROR] CalculateCapacity: Failed to read PCM data from MP3: %v", err)
			return nil, errors.New("could not read decoded audio stream: " + err.Error())
		}
		log.Printf("[DEBUG] CalculateCapacity: Successfully decoded MP3 to PCM (size: %d bytes)", len(pcmData))
	} else {
		// MP3 decoder failed, try WAV
		log.Printf("[DEBUG] CalculateCapacity: MP3 decoder failed, trying WAV: %v", err)

		// Parse WAV header to extract PCM data
		dataOffset, dataSize, err := parseWAVHeader(audioData)
		if err != nil {
			log.Printf("[ERROR] CalculateCapacity: Failed to parse WAV header: %v", err)
			return nil, errors.New("unsupported audio format: not a valid MP3 or WAV file")
		}

		// For capacity calculation, we need the PCM data
		if len(audioData) <= dataOffset {
			log.Printf("[ERROR] CalculateCapacity: WAV file too small to contain PCM data")
			return nil, errors.New("WAV file is too small or corrupted")
		}

		pcmData = audioData[dataOffset : dataOffset+int(dataSize)]
		log.Printf("[DEBUG] CalculateCapacity: Successfully extracted PCM from WAV (size: %d bytes, data_offset: %d)",
			len(pcmData), dataOffset)
	}

	// Calculate total samples (16-bit stereo = 2 bytes per sample)
	totalSamples := len(pcmData) / 2

	// Handle odd PCM data length
	if len(pcmData)%2 != 0 {
		totalSamples = (len(pcmData) - 1) / 2
		log.Printf("[WARN] CalculateCapacity: Odd PCM data length detected, adjusted samples to %d", totalSamples)
	}

	if totalSamples == 0 {
		log.Printf("[ERROR] CalculateCapacity: No samples found in audio data")
		return nil, errors.New("no audio samples found in the file")
	}

	// Calculate capacities for different LSB methods (in bytes)
	capacities := &models.CapacityResult{
		OneLSB:   (totalSamples * 1) / 8, // 1 bit per sample / 8 bits per byte
		TwoLSB:   (totalSamples * 2) / 8, // 2 bits per sample / 8 bits per byte
		ThreeLSB: (totalSamples * 3) / 8, // 3 bits per sample / 8 bits per byte
		FourLSB:  (totalSamples * 4) / 8, // 4 bits per sample / 8 bits per byte
	}

	duration := time.Since(start)
	log.Printf("[INFO] CalculateCapacity: Completed successfully (total_samples: %d, capacities: 1LSB=%d, 2LSB=%d, duration: %v)",
		totalSamples, capacities.OneLSB, capacities.TwoLSB, duration)

	return capacities, nil
}

// EmbedMessage embeds a secret message into audio data
func (s *steganographyService) EmbedMessage(req *models.EmbedRequest, secretData []byte, metadata []byte) ([]byte, float64, error) {
	start := time.Now()
	log.Printf("[DEBUG] EmbedMessage: Starting embed operation (audio_size: %d bytes, secret_size: %d bytes, metadata_size: %d bytes, nLSB: %d, encryption: %t, random_start: %t)",
		len(req.CoverAudio), len(secretData), len(metadata), req.NLsb, req.UseEncryption, req.UseRandomStart)

	// Decode MP3 to PCM
	audioReader := bytes.NewReader(req.CoverAudio)
	decoder, err := mp3.NewDecoder(audioReader)
	if err != nil {
		log.Printf("[ERROR] EmbedMessage: Failed to decode MP3: %v", err)
		return nil, 0, errors.New("failed to decode MP3: " + err.Error())
	}

	pcmData, err := io.ReadAll(decoder)
	if err != nil {
		log.Printf("[ERROR] EmbedMessage: Failed to read PCM data: %v", err)
		return nil, 0, errors.New("failed to read PCM data: " + err.Error())
	}

	log.Printf("[DEBUG] EmbedMessage: Successfully decoded MP3 to PCM (pcm_size: %d bytes)", len(pcmData))

	capacityResult, err := s.CalculateCapacity(req.CoverAudio)
	if err != nil {
		log.Printf("[ERROR] EmbedMessage: Failed to calculate capacity: %v", err)
		return nil, 0, fmt.Errorf("failed to calculate capacity: %v", err)
	}

	metadataSize := len(metadata)
	secretSize := len(secretData)
	totalDataSize := metadataSize + secretSize

	// Get the appropriate capacity based on nLSB
	var maxCapacity int
	switch req.NLsb {
	case 1:
		maxCapacity = capacityResult.OneLSB
	case 2:
		maxCapacity = capacityResult.TwoLSB
	case 3:
		maxCapacity = capacityResult.ThreeLSB
	case 4:
		maxCapacity = capacityResult.FourLSB
	default:
		log.Printf("[ERROR] EmbedMessage: Invalid nLSB value: %d", req.NLsb)
		return nil, 0, fmt.Errorf("invalid nLSB value: %d (must be 1-4)", req.NLsb)
	}

	log.Printf("[DEBUG] EmbedMessage: Capacity check (data_size: %d bytes, max_capacity: %d bytes for %d-LSB)",
		totalDataSize, maxCapacity, req.NLsb)

	if totalDataSize > maxCapacity {
		log.Printf("[ERROR] EmbedMessage: Insufficient capacity - need %d bytes, only %d bytes available", totalDataSize, maxCapacity)
		return nil, 0, fmt.Errorf("insufficient capacity: need %d bytes, but only %d bytes available", totalDataSize, maxCapacity)
	}

	originalPCM := make([]byte, len(pcmData))
	copy(originalPCM, pcmData)

	// Apply encryption if requested
	dataToEmbed := secretData
	if req.UseEncryption && req.StegoKey != "" {
		log.Printf("[DEBUG] EmbedMessage: Applying Vigenère cipher encryption with key length: %d", len(req.StegoKey))
		cryptoSvc := NewCryptographyService()
		dataToEmbed = cryptoSvc.VigenereCipher(secretData, req.StegoKey, true)
		log.Printf("[DEBUG] EmbedMessage: Data encrypted (original: %d bytes, encrypted: %d bytes)", len(secretData), len(dataToEmbed))
	}

	// Combine metadata and (possibly encrypted) secret data
	finalDataToEmbed := append(metadata, dataToEmbed...)
	log.Printf("[DEBUG] EmbedMessage: Final data to embed: %d bytes (metadata: %d + data: %d)", len(finalDataToEmbed), len(metadata), len(dataToEmbed))

	// Convert to bit array
	bitArray := bytesToBits(finalDataToEmbed)
	log.Printf("[DEBUG] EmbedMessage: Converted to bit array: %d bits", len(bitArray))

	// Calculate starting position
	totalSamples := len(pcmData) / 2
	var startPos int
	if req.UseRandomStart && req.StegoKey != "" {
		startPos = generateRandomStart(req.StegoKey, totalSamples, len(bitArray), req.NLsb)
		log.Printf("[DEBUG] EmbedMessage: Using random start position: %d", startPos)
	} else {
		startPos = 0
		log.Printf("[DEBUG] EmbedMessage: Using sequential start position: %d", startPos)
	}

	// Embed bits into PCM samples
	log.Printf("[DEBUG] EmbedMessage: Starting LSB embedding...")
	err = embedBitsIntoSamples(pcmData, bitArray, startPos, req.NLsb)
	if err != nil {
		log.Printf("[ERROR] EmbedMessage: Failed to embed bits: %v", err)
		return nil, 0, err
	}
	log.Printf("[DEBUG] EmbedMessage: LSB embedding completed successfully")

	// Calculate PSNR
	audioSvc := NewAudioService()
	psnr := audioSvc.CalculatePSNR(originalPCM, pcmData)
	log.Printf("[INFO] EmbedMessage: PSNR calculated: %.2f dB", psnr)

	// Encode modified PCM back to WAV format (preserves LSB steganography)
	log.Printf("[DEBUG] EmbedMessage: Encoding to WAV format...")
	encoder := NewAudioEncoder()
	wavData, err := encoder.EncodeToWAV(pcmData, decoder.SampleRate())
	if err != nil {
		log.Printf("[ERROR] EmbedMessage: Failed to encode to WAV: %v", err)
		return nil, 0, fmt.Errorf("failed to encode audio to WAV: %v", err)
	}

	duration := time.Since(start)
	log.Printf("[INFO] EmbedMessage: Embed operation completed successfully (output_size: %d bytes, psnr: %.2f dB, duration: %v)",
		len(wavData), psnr, duration)

	return wavData, psnr, nil
}

// ExtractMessage extracts a secret message from audio data
func (s *steganographyService) ExtractMessage(req *models.ExtractRequest, audioData []byte) ([]byte, string, error) {
	start := time.Now()
	log.Printf("[DEBUG] ExtractMessage: Starting extraction operation (audio_size: %d bytes, use_encryption: %t, use_random_start: %t)",
		len(req.StegoAudio), req.UseEncryption, req.UseRandomStart)

	var pcmData []byte
	var err error

	// Try to decode as our custom steganographic MP3 format first
	if len(req.StegoAudio) > 20 && string(req.StegoAudio[:3]) == "ID3" {
		log.Printf("[DEBUG] ExtractMessage: Detected steganographic MP3 format")
		encoder := NewAudioEncoder()
		pcmData, err = encoder.extractFromSteganographicMP3(req.StegoAudio)
		if err != nil {
			log.Printf("[DEBUG] ExtractMessage: Steganographic MP3 extraction failed, trying standard MP3: %v", err)
		} else {
			log.Printf("[DEBUG] ExtractMessage: Successfully extracted PCM from steganographic MP3 (pcm_size: %d bytes)", len(pcmData))
		}
	}

	// If steganographic MP3 extraction failed or wasn't detected, try standard MP3
	if pcmData == nil {
		audioReader := bytes.NewReader(req.StegoAudio)
		decoder, err := mp3.NewDecoder(audioReader)
		if err == nil {
			pcmData, err = io.ReadAll(decoder)
			if err != nil {
				log.Printf("[ERROR] ExtractMessage: Failed to read MP3 PCM data: %v", err)
				return nil, "", errors.New("failed to read MP3 PCM data: " + err.Error())
			}
			log.Printf("[DEBUG] ExtractMessage: Successfully decoded MP3 to PCM (pcm_size: %d bytes)", len(pcmData))
		} else {
			log.Printf("[DEBUG] ExtractMessage: MP3 decoding failed, trying WAV format: %v", err)
			// Try to handle as WAV format
			if len(req.StegoAudio) > 44 && string(req.StegoAudio[:4]) == "RIFF" && string(req.StegoAudio[8:12]) == "WAVE" {
				// Parse WAV header to find data chunk
				dataOffset, dataSize, parseErr := parseWAVHeader(req.StegoAudio)
				if parseErr != nil {
					log.Printf("[ERROR] ExtractMessage: Failed to parse WAV header: %v", parseErr)
					return nil, "", fmt.Errorf("failed to parse WAV header: %v", parseErr)
				}

				if dataOffset+int(dataSize) > len(req.StegoAudio) {
					log.Printf("[ERROR] ExtractMessage: Invalid WAV file - data chunk extends beyond file")
					return nil, "", errors.New("invalid WAV file structure")
				}

				pcmData = req.StegoAudio[dataOffset : dataOffset+int(dataSize)]
				log.Printf("[DEBUG] ExtractMessage: Successfully parsed WAV format (pcm_size: %d bytes, data_offset: %d)", len(pcmData), dataOffset)
			} else {
				log.Printf("[ERROR] ExtractMessage: Unable to decode audio format (not MP3, steganographic MP3, or WAV)")
				return nil, "", errors.New("unsupported audio format: file must be MP3 or WAV")
			}
		}
	}

	totalSamples := len(pcmData) / 2

	// Extract metadata first (fixed size: 56 bytes for current format)
	log.Printf("[DEBUG] Stego key: %s", req.StegoKey)
	metadataBits := 56 * 8 // 56 bytes = 448 bits
	startPos := 0
	if req.UseRandomStart && req.StegoKey != "" {
		startPos = generateRandomStart(req.StegoKey, totalSamples, metadataBits, req.NLsb)
		log.Printf("[DEBUG] ExtractMessage: Using random start position: %d", startPos)
	}

	log.Printf("[DEBUG] ExtractMessage: Extracting metadata from position %d (%d bits)", startPos, metadataBits)
	metadataBitArray := extractBitsFromSamples(pcmData, startPos, req.NLsb, metadataBits)
	metadata := bitsToBytes(metadataBitArray)

	// Parse metadata to get file info
	filename, fileSize, useEncryption, _, err := parseMetadata(metadata)
	if err != nil {
		log.Printf("[ERROR] ExtractMessage: Failed to parse metadata: %v", err)
		return nil, "", fmt.Errorf("failed to parse metadata: %v", err)
	}

	log.Printf("[DEBUG] ExtractMessage: Parsed metadata - filename: '%s', file_size: %d, encryption: %t", filename, fileSize, useEncryption)

	// Extract secret data
	currentBitPos := metadataBits
	secretBits := extractBitsFromSamples(pcmData, startPos, req.NLsb, currentBitPos+fileSize*8)
	secretData := bitsToBytes(secretBits[currentBitPos:])

	// Apply decryption if needed
	if useEncryption && req.StegoKey != "" {
		cryptoSvc := NewCryptographyService()
		secretData = cryptoSvc.VigenereCipher(secretData, req.StegoKey, false)
	}

	// Use provided filename or extracted filename
	outputFilename := req.OutputFilename
	if outputFilename == "" {
		if filename != "" {
			outputFilename = filename
		} else {
			outputFilename = "extracted_secret.bin"
		}
	}

	duration := time.Since(start)
	log.Printf("[INFO] ExtractMessage: Extraction completed successfully (extracted_size: %d bytes, filename: '%s', duration: %v)",
		len(secretData), outputFilename, duration)

	return secretData, outputFilename, nil
}

// CreateMetadata creates metadata for embedding
func (s *steganographyService) CreateMetadata(filename string, fileSize int, useEncryption, useRandomStart bool, nLsb int) []byte {
	log.Printf("[DEBUG] CreateMetadata: Creating metadata (filename: '%s', file_size: %d, encryption: %t, random_start: %t, nLSB: %d)",
		filename, fileSize, useEncryption, useRandomStart, nLsb)

	metadata := make([]byte, 56) // Fixed size metadata

	// Filename (first 32 bytes, null-padded)
	if len(filename) > 32 {
		filename = filename[:32]
	}
	copy(metadata[:32], filename)

	// File size (4 bytes, big-endian)
	metadata[32] = byte(fileSize >> 24)
	metadata[33] = byte(fileSize >> 16)
	metadata[34] = byte(fileSize >> 8)
	metadata[35] = byte(fileSize)

	// Flags (1 byte)
	var flags byte
	if useEncryption {
		flags |= 0x01
	}
	if useRandomStart {
		flags |= 0x02
	}
	metadata[36] = flags

	// n-LSB value (1 byte)
	metadata[37] = byte(nLsb)

	// Reserved space for future use (18 bytes)
	// metadata[38:56] remains zero-filled

	log.Printf("[DEBUG] CreateMetadata: Metadata created successfully (size: %d bytes)", len(metadata))
	return metadata
}
