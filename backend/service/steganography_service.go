package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"

	"github.com/Nerggg/Audio-Steganography-LSB/backend/models"
)

// Implementation struct which depends on Crypto and Audio services
type stegoService struct {
	crypto CryptographyService
	audio  AudioService
}

func NewStegoService(crypto CryptographyService, audio AudioService) SteganographyService {
	return &stegoService{crypto: crypto, audio: audio}
}

/*
 Format header (binary, fixed order):
 - 8 bytes magic: "ASTEGv2\000" (8 bytes) - v2 to support multiple methods
 - 1 byte method: 0=LSB, 1=Parity
 - 1 byte nLSB (1..4, only used for LSB method)
 - 1 byte flags: bit0 = UseEncryption, bit1 = UseRandomStart
 - 2 bytes filename length (uint16 big endian)
 - 4 bytes secret payload length (uint32 big endian)  <-- length AFTER encryption (i.e. stored)
 - filename bytes (utf-8) [filename length]
 - secret bytes ...
*/

// helper constants
var (
	magicBytes = []byte("ASTEGv2\x00")
)

// method constants
const (
	methodLSB    = 0
	methodParity = 1
)

// ------------------ Helpers ------------------

func checkSync(b byte) bool {
	// sync word first byte must be 0xFF, second byte top 3 bits 111 (0xE0)
	return b == 0xFF
}

func isFrameSyncAt(data []byte, i int) bool {
	if i+1 >= len(data) {
		return false
	}
	return data[i] == 0xFF && (data[i+1]&0xE0) == 0xE0
}

// parseID3v2Size returns offset after tag (i.e., first byte of audio or next items).
// If no ID3 header, returns 0.
func parseID3v2Size(data []byte) int {
	if len(data) < 10 {
		return 0
	}
	if string(data[0:3]) != "ID3" {
		return 0
	}
	// synchsafe size in bytes 6..9 (4 bytes)
	if len(data) < 10 {
		return 0
	}
	size := int((uint32(data[6])&0x7F)<<21 |
		(uint32(data[7])&0x7F)<<14 |
		(uint32(data[8])&0x7F)<<7 |
		(uint32(data[9]) & 0x7F))
	return 10 + size
}

// parseMP3FrameSize parses the MP3 frame header at pos and returns the frame size in bytes.
// Returns 0 if invalid header or insufficient data.
func parseMP3FrameSize(data []byte, pos int) int {
	if len(data) < pos+4 {
		return 0
	}
	if data[pos] != 0xFF || (data[pos+1]&0xE0) != 0xE0 {
		return 0
	}

	versionBits := (data[pos+1] >> 3) & 0x03
	if versionBits == 0x01 { // reserved
		return 0
	}
	layerBits := (data[pos+1] >> 1) & 0x03
	if layerBits == 0x00 { // reserved
		return 0
	}
	bitrateIdx := data[pos+2] >> 4
	if bitrateIdx == 0x0F || bitrateIdx == 0x00 { // bad or free (we treat free as invalid for simplicity)
		return 0
	}
	sampleRateIdx := (data[pos+2] >> 2) & 0x03
	if sampleRateIdx == 0x03 { // reserved
		return 0
	}
	padding := (data[pos+2] >> 1) & 0x01

	// Map version: 3=MPEG1, 2=MPEG2, 0=MPEG2.5
	var vid int
	if versionBits == 0x03 {
		vid = 0 // MPEG1
	} else if versionBits == 0x02 {
		vid = 1 // MPEG2
	} else if versionBits == 0x00 {
		vid = 2 // MPEG2.5
	} else {
		return 0
	}

	// Map layer: 3=Layer1, 2=Layer2, 1=Layer3
	var lid int
	if layerBits == 0x03 {
		lid = 0 // Layer1
	} else if layerBits == 0x02 {
		lid = 1 // Layer2
	} else if layerBits == 0x01 {
		lid = 2 // Layer3
	} else {
		return 0
	}

	// Bitrate table (kbps): [vid][lid][bitrateIdx]
	bitrateTable := [3][3][15]int{
		{ // MPEG1 (vid=0)
			{32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, 0}, // Layer1
			{32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384, 0},    // Layer2
			{32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0},     // Layer3
		},
		{ // MPEG2 (vid=1)
			{32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 0}, // Layer1
			{8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},      // Layer2
			{8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},      // Layer3
		},
		{ // MPEG2.5 (vid=2, same as MPEG2)
			{32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 0}, // Layer1
			{8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},      // Layer2
			{8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},      // Layer3
		},
	}
	bitrate := bitrateTable[vid][lid][bitrateIdx-1] // idx starts from 1
	if bitrate == 0 {
		return 0
	}

	// Sample rate table: [versionBits][sampleRateIdx]
	sampleRateTable := [4][3]int{
		{11025, 12000, 8000},  // MPEG2.5 (0)
		{0, 0, 0},             // reserved (1)
		{22050, 24000, 16000}, // MPEG2 (2)
		{44100, 48000, 32000}, // MPEG1 (3)
	}
	sr := sampleRateTable[versionBits][sampleRateIdx]
	if sr == 0 {
		return 0
	}

	// Calculate frame size
	var frameSize int
	if layerBits == 0x03 { // Layer1
		frameSize = ((12 * bitrate * 1000 / sr) + int(padding)) * 4
	} else { // Layer2/3
		frameSize = (144 * bitrate * 1000 / sr) + int(padding)
	}

	if frameSize < 4 || pos+frameSize > len(data) {
		return 0
	}
	return frameSize
}

