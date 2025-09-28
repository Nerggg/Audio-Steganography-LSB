package service

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os/exec"
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

	// Calculate PSNR using maximum possible value for 16-bit signed samples
	maxValue := 32767.0 // Maximum value for 16-bit signed integer
	psnr := 20 * math.Log10(maxValue/math.Sqrt(mse))

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

// EncodeToMP3 encodes PCM data to MP3 format using ID3 PRIV tag for steganography preservation
func (e *audioEncoder) EncodeToMP3(pcmData []byte, sampleRate int) ([]byte, error) {
	log.Printf("[DEBUG] EncodeToMP3: Starting MP3 encoding with ID3 PRIV steganography (pcm_size: %d bytes, sample_rate: %d)", len(pcmData), sampleRate)

	// Create clean PCM for normal MP3 encoding (without steganographic modifications)
	cleanPCM := make([]byte, len(pcmData))
	copy(cleanPCM, pcmData)

	// Encode clean PCM to WAV first
	wavData, err := e.EncodeToWAV(cleanPCM, sampleRate)
	if err != nil {
		log.Printf("[ERROR] EncodeToMP3: Failed to encode WAV: %v", err)
		return nil, fmt.Errorf("failed to encode WAV: %v", err)
	}

	// Convert WAV to MP3 using ffmpeg
	mp3Data, err := e.convertWAVToMP3WithFFmpeg(wavData)
	if err != nil {
		log.Printf("[WARN] EncodeToMP3: ffmpeg conversion failed, falling back to WAV: %v", err)
		// Fallback to WAV if MP3 encoding fails
		return wavData, nil
	}

	// Embed the original steganographic PCM data in ID3 PRIV tag
	const privOwner = "stego/lsb-pcm"
	resultMP3, err := e.EmbedPayloadInMP3(mp3Data, privOwner, pcmData)
	if err != nil {
		log.Printf("[ERROR] EncodeToMP3: Failed to embed PRIV payload: %v", err)
		return nil, fmt.Errorf("failed to embed steganographic data: %v", err)
	}

	log.Printf("[INFO] EncodeToMP3: Successfully encoded to MP3 with ID3 PRIV steganography (size: %d bytes)", len(resultMP3))
	return resultMP3, nil
}

