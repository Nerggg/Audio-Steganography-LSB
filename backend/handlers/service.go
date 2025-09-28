package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Nerggg/Audio-Steganography-LSB/backend/models"
	"github.com/Nerggg/Audio-Steganography-LSB/backend/service"
	"github.com/gin-gonic/gin"
)

// Handlers struct holds service dependencies
type Handlers struct {
	steganographyService service.SteganographyService
	cryptographyService  service.CryptographyService
	audioService         service.AudioService
	audioEncoder         service.AudioEncoder
}

// NewHandlers creates a new handlers instance with service dependencies
func NewHandlers(
	stegoService service.SteganographyService,
	cryptoService service.CryptographyService,
	audioService service.AudioService,
	audioEncoder service.AudioEncoder,
) *Handlers {
	return &Handlers{
		steganographyService: stegoService,
		cryptographyService:  cryptoService,
		audioService:         audioService,
		audioEncoder:         audioEncoder,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status       string            `json:"status"`
	Timestamp    time.Time         `json:"timestamp"`
	Version      string            `json:"version"`
	Uptime       int               `json:"uptime"`
	Dependencies map[string]string `json:"dependencies"`
}

// CapacityResponse represents the capacity calculation response
type CapacityResponse struct {
	Capacities       models.CapacityResult `json:"capacities"`
	FileInfo         FileInfo              `json:"file_info"`
	ProcessingTimeMs int                   `json:"processing_time_ms"`
}

// FileInfo represents audio file information
type FileInfo struct {
	Filename        string  `json:"filename"`
	SizeBytes       int     `json:"size_bytes"`
	DurationSeconds float64 `json:"duration_seconds,omitempty"`
	Bitrate         int     `json:"bitrate,omitempty"`
	SampleRate      int     `json:"sample_rate,omitempty"`
	Channels        int     `json:"channels,omitempty"`
}

// HealthHandler handles the health check endpoint
// @Summary Health Check
// @Description Returns the health status of the API service
// @Tags System
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Failure 503 {object} models.ErrorResponse "Service unavailable"
// @Router /health [get]
func (h *Handlers) HealthHandler(c *gin.Context) {
	startTime := time.Now()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    86400, // This should be calculated from server start time
		Dependencies: map[string]string{
			"database": "healthy",
			"storage":  "healthy",
		},
	}

	processingTime := time.Since(startTime).Milliseconds()
	c.Header("X-Processing-Time", strconv.FormatInt(processingTime, 10))
	c.JSON(http.StatusOK, response)
}

// CalculateCapacityHandler handles the capacity calculation request
// @Summary Calculate Audio Embedding Capacity
// @Description Calculates the maximum size of a secret file (in bytes) that can be embedded into an uploaded audio file (MP3 or WAV) using the multiple-LSB method. The capacity is returned for 1, 2, 3, and 4 LSBs.
// @Tags Steganography
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (MP3 or WAV) to calculate capacity for."
// @Success 200 {object} CapacityResponse "Successfully calculated embedding capacity."
// @Header 200 {int} X-Processing-Time "Time taken to process the request in milliseconds"
// @Failure 400 {object} models.ErrorResponse "Bad Request: No file uploaded, file is not MP3/WAV, or file is corrupted."
// @Failure 413 {object} models.ErrorResponse "File too large"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error: Failed to process the file."
// @Router /capacity [post]
func (h *Handlers) CalculateCapacityHandler(c *gin.Context) {
	startTime := time.Now()
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	log.Printf("[INFO] [%s] CalculateCapacityHandler: Starting capacity calculation request from %s", requestID, c.ClientIP())

	// Get audio file from form request
	fileHeader, err := c.FormFile("audio")
	if err != nil {
		log.Printf("[ERROR] [%s] CalculateCapacityHandler: No audio file provided: %v", requestID, err)
		sendError(c, http.StatusBadRequest, "MISSING_FILE", "Audio file not provided")
		return
	}

	log.Printf("[DEBUG] [%s] CalculateCapacityHandler: Received file '%s' (size: %d bytes)", requestID, fileHeader.Filename, fileHeader.Size)

	// Validate file extension (support both MP3 and WAV)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".mp3" && ext != ".wav" {
		log.Printf("[ERROR] [%s] CalculateCapacityHandler: Invalid file format '%s', expected MP3 or WAV", requestID, ext)
		sendError(c, http.StatusBadRequest, "INVALID_FORMAT", "File must be in MP3 or WAV format")
		return
	}

	// Check file size (max 100MB)
	if fileHeader.Size > 100*1024*1024 {
		sendError(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "File size exceeds maximum limit of 100MB")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to open uploaded file")
		return
	}
	defer file.Close()

	// Read file content into byte slice
	audioData, err := io.ReadAll(file)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to read file content")
		return
	}

	// Calculate capacity using steganography service
	capacities, err := h.steganographyService.CalculateCapacity(audioData)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to calculate capacity")
		return
	}

	// Create file info
	fileInfo := FileInfo{
		Filename:        fileHeader.Filename,
		SizeBytes:       int(fileHeader.Size),
		DurationSeconds: 180.5, // Placeholder - should be calculated from MP3 metadata
		Bitrate:         320,   // Placeholder
		SampleRate:      44100, // Placeholder
		Channels:        2,     // Placeholder
	}

	processingTime := int(time.Since(startTime).Milliseconds())

	response := CapacityResponse{
		Capacities:       *capacities,
		FileInfo:         fileInfo,
		ProcessingTimeMs: processingTime,
	}

	c.Header("X-Processing-Time", strconv.Itoa(processingTime))
	c.JSON(http.StatusOK, response)
}