// collectPayloadIndices returns a slice of indices of bytes that are considered "payload bytes"
// i.e., bytes between frame header and end of frame. This uses proper frame size calculation for robustness.
func collectPayloadIndices(data []byte) []int {
	var indices []int
	// start after ID3 tag
	start := parseID3v2Size(data)
	i := start
	for i < len(data)-4 { // need at least header size
		if !isFrameSyncAt(data, i) {
			i++
			continue
		}
		size := parseMP3FrameSize(data, i)
		if size <= 4 {
			i++
			continue
		}
		// add payload bytes: from i+4 to i+size-1
		for j := i + 4; j < i+size && j < len(data); j++ {
			indices = append(indices, j)
		}
		// jump to next frame
		i += size
	}
	return indices
}

// deterministicStartIndex chooses deterministic start bit index from key and capacityBits
func deterministicStartIndex(key string, capacityBits int) int {
	if capacityBits == 0 {
		return 0
	}
	h := sha256.Sum256([]byte(key))
	seed := int64(binary.BigEndian.Uint64(h[:8]))
	r := rand.New(rand.NewSource(seed))
	return r.Intn(capacityBits)
}

// ------------------ Interface Implementations ------------------

// CalculateCapacity calculates available embedding capacity for both LSB and Parity methods (in bytes).
func (s *stegoService) CalculateCapacity(audioData []byte) (*models.CapacityResult, error) {
	if len(audioData) == 0 {
		return nil, models.ErrInvalidMP3
	}
	indices := collectPayloadIndices(audioData)
	if len(indices) == 0 {
		return nil, models.ErrInvalidMP3
	}
	totalPayloadBytes := len(indices)
	// capacity for n LSB = floor(totalPayloadBytes * n / 8) bytes
	// capacity for parity = floor(totalPayloadBytes / 8) bytes (1 bit per byte)
	res := &models.CapacityResult{
		OneLSB:   (totalPayloadBytes * 1) / 8,
		TwoLSB:   (totalPayloadBytes * 2) / 8,
		ThreeLSB: (totalPayloadBytes * 3) / 8,
		FourLSB:  (totalPayloadBytes * 4) / 8,
		Parity:   totalPayloadBytes / 8, // 1 bit per byte
	}
	return res, nil
}

