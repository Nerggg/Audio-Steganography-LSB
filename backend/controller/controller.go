package controller

import (
    "bytes"
    "io"
    "errors"
	"encoding/binary"
	"math"

    "github.com/hajimehoshi/go-mp3"
    "github.com/Nerggg/Audio-Steganography-LSB/backend/models"
)

func CalculateCapacity(audioData []byte) (*models.CapacityResult, error) {
    audioReader := bytes.NewReader(audioData)

    decoder, err := mp3.NewDecoder(audioReader)
    if err != nil {
        return nil, models.ErrInvalidMP3
    }

    pcmData, err := io.ReadAll(decoder)
    if err != nil {
        return nil, errors.New("could not read decoded audio stream: " + err.Error())
    }

    totalSamples := len(pcmData) / 2

    if len(pcmData)%2 != 0 {
        totalSamples = (len(pcmData) - 1) / 2
    }

    if totalSamples == 0 {
        return nil, models.ErrInvalidMP3
    }

    capacities := &models.CapacityResult{
        OneLSB:   (totalSamples * 1) / 8,
        TwoLSB:   (totalSamples * 2) / 8,
        ThreeLSB: (totalSamples * 3) / 8,
        FourLSB:  (totalSamples * 4) / 8,
    }

    return capacities, nil
}

// createMetadata creates metadata to be embedded with the secret file
func CreateMetadata(filename string, fileSize int, useEncryption, useRandomStart bool, nLsb int) []byte {
	var metadata bytes.Buffer

	// Magic signature (4 bytes)
	metadata.Write([]byte("STEG"))

	// Flags (1 byte): bit 0=encryption, bit 1=random start, bits 2-3=nLsb-1
	var flags byte
	if useEncryption {
		flags |= 0x01
	}
	if useRandomStart {
		flags |= 0x02
	}
	flags |= byte((nLsb-1)<<2) & 0x0C

	metadata.WriteByte(flags)

	// File size (4 bytes, little endian)
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, uint32(fileSize))
	metadata.Write(sizeBytes)

	// Filename length (1 byte) and filename
	filenameBytes := []byte(filename)
	if len(filenameBytes) > 255 {
		filenameBytes = filenameBytes[:255]
	}
	metadata.WriteByte(byte(len(filenameBytes)))
	metadata.Write(filenameBytes)

	// End signature (4 bytes)
	metadata.Write([]byte("GEND"))

	return metadata.Bytes()
}

// embedMessage embeds the secret message into the cover audio
func EmbedMessage(req *models.EmbedRequest, secretData, metadata []byte) ([]byte, float64, error) {
	// Decode MP3 to PCM
	audioReader := bytes.NewReader(req.CoverAudio)
	decoder, err := mp3.NewDecoder(audioReader)
	if err != nil {
		return nil, 0, errors.New("failed to decode MP3: " + err.Error())
	}

	pcmData, err := io.ReadAll(decoder)
	if err != nil {
		return nil, 0, errors.New("failed to read PCM data: " + err.Error())
	}

	originalPCM := make([]byte, len(pcmData))
	copy(originalPCM, pcmData)

	// Combine metadata and secret data
	dataToEmbed := append(metadata, secretData...)

	// Convert to bit array
	bitArray := bytesToBits(dataToEmbed)

	// Calculate starting position
	startPos := 0
	if req.UseRandomStart {
		startPos = generateRandomStart(req.StegoKey, len(pcmData)/2, len(bitArray), req.NLsb)
	}

	// Embed bits into PCM samples
	err = embedBitsIntoSamples(pcmData, bitArray, startPos, req.NLsb)
	if err != nil {
		return nil, 0, err
	}

	// Calculate PSNR
	psnr := calculatePSNR(originalPCM, pcmData)

	// For this implementation, we'll return the modified PCM as raw data
	// In a real implementation, you'd need to encode back to MP3
	// This is a simplified approach - you might need a proper MP3 encoder
	return pcmData, psnr, nil
}

// bytesToBits converts byte array to bit array
func bytesToBits(data []byte) []int {
	var bits []int
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bits = append(bits, int((b>>i)&1))
		}
	}
	return bits
}

// generateRandomStart generates a random starting position based on the seed
func generateRandomStart(seed string, totalSamples, bitsToEmbed, nLsb int) int {
	// Convert seed to number
	var seedValue int64
	for _, char := range seed {
		seedValue += int64(char)
	}

	// Calculate maximum safe starting position
	maxSafeStart := totalSamples - (bitsToEmbed / nLsb)
	if maxSafeStart <= 0 {
		return 0
	}

	// Use seed to generate pseudo-random position
	return int(seedValue % int64(maxSafeStart))
}

// embedBitsIntoSamples embeds bits into audio samples using LSB
func embedBitsIntoSamples(pcmData []byte, bits []int, startPos, nLsb int) error {
	sampleIndex := startPos
	bitIndex := 0

	for bitIndex < len(bits) && (sampleIndex+1)*2 <= len(pcmData) {
		// Get 16-bit sample (little endian)
		sample := int16(binary.LittleEndian.Uint16(pcmData[sampleIndex*2 : sampleIndex*2+2]))

		// Embed n bits into this sample
		for i := 0; i < nLsb && bitIndex < len(bits); i++ {
			// Clear the LSB bit
			sample &= ^(1 << i)
			// Set the bit from our data
			sample |= int16(bits[bitIndex]) << i
			bitIndex++
		}

		// Write sample back (little endian)
		binary.LittleEndian.PutUint16(pcmData[sampleIndex*2:sampleIndex*2+2], uint16(sample))
		sampleIndex++
	}

	if bitIndex < len(bits) {
		return errors.New("not enough samples to embed all data")
	}

	return nil
}

// calculatePSNR calculates Peak Signal-to-Noise Ratio
func calculatePSNR(original, modified []byte) float64 {
	if len(original) != len(modified) {
		return 0
	}

	var mse float64
	maxVal := 65535.0 // 16-bit audio max value

	// Calculate Mean Squared Error for 16-bit samples
	for i := 0; i+1 < len(original); i += 2 {
		origSample := int16(binary.LittleEndian.Uint16(original[i : i+2]))
		modSample := int16(binary.LittleEndian.Uint16(modified[i : i+2]))
		diff := float64(origSample - modSample)
		mse += diff * diff
	}

	numSamples := float64(len(original) / 2)
	mse /= numSamples

	if mse == 0 {
		return math.Inf(1) // Perfect match
	}

	psnr := 20 * math.Log10(maxVal/math.Sqrt(mse))
	return psnr
}

// vigenereCipher performs Vigenère cipher encryption/decryption
func VigenereCipher(data []byte, key string, encrypt bool) []byte {
	if len(key) == 0 {
		return data
	}

	keyBytes := []byte(key)
	result := make([]byte, len(data))

	for i, b := range data {
		keyByte := keyBytes[i%len(keyBytes)]
		if encrypt {
			result[i] = b ^ keyByte
		} else {
			result[i] = b ^ keyByte
		}
	}

	return result
}