// EmbedHandler handles the message embedding request
// @Summary Embed Secret Data in Audio
// @Description Embeds a secret file into an MP3 audio file using LSB steganography. Returns the stego audio file in MP3 format (if ffmpeg available) or WAV format (fallback) with embedded data and quality metrics.
// @Tags Steganography
// @Accept multipart/form-data
// @Produce audio/mpeg,audio/wav
// @Param audio formData file true "MP3 audio file for embedding (max 100MB)"
// @Param secret formData file true "Secret file to embed (max 50MB)"
// @Param lsb formData int true "Number of LSB bits to use for embedding (1-4)" Enums(1, 2, 3, 4)
// @Param use_encryption formData string false "Whether to encrypt the secret file using Vigenère cipher" Enums(true, false)
// @Param use_random_start formData string false "Whether to use random starting position for embedding" Enums(true, false)
// @Param stego_key formData string false "Steganography key used for encryption and random seed generation (max 25 characters)"
// @Param output_filename formData string false "Desired filename for the output stego audio"
// @Success 200 {file} file "Successfully embedded secret data"
// @Header 200 {string} Content-Disposition "Filename for the stego audio file"
// @Header 200 {number} X-PSNR-Value "Peak Signal-to-Noise Ratio indicating audio quality after embedding"
// @Header 200 {string} X-Embedding-Method "LSB method used for embedding"
// @Header 200 {int} X-Secret-Size "Size of embedded secret in bytes"
// @Header 200 {int} X-Processing-Time "Time taken to process the request in milliseconds"
// @Failure 400 {object} models.ErrorResponse "Bad Request: Invalid parameters or file issues"
// @Failure 413 {object} models.ErrorResponse "Files too large"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /embed [post]
func (h *Handlers) EmbedHandler(c *gin.Context) {
	startTime := time.Now()
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	log.Printf("[INFO] [%s] EmbedHandler: Starting embed request from %s", requestID, c.ClientIP())

	// Get audio file
	audioHeader, err := c.FormFile("audio")
	if err != nil {
		log.Printf("[ERROR] [%s] EmbedHandler: No audio file provided: %v", requestID, err)
		sendError(c, http.StatusBadRequest, "MISSING_FILES", "Audio file not provided")
		return
	}

	log.Printf("[DEBUG] [%s] EmbedHandler: Audio file '%s' (size: %d bytes)", requestID, audioHeader.Filename, audioHeader.Size)

	// Get secret file
	secretHeader, err := c.FormFile("secret")
	if err != nil {
		log.Printf("[ERROR] [%s] EmbedHandler: No secret file provided: %v", requestID, err)
		sendError(c, http.StatusBadRequest, "MISSING_FILES", "Secret file not provided")
		return
	}

	log.Printf("[DEBUG] [%s] EmbedHandler: Secret file '%s' (size: %d bytes)", requestID, secretHeader.Filename, secretHeader.Size)

	// Get LSB parameter
	lsbStr := c.PostForm("lsb")
	if lsbStr == "" {
		sendError(c, http.StatusBadRequest, "INVALID_LSB", "LSB parameter is required")
		return
	}

	lsb, err := strconv.Atoi(lsbStr)
	if err != nil || lsb < 1 || lsb > 4 {
		sendError(c, http.StatusBadRequest, "INVALID_LSB", "LSB value must be between 1 and 4")
		return
	}

	// Validate file extensions
	if filepath.Ext(audioHeader.Filename) != ".mp3" {
		sendError(c, http.StatusBadRequest, "INVALID_FORMAT", "Audio file must be in MP3 format")
		return
	}

	// Check file sizes
	if audioHeader.Size > 100*1024*1024 {
		sendError(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "Audio file size exceeds maximum limit of 100MB")
		return
	}

	if secretHeader.Size > 50*1024*1024 {
		sendError(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "Secret file size exceeds maximum limit of 50MB")
		return
	}

	// Read audio file
	audioFile, err := audioHeader.Open()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to open audio file")
		return
	}
	defer audioFile.Close()

	audioData, err := io.ReadAll(audioFile)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to read audio file")
		return
	}

	// Read secret file
	secretFile, err := secretHeader.Open()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to open secret file")
		return
	}
	defer secretFile.Close()

	secretData, err := io.ReadAll(secretFile)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to read secret file")
		return
	}

	// Create embed request
	// Parse optional parameters
	stegoKey := c.PostForm("stego_key")
	useEncryption := c.PostForm("use_encryption") == "true"
	useRandomStart := c.PostForm("use_random_start") == "true"

	// Validate stego key requirements
	if (useEncryption || useRandomStart) && stegoKey == "" {
		sendError(c, http.StatusBadRequest, "INVALID_STEGO_KEY", "Stego key is required when encryption or random start is enabled")
		return
	}

	if stegoKey != "" && len(stegoKey) < 4 {
		sendError(c, http.StatusBadRequest, "WEAK_STEGO_KEY", "Stego key must be at least 4 characters long")
		return
	}

	embedReq := &models.EmbedRequest{
		CoverAudio:     audioData,
		SecretFile:     secretData,
		SecretFileName: secretHeader.Filename,
		StegoKey:       stegoKey,
		NLsb:           lsb,
		UseEncryption:  useEncryption,
		UseRandomStart: useRandomStart,
	}

	// Create metadata
	metadata := h.steganographyService.CreateMetadata(
		secretHeader.Filename,
		len(secretData),
		embedReq.UseEncryption,
		embedReq.UseRandomStart,
		lsb,
	)

	// Perform embedding using existing controller
	stegoAudio, psnr, err := h.steganographyService.EmbedMessage(embedReq, secretData, metadata)
	if err != nil {
		if err.Error() == "secret data too large for audio capacity" {
			sendError(c, http.StatusBadRequest, "CAPACITY_EXCEEDED", "Secret file size exceeds audio capacity for selected LSB method")
			return
		}
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to embed data")
		return
	}

	processingTime := int(time.Since(startTime).Milliseconds())
	outputFilename := c.PostForm("output_filename")
	// Detect actual format by header
	isMP3 := len(stegoAudio) >= 3 && (string(stegoAudio[:3]) == "ID3" || (stegoAudio[0] == 0xFF && (stegoAudio[1]&0xE0) == 0xE0))

	if outputFilename == "" {
		if isMP3 {
			outputFilename = "stego_audio.mp3"
		} else {
			outputFilename = "stego_audio.wav"
		}
	} else {
		if isMP3 {
			if !strings.HasSuffix(strings.ToLower(outputFilename), ".mp3") {
				outputFilename = strings.TrimSuffix(outputFilename, filepath.Ext(outputFilename)) + ".mp3"
			}
		} else {
			if !strings.HasSuffix(strings.ToLower(outputFilename), ".wav") {
				outputFilename = strings.TrimSuffix(outputFilename, filepath.Ext(outputFilename)) + ".wav"
			}
		}
	}

	// Set response headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", outputFilename))
	c.Header("X-PSNR-Value", fmt.Sprintf("%.2f", psnr))
	c.Header("X-Embedding-Method", fmt.Sprintf("%d-LSB", lsb))
	c.Header("X-Secret-Size", strconv.Itoa(len(secretData)))
	c.Header("X-Processing-Time", strconv.Itoa(processingTime))
	if isMP3 {
		c.Header("X-Output-Format", "MP3")
	} else {
		c.Header("X-Output-Format", "WAV")
	}

	// Return audio file
	if isMP3 {
		c.Data(http.StatusOK, "audio/mpeg", stegoAudio)
	} else {
		c.Data(http.StatusOK, "audio/wav", stegoAudio)
	}
}