// EmbedMessage embeds secretData (and metadata) into req.CoverAudio using LSB or Parity method.
func (s *stegoService) EmbedMessage(req *models.EmbedRequest, secretData []byte, metadata []byte) ([]byte, float64, error) {
	// validate method
	if !req.Method.IsValid() {
		return nil, 0, models.ErrInvalidMethod
	}

	// validate LSB count for LSB method
	if req.Method == models.MethodLSB && (req.NLsb < 1 || req.NLsb > 4) {
		return nil, 0, models.ErrInvalidLSB
	}

	cover := make([]byte, len(req.CoverAudio))
	copy(cover, req.CoverAudio)

	// Optional encryption
	secretToStore := make([]byte, len(secretData))
	copy(secretToStore, secretData)
	if req.UseEncryption {
		if req.StegoKey == "" {
			return nil, 0, models.ErrInvalidStegoKey
		}
		// Add a simple checksum (first 4 bytes of data hash) before encryption for integrity verification
		checksum := calculateChecksum(secretData)
		dataWithChecksum := append(checksum[:], secretData...)
		secretToStore = s.crypto.VigenereCipher(dataWithChecksum, req.StegoKey, true)
	}

	// Build header+payload:
	// [magic(8)][method(1)][nLSB(1)][flags(1)][filenameLen(2)][secretLen(4)][filename][metadataLen(2)][metadata][secret bytes]
	buf := bytes.Buffer{}
	buf.Write(magicBytes)

	// Write method type
	if req.Method == models.MethodLSB {
		buf.WriteByte(methodLSB)
	} else {
		buf.WriteByte(methodParity)
	}

	// Write nLSB (only meaningful for LSB method, but always present for format consistency)
	nLsb := req.NLsb
	if req.Method == models.MethodParity {
		nLsb = 1 // Parity method uses 1 bit per byte
	}
	buf.WriteByte(byte(nLsb))

	flags := byte(0)
	if req.UseEncryption {
		flags |= 1 << 0
	}
	if req.UseRandomStart {
		flags |= 1 << 1
	}
	buf.WriteByte(flags)

	// filename
	filename := req.SecretFileName
	if filename == "" {
		filename = "secret.bin"
	}
	if len(filename) > 0xFFFF {
		return nil, 0, models.ErrFileTooLarge
	}
	binary.Write(&buf, binary.BigEndian, uint16(len(filename)))
	binary.Write(&buf, binary.BigEndian, uint32(len(secretToStore)))
	buf.WriteString(filename)

	// metadata (arbitrary bytes) - allow zero length
	if metadata == nil {
		metadata = []byte{}
	}
	if len(metadata) > 0xFFFF {
		return nil, 0, models.ErrFileTooLarge
	}
	binary.Write(&buf, binary.BigEndian, uint16(len(metadata)))
	if len(metadata) > 0 {
		buf.Write(metadata)
	}

	// secret bytes
	buf.Write(secretToStore)
	toEmbedBytes := buf.Bytes()
	toEmbedBits := bytesToBits(toEmbedBytes)

	// collect payload positions (byte indices in cover)
	payloadIdxs := collectPayloadIndices(cover)
	if len(payloadIdxs) == 0 {
		return nil, 0, models.ErrInvalidMP3
	}

	// Calculate capacity based on method
	var totalCapacityBits int
	if req.Method == models.MethodLSB {
		totalCapacityBits = len(payloadIdxs) * req.NLsb
	} else { // Parity method
		totalCapacityBits = len(payloadIdxs) // 1 bit per byte
	}

	if len(toEmbedBits) > totalCapacityBits {
		return nil, 0, models.ErrInsufficientCapacity
	}

	// determine start bit
	startBit := 0
	if req.UseRandomStart {
		if req.StegoKey == "" {
			return nil, 0, models.ErrInvalidStegoKey
		}
		startBit = deterministicStartIndex(req.StegoKey, totalCapacityBits)
	}

	// Embed bits using the selected method
	if req.Method == models.MethodLSB {
		// LSB embedding - embed bits sequentially into LSBs of payload bytes
		bitPos := startBit
		for i := 0; i < len(toEmbedBits); {
			if bitPos >= totalCapacityBits {
				// wrap around to beginning (deterministic)
				bitPos = 0
			}
			// find which payload byte and which bit-in-byte slot
			payloadByteIndex := bitPos / req.NLsb         // which payload byte (index in payloadIdxs)
			slotIndex := bitPos % req.NLsb                // which LSB slot in that byte (0..n-1)
			coverBytePos := payloadIdxs[payloadByteIndex] // actual byte index in cover
			// set or clear that specific LSB slot according to next bit
			bit := toEmbedBits[i]
			if bit == 1 {
				cover[coverBytePos] |= (1 << uint(slotIndex))
			} else {
				cover[coverBytePos] &^= (1 << uint(slotIndex))
			}
			i++
			bitPos++
		}
	} else { // Parity method
		// Parity embedding - embed bits by adjusting parity of payload bytes
		bitPos := startBit
		for i := 0; i < len(toEmbedBits); {
			if bitPos >= totalCapacityBits {
				// wrap around to beginning (deterministic)
				bitPos = 0
			}
			coverBytePos := payloadIdxs[bitPos] // direct mapping: bit index to payload byte
			bit := toEmbedBits[i]
			cover[coverBytePos] = embedParityBit(cover[coverBytePos], bit)
			i++
			bitPos++
		}
	}

	// calculate PSNR using audio service
	psnr := s.audio.CalculatePSNR(req.CoverAudio, cover)

	return cover, psnr, nil
}

