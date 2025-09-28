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

// samplesNeeded returns how many samples are required to embed totalBits using nLsb per sample.
func samplesNeeded(totalBits, nLsb int) int {
	if nLsb <= 0 {
		return 0
	}
	return (totalBits + nLsb - 1) / nLsb
}

// generatePayloadStart computes a deterministic start sample for payload embedding based on the stego key.
// It reserves the first reservedSamples (used by metadata) and fits payloadSamples fully within totalSamples.
// If fitting is impossible, returns -1.
func generatePayloadStart(seed string, totalSamples, reservedSamples, payloadSamples int) int {
	available := totalSamples - reservedSamples - payloadSamples
	if available < 0 {
		return -1
	}
	hasher := fnv.New32a()
	hasher.Write([]byte(seed))
	hv := hasher.Sum32()
	if available == 0 {
		return reservedSamples
	}
	offset := int(hv % uint32(available+1))
	return reservedSamples + offset
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

// hasExtension checks if a filename has an extension
func hasExtension(filename string) bool {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return i < len(filename)-1 // has extension if there's at least one char after the dot
		}
		if filename[i] == '/' || filename[i] == '\\' {
			break // reached path separator, no extension found
		}
	}
	return false
}

// detectFileExtension detects file type based on file signature (magic bytes) and returns appropriate extension
func detectFileExtension(data []byte) string {
	if len(data) < 4 {
		return ""
	}

	// Check common file signatures
	switch {
	// Images
	case len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8:
		return ".jpg"
	case len(data) >= 8 && string(data[:8]) == "\x89PNG\r\n\x1a\n":
		return ".png"
	case len(data) >= 6 && string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a":
		return ".gif"
	case len(data) >= 2 && string(data[:2]) == "BM":
		return ".bmp"
	case len(data) >= 4 && string(data[:4]) == "RIFF":
		// Could be WAV, WebP, or other RIFF formats
		if len(data) >= 12 && string(data[8:12]) == "WAVE" {
			return ".wav"
		}
		return ".webp" // assume WebP if not WAV

	// Documents
	case len(data) >= 4 && string(data[:4]) == "%PDF":
		return ".pdf"
	case len(data) >= 8 &&
		(data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04) || // ZIP
		(data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x05 && data[3] == 0x06) || // Empty ZIP
		(data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x07 && data[3] == 0x08): // Spanned ZIP
		// Could be ZIP, DOCX, XLSX, etc.
		if len(data) > 30 {
			content := strings.ToLower(string(data[:100]))
			if strings.Contains(content, "word/") {
				return ".docx"
			} else if strings.Contains(content, "xl/") {
				return ".xlsx"
			} else if strings.Contains(content, "ppt/") {
				return ".pptx"
			}
		}
		return ".zip"
	case len(data) >= 8 && data[0] == 0xD0 && data[1] == 0xCF && data[2] == 0x11 && data[3] == 0xE0:
		return ".doc" // or .xls, .ppt - generic Microsoft Office

	// Archives
	case len(data) >= 6 && string(data[:6]) == "Rar!\x1a\x07":
		return ".rar"
	case len(data) >= 4 && data[0] == 0x37 && data[1] == 0x7A && data[2] == 0xBC && data[3] == 0xAF:
		return ".7z"

	// Text/Code files
	case len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF:
		return ".txt" // UTF-8 BOM
	default:
		// Check if it's likely text content
		if isLikelyText(data) {
			return ".txt"
		}
	}

	return "" // Unknown file type
}

// isLikelyText checks if the data appears to be text-based
func isLikelyText(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Sample first 512 bytes or less
	sampleSize := len(data)
	if sampleSize > 512 {
		sampleSize = 512
	}

	textChars := 0
	for i := 0; i < sampleSize; i++ {
		b := data[i]
		// Count printable ASCII characters, tabs, newlines, carriage returns
		if (b >= 32 && b <= 126) || b == 9 || b == 10 || b == 13 {
			textChars++
		}
	}

	// If more than 80% of the sample are text characters, consider it text
	return float64(textChars)/float64(sampleSize) > 0.8
}