// convertWAVToMP3WithFFmpeg converts WAV data to MP3 using ffmpeg
func (e *audioEncoder) convertWAVToMP3WithFFmpeg(wavData []byte) ([]byte, error) {
	log.Printf("[DEBUG] convertWAVToMP3WithFFmpeg: Starting ffmpeg conversion (input: %d bytes)", len(wavData))

	var stdout, stderr bytes.Buffer

	// Create ffmpeg command: input from stdin, output to stdout
	cmd := exec.Command("ffmpeg",
		"-f", "wav", // Input format
		"-i", "-", // Input from stdin
		"-codec:a", "mp3", // Audio codec
		"-b:a", "320k", // Bitrate
		"-f", "mp3", // Output format
		"-") // Output to stdout

	cmd.Stdin = bytes.NewReader(wavData)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg encoding failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// ConvertWAVToMP3 converts WAV bytes to MP3 bytes using ffmpeg (public method per interface)
func (e *audioEncoder) ConvertWAVToMP3(wavData []byte) ([]byte, error) {
	return e.convertWAVToMP3WithFFmpeg(wavData)
}

// synchsafeEncode encodes a 32-bit integer into a 28-bit synchsafe integer (for ID3v2 tag size)
func synchsafeEncode(v uint32) [4]byte {
	// Only lowest 28 bits are used, each 7 bits per byte
	var out [4]byte
	out[0] = byte((v >> 21) & 0x7F)
	out[1] = byte((v >> 14) & 0x7F)
	out[2] = byte((v >> 7) & 0x7F)
	out[3] = byte(v & 0x7F)
	return out
}

// buildID3v23PrivTag builds a minimal ID3v2.3 tag containing a single PRIV frame with the given owner and payload.
func buildID3v23PrivTag(owner string, payload []byte) []byte {
	// Frame payload = owner (ISO-8859-1) + 0x00 + private data
	ownerBytes := []byte(owner)
	frameData := make([]byte, 0, len(ownerBytes)+1+len(payload))
	frameData = append(frameData, ownerBytes...)
	frameData = append(frameData, 0x00)
	frameData = append(frameData, payload...)

	// PRIV frame header (ID3v2.3):
	// ID (4 bytes) + Size (4 bytes, big-endian) + Flags (2 bytes)
	var frame bytes.Buffer
	frame.WriteString("PRIV")
	// Size is length of frameData (not including header), big-endian
	binary.Write(&frame, binary.BigEndian, uint32(len(frameData)))
	// Flags (2 bytes)
	frame.Write([]byte{0x00, 0x00})
	frame.Write(frameData)

	// Tag header (10 bytes): 'ID3' + ver(3) + rev(0) + flags(0) + size(4 synchsafe)
	var tag bytes.Buffer
	tag.WriteString("ID3")
	tag.WriteByte(0x03) // version 2.3.0
	tag.WriteByte(0x00) // revision
	tag.WriteByte(0x00) // flags

	// Tag size is size of all frames (no header) encoded as 4 synchsafe bytes
	size := synchsafeEncode(uint32(frame.Len()))
	tag.Write(size[:])
	tag.Write(frame.Bytes())

	return tag.Bytes()
}

// EmbedPayloadInMP3 prepends an ID3v2.3 PRIV tag carrying the payload to the provided MP3 bytes.
// The resulting file remains a valid and playable MP3 in most players.
func (e *audioEncoder) EmbedPayloadInMP3(originalMP3 []byte, owner string, payload []byte) ([]byte, error) {
	if len(originalMP3) == 0 {
		return nil, fmt.Errorf("empty MP3 input")
	}

	tag := buildID3v23PrivTag(owner, payload)

	// If original already starts with ID3, we still safely prepend a new tag.
	// Most players will skip tags and proceed to audio frames.
	var out bytes.Buffer
	out.Grow(len(tag) + len(originalMP3))
	out.Write(tag)
	out.Write(originalMP3)

	log.Printf("[DEBUG] EmbedPayloadInMP3: Embedded %d bytes payload via ID3 PRIV (owner='%s'), output size=%d", len(payload), owner, out.Len())
	return out.Bytes(), nil
}

// ExtractPayloadFromMP3 looks for an ID3v2 tag and reads a PRIV frame for the given owner identifier.
// Returns the payload if found.
func (e *audioEncoder) ExtractPayloadFromMP3(mp3Data []byte, owner string) ([]byte, bool, error) {
	if len(mp3Data) < 10 || string(mp3Data[:3]) != "ID3" {
		return nil, false, nil // No ID3 tag at start; not an error
	}

	version := mp3Data[3]
	if version != 2 && version != 3 && version != 4 {
		// Unknown/unsupported version, stop parsing
		return nil, false, fmt.Errorf("unsupported ID3v2 version: %d", version)
	}

	// Tag size (synchsafe)
	tagSize := uint32(mp3Data[6]&0x7F)<<21 | uint32(mp3Data[7]&0x7F)<<14 | uint32(mp3Data[8]&0x7F)<<7 | uint32(mp3Data[9]&0x7F)
	if int(10+tagSize) > len(mp3Data) {
		// Corrupt tag size
		return nil, false, fmt.Errorf("invalid ID3 tag size")
	}

	// Parse frames in the tag body
	offset := 10
	end := 10 + int(tagSize)
	for offset+10 <= end { // need at least 10 bytes for a frame header
		frameID := string(mp3Data[offset : offset+4])
		// Stop if we hit padding (zeroed frame ID)
		if frameID == "\x00\x00\x00\x00" {
			break
		}

		frameSize := binary.BigEndian.Uint32(mp3Data[offset+4 : offset+8])
		// Flags := mp3Data[offset+8 : offset+10] // not used
		offset += 10

		if int(offset+int(frameSize)) > end || offset+int(frameSize) > len(mp3Data) {
			return nil, false, fmt.Errorf("invalid frame size")
		}

		if frameID == "PRIV" {
			frameData := mp3Data[offset : offset+int(frameSize)]
			// Find owner identifier (null-terminated ISO-8859-1)
			sep := -1
			for i, b := range frameData {
				if b == 0x00 {
					sep = i
					break
				}
			}
			if sep >= 0 {
				gotOwner := string(frameData[:sep])
				if gotOwner == owner {
					payload := frameData[sep+1:]
					log.Printf("[DEBUG] ExtractPayloadFromMP3: Found PRIV payload (owner='%s', size=%d)", owner, len(payload))
					return payload, true, nil
				}
			}
		}

		// Move to next frame (note: frames may be padded to align, but v2.3 doesn't require padding)
		offset += int(frameSize)
	}

	return nil, false, nil
}