// ExtractMessage extracts embedded data from audioData using method and parameters stored in header.
// If req.StegoKey is required to decrypt, it will be used.
func (s *stegoService) ExtractMessage(req *models.ExtractRequest, audioData []byte) ([]byte, string, error) {
	if len(audioData) == 0 {
		return nil, "", models.ErrInvalidMP3
	}
	cover := audioData
	payloadIdxs := collectPayloadIndices(cover)
	if len(payloadIdxs) == 0 {
		return nil, "", models.ErrInvalidMP3
	}

	// Try both methods if not specified, or use specified method
	methodsToTry := []int{methodLSB, methodParity}
	if req.Method.IsValid() {
		if req.Method == models.MethodLSB {
			methodsToTry = []int{methodLSB}
		} else {
			methodsToTry = []int{methodParity}
		}
	}

	for _, method := range methodsToTry {
		var result []byte
		var filename string
		var err error

		if method == methodLSB {
			result, filename, err = s.extractLSBMethod(req, cover, payloadIdxs)
		} else {
			result, filename, err = s.extractParityMethod(req, cover, payloadIdxs)
		}

		if err == nil && result != nil {
			return result, filename, nil
		}
	}

	return nil, "", models.ErrExtractionFailed
}

// extractLSBMethod extracts data using LSB method (tries different n values)
func (s *stegoService) extractLSBMethod(req *models.ExtractRequest, cover []byte, payloadIdxs []int) ([]byte, string, error) {
	// Try n = 1..4 LSBs since we don't know which was used
	for n := 1; n <= 4; n++ {
		totalBits := len(payloadIdxs) * n
		bits := make([]uint8, 0, totalBits)
		// get linear bit sequence in LSB order (slot 0..n-1 per payload byte)
		for _, idx := range payloadIdxs {
			for slot := 0; slot < n; slot++ {
				bits = append(bits, (cover[idx]>>uint(slot))&1)
			}
		}

		result, filename, err := s.tryExtractFromBits(req, bits, totalBits, methodLSB, n)
		if err == nil {
			return result, filename, nil
		}
	}
	return nil, "", models.ErrExtractionFailed
}

