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

		// For capacity calculation, get the PCM data
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

	// Get PCM data from input (MP3 decode or WAV data chunk)
	var pcmData []byte
	var originalPCM []byte
	var isWAV bool
	var wavDataCopy []byte
	var sampleRateForEncode int

	// Try as MP3 first
	audioReader := bytes.NewReader(req.CoverAudio)
	decoder, err := mp3.NewDecoder(audioReader)
	if err == nil {
		// Successfully decoded as MP3
		pcmData, err = io.ReadAll(decoder)
		if err != nil {
			log.Printf("[ERROR] EmbedMessage: Failed to read PCM data from MP3: %v", err)
			return nil, 0, errors.New("failed to read PCM data: " + err.Error())
		}
		log.Printf("[DEBUG] EmbedMessage: Successfully decoded MP3 to PCM (pcm_size: %d bytes)", len(pcmData))
		// Keep a copy for PSNR
		originalPCM = make([]byte, len(pcmData))
		copy(originalPCM, pcmData)
		sampleRateForEncode = decoder.SampleRate()
		isWAV = false
	} else {
		// Not MP3; try WAV path by parsing header and getting a slice to data chunk
		log.Printf("[DEBUG] EmbedMessage: MP3 decode failed, trying WAV path: %v", err)
		dataOffset, dataSize, parseErr := parseWAVHeader(req.CoverAudio)
		if parseErr != nil {
			log.Printf("[ERROR] EmbedMessage: Unsupported audio format (not MP3 or WAV): %v", parseErr)
			return nil, 0, errors.New("unsupported audio format: only MP3 or WAV are supported")
		}

		// Create a working copy of the entire WAV file so we never mutate the input in-place
		wavDataCopy = make([]byte, len(req.CoverAudio))
		copy(wavDataCopy, req.CoverAudio)

		// Obtain a slice referencing ONLY the data chunk; edits will affect wavDataCopy's data chunk
		end := dataOffset + int(dataSize)
		if end > len(wavDataCopy) {
			log.Printf("[ERROR] EmbedMessage: WAV data chunk exceeds file bounds (offset=%d size=%d len=%d)", dataOffset, dataSize, len(wavDataCopy))
			return nil, 0, errors.New("invalid WAV structure: data chunk out of bounds")
		}
		pcmData = wavDataCopy[dataOffset:end]

		// Prepare original PCM for PSNR (copy from original input's data chunk)
		originalPCM = make([]byte, len(pcmData))
		copy(originalPCM, req.CoverAudio[dataOffset:end])

		// In WAV path, we'll return the modified WAV with headers untouched
		isWAV = true
	}

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

	// Convert to bit arrays for metadata and payload
	metadataBits := bytesToBits(metadata)
	payloadBits := bytesToBits(dataToEmbed)
	log.Printf("[DEBUG] EmbedMessage: Bit sizes -> metadata: %d, payload: %d", len(metadataBits), len(payloadBits))

	// Determine layout: always put metadata at the start (sample 0), and payload after a randomized offset (if enabled)
	totalSamples := len(pcmData) / 2
	metaSamples := samplesNeeded(len(metadataBits), req.NLsb)
	payloadSamples := samplesNeeded(len(payloadBits), req.NLsb)

	// Capacity re-check in sample terms
	if metaSamples+payloadSamples > totalSamples {
		return nil, 0, fmt.Errorf("insufficient capacity: need %d samples (meta %d + payload %d), have %d", metaSamples+payloadSamples, metaSamples, payloadSamples, totalSamples)
	}

	// Embed metadata at start position 0
	if err = embedBitsIntoSamples(pcmData, metadataBits, 0, req.NLsb); err != nil {
		log.Printf("[ERROR] EmbedMessage: Failed to embed metadata: %v", err)
		return nil, 0, err
	}

	// Compute payload start
	startPos := metaSamples
	if req.UseRandomStart && req.StegoKey != "" {
		pos := generatePayloadStart(req.StegoKey, totalSamples, metaSamples, payloadSamples)
		if pos < 0 {
			return nil, 0, fmt.Errorf("insufficient capacity for randomized payload placement")
		}
		startPos = pos
	}
	log.Printf("[DEBUG] EmbedMessage: Payload start position: %d (metaSamples=%d, totalSamples=%d)", startPos, metaSamples, totalSamples)

	// Embed payload bits at startPos
	log.Printf("[DEBUG] EmbedMessage: Starting LSB embedding for payload...")
	err = embedBitsIntoSamples(pcmData, payloadBits, startPos, req.NLsb)
	if err != nil {
		log.Printf("[ERROR] EmbedMessage: Failed to embed payload bits: %v", err)
		return nil, 0, err
	}
	log.Printf("[DEBUG] EmbedMessage: LSB embedding completed successfully (meta at 0, payload at %d)", startPos)

	// Calculate PSNR using original vs modified PCM
	audioSvc := NewAudioService()
	psnr := audioSvc.CalculatePSNR(originalPCM, pcmData)
	log.Printf("[INFO] EmbedMessage: PSNR calculated: %.2f dB", psnr)

	// Return output depending on input type
	if isWAV {
		// For WAV input, return the entire WAV file with only the data chunk modified
		duration := time.Since(start)
		log.Printf("[INFO] EmbedMessage: Embed (WAV) completed successfully (output_size: %d bytes, psnr: %.2f dB, duration: %v)",
			len(wavDataCopy), psnr, duration)
		return wavDataCopy, psnr, nil
	}

	// For MP3 input, encode the modified PCM to WAV
	log.Printf("[DEBUG] EmbedMessage: Encoding modified PCM to WAV (from MP3 input)...")
	encoder := NewAudioEncoder()
	wavData, err := encoder.EncodeToWAV(pcmData, sampleRateForEncode)
	if err != nil {
		log.Printf("[ERROR] EmbedMessage: Failed to encode to WAV: %v", err)
		return nil, 0, fmt.Errorf("failed to encode audio to WAV: %v", err)
	}

	duration := time.Since(start)
	log.Printf("[INFO] EmbedMessage: Embed (MP3->WAV) completed successfully (output_size: %d bytes, psnr: %.2f dB, duration: %v)",
		len(wavData), psnr, duration)

	return wavData, psnr, nil
}

