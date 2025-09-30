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
//
//	@Summary		Health Check
//	@Description	Returns the health status of the API service
//	@Tags			System
//	@Produce		json
//	@Success		200	{object}	HealthResponse			"Service is healthy"
//	@Failure		503	{object}	models.ErrorResponse	"Service unavailable"
//	@Router			/health [get]
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
//
//	@Summary		Calculate Audio Embedding Capacity
//	@Description	Calculates the maximum size of a secret file (in bytes) that can be embedded into an uploaded audio file (MP3 or WAV) using the multiple-LSB method. The capacity is returned for 1, 2, 3, and 4 LSBs.
//	@Tags			Steganography
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			audio	formData	file					true	"Audio file (MP3 or WAV) to calculate capacity for."
//	@Success		200		{object}	CapacityResponse		"Successfully calculated embedding capacity."
//	@Header			200		{int}		X-Processing-Time		"Time taken to process the request in milliseconds"
//	@Failure		400		{object}	models.ErrorResponse	"Bad Request: No file uploaded, file is not MP3/WAV, or file is corrupted."
//	@Failure		413		{object}	models.ErrorResponse	"File too large"
//	@Failure		500		{object}	models.ErrorResponse	"Internal Server Error: Failed to process the file."
//	@Router			/capacity [post]
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

// EmbedHandler embeds a secret file into an audio file using LSB steganography
// @Summary      Embed secret file into audio
// @Description  Embeds a secret file into the provided audio file using n-LSB steganography. Supports optional Vigenère encryption and random embedding start using a stego key. Metadata (filename, format, size, method, flags) is automatically stored inside the stego file.
// @Tags         Steganography
// @Accept       multipart/form-data
// @Produce      audio/mpeg
// @Param        audio            formData  file   true  "Cover audio file (MP3)"
// @Param        secret           formData  file   true  "Secret file to embed"
// @Param        lsb              formData  int    true  "Number of LSBs to use (1-4)"
// @Param        stego_key        formData  string false "Key for encryption and/or random start"
// @Param        use_encryption   formData  bool   false "Enable Vigenère encryption"
// @Param        use_random_start formData  bool   false "Enable random start embedding"
// @Param        output_filename  formData  string false "Output stego audio filename"
// @Success      200  {file}  binary  "Stego audio file with embedded secret"
// @Failure      400  {object}  models.ErrorResponse "Invalid input"
// @Failure      500  {object}  models.ErrorResponse "Processing error"
// @Router       /embed [post]
func (h *Handlers) EmbedHandler(c *gin.Context) {
	startTime := time.Now()

	// === Ambil file audio ===
	audioHeader, err := c.FormFile("audio")
	if err != nil {
		sendError(c, http.StatusBadRequest, "MISSING_FILES", "Audio file not provided")
		return
	}
	audioFile, _ := audioHeader.Open()
	defer audioFile.Close()
	audioData, _ := io.ReadAll(audioFile)

	// === Ambil file secret ===
	secretHeader, err := c.FormFile("secret")
	if err != nil {
		sendError(c, http.StatusBadRequest, "MISSING_FILES", "Secret file not provided")
		return
	}
	secretFile, _ := secretHeader.Open()
	defer secretFile.Close()
	secretData, _ := io.ReadAll(secretFile)

	// === Ambil parameter ===
	lsbStr := c.PostForm("lsb")
	lsb, err := strconv.Atoi(lsbStr)
	if err != nil || lsb < 1 || lsb > 4 {
		sendError(c, http.StatusBadRequest, "INVALID_LSB", "LSB value must be between 1 and 4")
		return
	}

	stegoKey := c.PostForm("stego_key")
	useEncryption := c.PostForm("use_encryption") == "true"
	useRandomStart := c.PostForm("use_random_start") == "true"

	if (useEncryption || useRandomStart) && stegoKey == "" {
		sendError(c, http.StatusBadRequest, "INVALID_STEGO_KEY", "Stego key is required when encryption or random start is enabled")
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

	// === Embed melalui service ===
	stegoAudio, psnr, err := h.steganographyService.EmbedMessage(embedReq, secretData, nil)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to embed data: "+err.Error())
		return
	}

	processingTime := int(time.Since(startTime).Milliseconds())
	outputFilename := c.PostForm("output_filename")
	if outputFilename == "" {
		outputFilename = "stego_audio.mp3"
	}

	// === Set header response ===
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", outputFilename))
	c.Header("X-PSNR-Value", fmt.Sprintf("%.2f", psnr))
	c.Header("X-Embedding-Method", fmt.Sprintf("%d-LSB", lsb))
	c.Header("X-Secret-Size", strconv.Itoa(len(secretData)))
	c.Header("X-Processing-Time", strconv.Itoa(processingTime))
	c.Header("X-Output-Format", "MP3")

	c.Data(http.StatusOK, "audio/mpeg", stegoAudio)
}

// ExtractHandler extracts a secret file from an audio file using LSB steganography
// @Summary      Extract secret file from audio
// @Description  Extracts a secret file that was previously embedded in an audio file using n-LSB steganography. Supports optional Vigenère decryption and random start. Automatically restores original filename and metadata.
// @Tags         Steganography
// @Accept       multipart/form-data
// @Produce      application/octet-stream
// @Param        stego_audio      formData  file   true  "Stego audio file (MP3 with embedded data)"
// @Param        stego_key        formData  string false "Key for decryption and/or random start"
// @Param        output_filename  formData  string false "Optional output filename override"
// @Success      200  {file}  binary  "Extracted secret file"
// @Failure      400  {object}  models.ErrorResponse "Invalid input"
// @Failure      500  {object}  models.ErrorResponse "Extraction error"
// @Router       /extract [post]
func (h *Handlers) ExtractHandler(c *gin.Context) {
	startTime := time.Now()

	stegoHeader, err := c.FormFile("stego_audio")
	if err != nil {
		sendError(c, http.StatusBadRequest, "MISSING_FILE", "Stego audio file not provided")
		return
	}

	stegoFile, _ := stegoHeader.Open()
	defer stegoFile.Close()
	stegoData, _ := io.ReadAll(stegoFile)

	stegoKey := c.PostForm("stego_key")
	outputFilename := c.PostForm("output_filename")

	extractReq := &models.ExtractRequest{
		StegoAudio:     stegoData,
		StegoKey:       stegoKey,
		OutputFilename: outputFilename,
	}

	secretData, filename, err := h.steganographyService.ExtractMessage(extractReq, stegoData)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "EXTRACTION_ERROR", "Failed to extract data: "+err.Error())
		return
	}

	processingTime := int(time.Since(startTime).Milliseconds())
	if outputFilename == "" {
		outputFilename = filename
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", outputFilename))
	c.Header("X-Extraction-Method", "Auto-detected LSB")
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