// extractParityMethod extracts data using Parity method
func (s *stegoService) extractParityMethod(req *models.ExtractRequest, cover []byte, payloadIdxs []int) ([]byte, string, error) {
	totalBits := len(payloadIdxs) // 1 bit per byte for parity
	bits := make([]uint8, 0, totalBits)

	// Extract parity bits from each payload byte
	for _, idx := range payloadIdxs {
		bit := extractParityBit(cover[idx])
		bits = append(bits, bit)
	}

	return s.tryExtractFromBits(req, bits, totalBits, methodParity, 1)
}

// tryExtractFromBits attempts to extract data from a bit stream
func (s *stegoService) tryExtractFromBits(req *models.ExtractRequest, bits []uint8, totalBits int, expectedMethod int, expectedN int) ([]byte, string, error) {
	// Try possible random start positions
	tryStarts := []int{0}
	if req.StegoKey != "" {
		start := deterministicStartIndex(req.StegoKey, totalBits)
		tryStarts = append(tryStarts, start)
	}

	for _, start := range tryStarts {
		// rotate bits so that start becomes 0
		rot := make([]uint8, len(bits))
		for i := 0; i < len(bits); i++ {
			rot[i] = bits[(start+i)%len(bits)]
		}
		// convert first bytes enough to check magic and header sizes
		raw := bitsToBytes(rot)
		// need at least header length: magic(8)+method(1)+nLSB(1)+flags(1)+filenameLen(2)+secretLen(4) = 17 bytes
		if len(raw) < 17 {
			continue
		}
		if !bytes.Equal(raw[0:8], magicBytes) {
			continue
		}

		embeddedMethod := int(raw[8])
		embeddedN := int(raw[9])
		flags := raw[10]

		// verify method and n match expected values
		if embeddedMethod != expectedMethod || embeddedN != expectedN {
			continue
		}

		// read filename len and secret len
		filenameLen := int(binary.BigEndian.Uint16(raw[11:13]))
		secretLen := int(binary.BigEndian.Uint32(raw[13:17]))
		// check lengths sanity
		headerTotal := 8 + 1 + 1 + 1 + 2 + 4 + filenameLen + 2 // magic+method+nLSB+flags+filenameLen+secretLen+filename+metadataLen
		// need to ensure we extracted enough bytes to read metadataLen too
		if len(raw) < headerTotal {
			// insufficient raw, continue
			continue
		}
		filenameStart := 17
		if filenameLen > 0 {
			if len(raw) < filenameStart+filenameLen+2 {
				continue
			}
		}
		filename := string(raw[17 : 17+filenameLen])
		metaLenOff := 17 + filenameLen
		metadataLen := int(binary.BigEndian.Uint16(raw[metaLenOff : metaLenOff+2]))
		metaStart := metaLenOff + 2
		if len(raw) < metaStart+metadataLen+secretLen {
			// maybe we didn't extract whole payload yet; but if not enough capacity, skip
			// However we only need to return the secret if lengths valid
			// continue to next attempt
			continue
		}
		secretStart := metaStart + metadataLen
		secretBytes := raw[secretStart : secretStart+secretLen]
		// If encryption flag set, and key provided, decrypt
		encFlag := (flags & (1 << 0)) != 0
		if encFlag {
			if req.StegoKey == "" {
				return nil, "", models.ErrInvalidStegoKey
			}
			decrypted := s.crypto.VigenereCipher(secretBytes, req.StegoKey, false)

			// Validate checksum (first 4 bytes)
			if len(decrypted) < 4 {
				return nil, "", models.ErrInvalidStegoKey
			}

			actualData := decrypted[4:]
			expectedChecksum := calculateChecksum(actualData)

			// Compare checksums
			for i := 0; i < 4; i++ {
				if decrypted[i] != expectedChecksum[i] {
					return nil, "", models.ErrInvalidStegoKey
				}
			}

			secretBytes = actualData
		}
		// success
		return secretBytes, filename, nil
	}

	return nil, "", models.ErrExtractionFailed
}