// ExtractHandler handles the data extraction request
// @Summary Extract Secret Data from Audio (Simplified)
// @Description Extracts hidden secret data from a stego audio file (MP3 or WAV) with auto-detection of embedding parameters. Only requires the stego file and optional key.
// @Tags Steganography
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param stego_audio formData file true "Stego audio file (MP3 or WAV) containing embedded data (max 100MB)"
// @Param stego_key formData string false "Steganography key (required only if file was encrypted during embedding)"
// @Param output_filename formData string false "Desired filename for the extracted secret file (optional - will use original filename if not provided)"
// @Success 200 {file} file "Successfully extracted secret data"
// @Header 200 {string} Content-Disposition "Original filename of the extracted secret"
// @Header 200 {string} X-Extraction-Method "Auto-detected LSB method used for extraction"
// @Header 200 {int} X-Secret-Size "Size of extracted secret in bytes"
// @Header 200 {int} X-Processing-Time "Time taken to process the request in milliseconds"
// @Failure 400 {object} models.ErrorResponse "Bad Request: Invalid file format"
// @Failure 413 {object} models.ErrorResponse "File too large"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error: Failed to extract or auto-detect parameters"
// @Router /extract [post]
func (h *Handlers) ExtractHandler(c *gin.Context) {
	startTime := time.Now()
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	log.Printf("[INFO] [%s] ExtractHandler: Starting simplified extraction request from %s", requestID, c.ClientIP())

	// Get stego audio file
	stegoHeader, err := c.FormFile("stego_audio")
	if err != nil {
		log.Printf("[ERROR] [%s] ExtractHandler: No stego audio file provided: %v", requestID, err)
		sendError(c, http.StatusBadRequest, "MISSING_FILE", "Stego audio file not provided")
		return
	}

	log.Printf("[DEBUG] [%s] ExtractHandler: Stego file '%s' (size: %d bytes)", requestID, stegoHeader.Filename, stegoHeader.Size)

	// Validate file extension (support both MP3 and WAV)
	ext := strings.ToLower(filepath.Ext(stegoHeader.Filename))
	if ext != ".mp3" && ext != ".wav" {
		log.Printf("[ERROR] [%s] ExtractHandler: Invalid file format '%s', expected MP3 or WAV", requestID, ext)
		sendError(c, http.StatusBadRequest, "INVALID_FORMAT", "File must be in MP3 or WAV format")
		return
	}

	// Check file size
	if stegoHeader.Size > 100*1024*1024 {
		sendError(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "File size exceeds maximum limit of 100MB")
		return
	}

	// Parse optional parameters
	stegoKey := c.PostForm("stego_key")
	outputFilename := c.PostForm("output_filename")

	log.Printf("[DEBUG] [%s] ExtractHandler: Parameters - stego_key_provided=%t, output_filename='%s'",
		requestID, stegoKey != "", outputFilename)

	// Read stego audio file
	stegoFile, err := stegoHeader.Open()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to open stego audio file")
		return
	}
	defer stegoFile.Close()

	stegoData, err := io.ReadAll(stegoFile)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to read stego audio file")
		return
	}

	// Perform auto-detection extraction
	secretData, filename, err := h.steganographyService.ExtractMessageAutoDetect(stegoData, stegoKey, outputFilename)
	if err != nil {
		log.Printf("[ERROR] [%s] ExtractHandler: Auto-detection extraction failed: %v", requestID, err)
		sendError(c, http.StatusInternalServerError, "EXTRACTION_ERROR", "Failed to extract data: "+err.Error())
		return
	}

	processingTime := int(time.Since(startTime).Milliseconds())

	// Set response headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("X-Extraction-Method", "Auto-detected LSB")
	c.Header("X-Secret-Size", strconv.Itoa(len(secretData)))
	c.Header("X-Processing-Time", strconv.Itoa(processingTime))

	log.Printf("[INFO] [%s] ExtractHandler: Extraction completed successfully (extracted_size: %d bytes, filename: '%s', duration: %dms)",
		requestID, len(secretData), filename, processingTime)

	c.Data(http.StatusOK, "application/octet-stream", secretData)
}