// ExtractMessage extracts a secret message from audio data with auto-detection of parameters
func (s *steganographyService) ExtractMessage(req *models.ExtractRequest, audioData []byte) ([]byte, string, error) {
	start := time.Now()
	log.Printf("[DEBUG] ExtractMessage: Starting extraction operation (audio_size: %d bytes, stego_key_provided: %t)",
		len(req.StegoAudio), req.StegoKey != "")

	var pcmData []byte

	// Try to decode as standard MP3 format
	{
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
				log.Printf("[ERROR] ExtractMessage: Unable to decode audio format (not MP3 or WAV)")
				return nil, "", errors.New("unsupported audio format: file must be MP3 or WAV")
			}
		}
	}

	totalSamples := len(pcmData) / 2
	metadataBits := 56 * 8

	var filename string
	var fileSize int
	var useEncryption bool
	var useRandomStart bool
	var detectedNLsb int

	for tryNLsb := 1; tryNLsb <= 4; tryNLsb++ {
		metaBits := extractBitsFromSamples(pcmData, 0, tryNLsb, metadataBits)
		md := bitsToBytes(metaBits)
		f, sz, enc, rnd, err := parseMetadata(md)
		if err != nil {
			continue
		}
		if sz <= 0 || sz > 50*1024*1024 {
			continue
		}
		// basic capacity check
		metaSamples := samplesNeeded(metadataBits, tryNLsb)
		payloadSamples := samplesNeeded(sz*8, tryNLsb)
		if metaSamples+payloadSamples > totalSamples {
			continue
		}
		filename, fileSize, useEncryption, useRandomStart = f, sz, enc, rnd
		detectedNLsb = tryNLsb
		break
	}

	if detectedNLsb == 0 {
		return nil, "", fmt.Errorf("failed to auto-detect metadata: unsupported or corrupted stego audio")
	}

	// Compute payload start (deterministic if random start was used during embed)
	metaSamples := samplesNeeded(metadataBits, detectedNLsb)
	payloadSamples := samplesNeeded(fileSize*8, detectedNLsb)
	startPos := metaSamples
	if useRandomStart && req.StegoKey == "" {
		return nil, "", fmt.Errorf("stego key is required to extract data when random start was used")
	}
	if useRandomStart && req.StegoKey != "" {
		pos := generatePayloadStart(req.StegoKey, totalSamples, metaSamples, payloadSamples)
		if pos < 0 {
			return nil, "", fmt.Errorf("invalid stego layout: not enough samples for payload")
		}
		startPos = pos
	}

	// Extract payload directly
	secretBits := extractBitsFromSamples(pcmData, startPos, detectedNLsb, fileSize*8)
	secretData := bitsToBytes(secretBits)

	// Apply decryption if needed
	if useEncryption && req.StegoKey != "" {
		cryptoSvc := NewCryptographyService()
		secretData = cryptoSvc.VigenereCipher(secretData, req.StegoKey, false)
	}

	// Use provided filename or extracted filename with extension auto-detection
	outputFilename := req.OutputFilename
	if outputFilename == "" {
		if filename != "" {
			outputFilename = filename
			// Ensure the filename has an extension by auto-detecting file type if needed
			if !hasExtension(outputFilename) {
				detectedExt := detectFileExtension(secretData)
				if detectedExt != "" {
					outputFilename += detectedExt
					log.Printf("[DEBUG] ExtractMessage: Added detected extension '%s' to filename '%s'", detectedExt, filename)
				}
			}
		} else {
			// Try to detect file type and provide appropriate extension
			detectedExt := detectFileExtension(secretData)
			if detectedExt != "" {
				outputFilename = "extracted_secret" + detectedExt
			} else {
				outputFilename = "extracted_secret.bin"
			}
			log.Printf("[DEBUG] ExtractMessage: Generated filename '%s' with detected extension", outputFilename)
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

	log.Printf("[DEBUG] CreateMetadata: Metadata created successfully (size: %d bytes)", len(metadata))
	return metadata
}
