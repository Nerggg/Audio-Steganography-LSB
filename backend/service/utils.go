package service

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"log"
	"strings"
)

// bytesToBits converts byte array to bit array
func bytesToBits(data []byte) []int {
	bits := make([]int, len(data)*8)

	for i, b := range data {
		for j := 0; j < 8; j++ {
			if (b & (1 << (7 - j))) != 0 {
				bits[i*8+j] = 1
			} else {
				bits[i*8+j] = 0
			}
		}
	}

	return bits
}

// bitsToBytes converts bit array back to byte array
func bitsToBytes(bits []int) []byte {
	// Ensure we have complete bytes
	byteCount := (len(bits) + 7) / 8
	data := make([]byte, byteCount)

	for i := 0; i < len(bits); i++ {
		byteIndex := i / 8
		bitIndex := i % 8

		if bits[i] == 1 {
			data[byteIndex] |= (1 << (7 - bitIndex))
		}
	}

	return data
}

// generateRandomStart generates a random starting position based on the seed
func generateRandomStart(seed string, totalSamples, bitsToEmbed, nLsb int) int {
	// Use FNV hash for consistent pseudo-random generation
	hasher := fnv.New32a()
	hasher.Write([]byte(seed))
	hashValue := hasher.Sum32()

	// Calculate maximum safe starting position
	maxSafeBits := totalSamples * nLsb
	maxSafeStart := maxSafeBits - bitsToEmbed

	if maxSafeStart <= 0 {
		log.Printf("[WARN] generateRandomStart: No safe random position available, using start position 0")
		return 0
	}

	// Convert to sample position (bits to samples)
	maxSafeSampleStart := maxSafeStart / nLsb

	// Generate random start position within safe bounds
	randomStart := int(hashValue) % maxSafeSampleStart
	if randomStart < 0 {
		randomStart = -randomStart
	}

	log.Printf("[DEBUG] generateRandomStart: Generated position %d (max_safe: %d, total_samples: %d, bits_to_embed: %d)",
		randomStart, maxSafeSampleStart, totalSamples, bitsToEmbed)

	return randomStart
}

// embedBitsIntoSamples embeds bits into audio samples using LSB
func embedBitsIntoSamples(pcmData []byte, bits []int, startPos, nLsb int) error {
	totalSamples := len(pcmData) / 2 // 16-bit samples
	bitsNeeded := len(bits)
	maxBitsAvailable := (totalSamples - startPos) * nLsb

	if bitsNeeded > maxBitsAvailable {
		return fmt.Errorf("insufficient capacity: need %d bits, but only %d bits available from position %d",
			bitsNeeded, maxBitsAvailable, startPos)
	}

	bitIndex := 0
	for sampleIndex := startPos; sampleIndex < totalSamples && bitIndex < len(bits); sampleIndex++ {
		// Calculate byte offset for this sample (2 bytes per 16-bit sample)
		byteOffset := sampleIndex * 2

		// Extract 16-bit sample (little-endian)
		sample := binary.LittleEndian.Uint16(pcmData[byteOffset : byteOffset+2])

		// Embed up to nLsb bits in this sample
		for lsbPos := 0; lsbPos < nLsb && bitIndex < len(bits); lsbPos++ {
			// Clear the target bit
			sample &= ^(uint16(1) << lsbPos)

			// Set the bit if needed
			if bits[bitIndex] == 1 {
				sample |= (uint16(1) << lsbPos)
			}

			bitIndex++
		}

		// Write the modified sample back
		binary.LittleEndian.PutUint16(pcmData[byteOffset:byteOffset+2], sample)
	}

	log.Printf("[DEBUG] embedBitsIntoSamples: Successfully embedded %d bits starting from sample %d using %d-LSB",
		bitIndex, startPos, nLsb)

	return nil
}

// extractBitsFromSamples extracts bits from audio samples using LSB
func extractBitsFromSamples(pcmData []byte, startPos, nLsb, totalBits int) []int {
	totalSamples := len(pcmData) / 2 // 16-bit samples
	bits := make([]int, totalBits)

	bitIndex := 0
	for sampleIndex := startPos; sampleIndex < totalSamples && bitIndex < totalBits; sampleIndex++ {
		// Calculate byte offset for this sample
		byteOffset := sampleIndex * 2

		// Extract 16-bit sample (little-endian)
		sample := binary.LittleEndian.Uint16(pcmData[byteOffset : byteOffset+2])

		// Extract up to nLsb bits from this sample
		for lsbPos := 0; lsbPos < nLsb && bitIndex < totalBits; lsbPos++ {
			// Extract the bit
			if (sample & (uint16(1) << lsbPos)) != 0 {
				bits[bitIndex] = 1
			} else {
				bits[bitIndex] = 0
			}

			bitIndex++
		}
	}

	log.Printf("[DEBUG] extractBitsFromSamples: Extracted %d bits starting from sample %d using %d-LSB",
		bitIndex, startPos, nLsb)

	return bits
}