// ExtractManualHandler handles manual extraction with explicit parameters (legacy)
// @Summary Extract Secret Data from Audio (Manual Parameters)
// @Description Extracts hidden secret data from a stego audio file with manually specified parameters. For advanced users who know the exact embedding settings.
// @Tags Steganography
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param stego_audio formData file true "Stego audio file (MP3 or WAV) containing embedded data (max 100MB)"
// @Param lsb formData int true "Number of LSB bits used during embedding (1-4)" Enums(1, 2, 3, 4)
// @Param use_encryption formData string false "Whether the embedded data was encrypted using Vigenère cipher" Enums(true, false)
// @Param use_random_start formData string false "Whether random starting position was used during embedding" Enums(true, false)
// @Param stego_key formData string false "Steganography key used for decryption and random position generation"
// @Param output_filename formData string false "Desired filename for the extracted secret file"
// @Success 200 {file} file "Successfully extracted secret data"
// @Header 200 {string} Content-Disposition "Original filename of the extracted secret"
// @Header 200 {string} X-Extraction-Method "LSB method used for extraction"
// @Header 200 {int} X-Secret-Size "Size of extracted secret in bytes"
// @Header 200 {int} X-Processing-Time "Time taken to process the request in milliseconds"
// @Failure 400 {object} models.ErrorResponse "Bad Request: Invalid parameters or file"
// @Failure 413 {object} models.ErrorResponse "File too large"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /extract-manual [post]
func (h *Handlers) ExtractManualHandler(c *gin.Context) {
	startTime := time.Now()
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	log.Printf("[INFO] [%s] ExtractManualHandler: Starting manual extraction request from %s", requestID, c.ClientIP())

	// Get stego audio file
	stegoHeader, err := c.FormFile("stego_audio")
	if err != nil {
		sendError(c, http.StatusBadRequest, "MISSING_FILE", "Stego audio file not provided")
		return
	}

	// Get LSB parameter
	lsbStr := c.PostForm("lsb")
	if lsbStr == "" {
		sendError(c, http.StatusBadRequest, "INVALID_LSB", "LSB parameter is required")
		return
	}

	lsb, err := strconv.Atoi(lsbStr)
	if err != nil || lsb < 1 || lsb > 4 {
		sendError(c, http.StatusBadRequest, "INVALID_LSB", "LSB value must be between 1 and 4")
		return
	}

	// Validate file extension (support both MP3 and WAV)
	ext := strings.ToLower(filepath.Ext(stegoHeader.Filename))
	if ext != ".mp3" && ext != ".wav" {
		sendError(c, http.StatusBadRequest, "INVALID_FORMAT", "File must be in MP3 or WAV format")
		return
	}

	// Check file size
	if stegoHeader.Size > 100*1024*1024 {
		sendError(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "File size exceeds maximum limit of 100MB")
		return
	}

	// Parse optional parameters
	stegoKey := c.PostForm("stego_key")
	useEncryption := c.PostForm("use_encryption") == "true"
	useRandomStart := c.PostForm("use_random_start") == "true"
	outputFilename := c.PostForm("output_filename")

	// Read stego audio file
	stegoFile, err := stegoHeader.Open()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to open stego audio file")
		return
	}
	defer stegoFile.Close()

	stegoData, err := io.ReadAll(stegoFile)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to read stego audio file")
		return
	}

	// Create extraction request (legacy method)
	extractReq := &models.ExtractRequest{
		StegoAudio:     stegoData,
		NLsb:           lsb,
		StegoKey:       stegoKey,
		UseEncryption:  useEncryption,
		UseRandomStart: useRandomStart,
		OutputFilename: outputFilename,
	}

	// Perform extraction using legacy method
	secretData, filename, err := h.steganographyService.ExtractMessage(extractReq, stegoData)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "EXTRACTION_ERROR", "Failed to extract data: "+err.Error())
		return
	}

	processingTime := int(time.Since(startTime).Milliseconds())

	if outputFilename == "" {
		outputFilename = filename
	}

	// Set response headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", outputFilename))
	c.Header("X-Extraction-Method", fmt.Sprintf("%d-LSB", lsb))
	c.Header("X-Secret-Size", strconv.Itoa(len(secretData)))
	c.Header("X-Processing-Time", strconv.Itoa(processingTime))

	c.Data(http.StatusOK, "application/octet-stream", secretData)
}

// sendError sends a standardized error response
func sendError(c *gin.Context, statusCode int, code string, message string) {
	errorResponse := models.ErrorResponse{
		Success: false,
		Error: models.ErrorDetail{
			Message: message,
			Details: map[string]interface{}{
				"code": code,
			},
		},
	}

	c.JSON(statusCode, errorResponse)
}