// parseMetadata parses metadata extracted from embedded data
func parseMetadata(metadata []byte) (filename string, fileSize int, useEncryption, useRandomStart bool, err error) {
	if len(metadata) < 38 {
		return "", 0, false, false, fmt.Errorf("metadata too short: expected at least 38 bytes, got %d", len(metadata))
	}

	// Extract filename (first 32 bytes, null-terminated)
	filenameBytes := metadata[:32]
	// Find the first null byte to get the actual filename
	nullIndex := -1
	for i, b := range filenameBytes {
		if b == 0 {
			nullIndex = i
			break
		}
	}

	if nullIndex >= 0 {
		filename = string(filenameBytes[:nullIndex])
	} else {
		filename = string(filenameBytes)
	}

	// Clean up filename (remove any remaining null characters and trim)
	filename = strings.Trim(filename, "\x00")
	filename = strings.TrimSpace(filename)

	// Extract file size (bytes 32-35, big-endian)
	fileSize = int(binary.BigEndian.Uint32(metadata[32:36]))

	// Extract flags (byte 36)
	flags := metadata[36]
	useEncryption = (flags & 0x01) != 0
	useRandomStart = (flags & 0x02) != 0

	// Extract n-LSB value (byte 37)
	nLsb := int(metadata[37])

	log.Printf("[DEBUG] parseMetadata: filename='%s', file_size=%d, encryption=%t, random_start=%t, nLSB=%d",
		filename, fileSize, useEncryption, useRandomStart, nLsb)

	return filename, fileSize, useEncryption, useRandomStart, nil
}

// parseMetadataWithNLsb parses metadata and returns nLSB value as well
func parseMetadataWithNLsb(metadata []byte) (filename string, fileSize int, useEncryption, useRandomStart bool, nLsb int, err error) {
	if len(metadata) < 38 {
		return "", 0, false, false, 0, fmt.Errorf("metadata too short: expected at least 38 bytes, got %d", len(metadata))
	}

	// Extract filename (first 32 bytes, null-terminated)
	filenameBytes := metadata[:32]
	// Find the first null byte to get the actual filename
	nullIndex := -1
	for i, b := range filenameBytes {
		if b == 0 {
			nullIndex = i
			break
		}
	}

	if nullIndex >= 0 {
		filename = string(filenameBytes[:nullIndex])
	} else {
		filename = string(filenameBytes)
	}

	// Clean up filename (remove any remaining null characters and trim)
	filename = strings.Trim(filename, "\x00")
	filename = strings.TrimSpace(filename)

	// Extract file size (bytes 32-35, big-endian)
	fileSize = int(binary.BigEndian.Uint32(metadata[32:36]))

	// Extract flags (byte 36)
	flags := metadata[36]
	useEncryption = (flags & 0x01) != 0
	useRandomStart = (flags & 0x02) != 0

	// Extract n-LSB value (byte 37)
	nLsb = int(metadata[37])

	log.Printf("[DEBUG] parseMetadataWithNLsb: filename='%s', file_size=%d, encryption=%t, random_start=%t, nLSB=%d",
		filename, fileSize, useEncryption, useRandomStart, nLsb)

	return filename, fileSize, useEncryption, useRandomStart, nLsb, nil
}

// parseWAVHeader parses a WAV file header and returns the data chunk offset and size
func parseWAVHeader(wavData []byte) (dataOffset int, dataSize uint32, err error) {
	if len(wavData) < 44 {
		return 0, 0, fmt.Errorf("WAV file too short: need at least 44 bytes, got %d", len(wavData))
	}

	// Check RIFF header
	if string(wavData[:4]) != "RIFF" {
		return 0, 0, fmt.Errorf("invalid WAV file: missing RIFF header")
	}

	// Check WAVE format
	if string(wavData[8:12]) != "WAVE" {
		return 0, 0, fmt.Errorf("invalid WAV file: not WAVE format")
	}

	// Find the data chunk by parsing all chunks
	offset := 12 // Start after "RIFF" + size + "WAVE"

	for offset+8 <= len(wavData) {
		// Read chunk ID and size
		chunkID := string(wavData[offset : offset+4])
		chunkSize := binary.LittleEndian.Uint32(wavData[offset+4 : offset+8])

		log.Printf("[DEBUG] parseWAVHeader: Found chunk '%s' at offset %d, size %d", chunkID, offset, chunkSize)

		if chunkID == "data" {
			dataOffset = offset + 8 // Skip chunk ID and size
			dataSize = chunkSize
			log.Printf("[DEBUG] parseWAVHeader: Found data chunk at offset %d, size %d bytes", dataOffset, dataSize)
			return dataOffset, dataSize, nil
		}

		// Move to next chunk (add 8 for header + chunk size, with padding)
		nextOffset := offset + 8 + int(chunkSize)
		if chunkSize%2 == 1 {
			nextOffset++ // WAV chunks are padded to even byte boundaries
		}

		if nextOffset <= offset {
			return 0, 0, fmt.Errorf("invalid WAV file: infinite loop detected in chunk parsing")
		}

		offset = nextOffset
	}

	return 0, 0, fmt.Errorf("WAV file does not contain a data chunk")
}
