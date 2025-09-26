package service

import (
    "io"
    "strconv"
    "net/http"
    "strings"
    "fmt"
    "path/filepath"

    "github.com/gin-gonic/gin"
    "github.com/Nerggg/Audio-Steganography-LSB/backend/controller"
    "github.com/Nerggg/Audio-Steganography-LSB/backend/models"
)

// CapacityHandler handles the capacity calculation request.
// @Summary Calculate Audio Embedding Capacity
// @Description Calculates the maximum size of a secret file (in bytes) that can be embedded into an uploaded MP3 file using the multiple-LSB method. The capacity is returned for 1, 2, 3, and 4 LSBs.
// @Tags Steganography
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "MP3 audio file to calculate capacity for."
// @Success 200 {object} models.CapacityResult "Successfully calculated embedding capacity."
// @Failure 400 {object} map[string]string "Bad Request: No file uploaded, file is not an MP3, or file is corrupted."
// @Failure 500 {object} map[string]string "Internal Server Error: Failed to process the file."
// @Router /api/capacity [post]
func CalculateCapacityHandler(c *gin.Context)  {
    // ambil berkas dari form request
    fileHeader, err := c.FormFile("audio")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file not provided."})
        return
    }

    file, err := fileHeader.Open()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file."})
        return
    }
    defer file.Close()

    // baca konten berkas ke dalam byte slice
    audioData, err := io.ReadAll(file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content."})
        return
    }

    // itung kapasitas
    capacities, err := controller.CalculateCapacity(audioData)
    if err != nil {
        if err == models.ErrInvalidMP3 {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    // kirim hasil
    c.JSON(http.StatusOK, capacities)
}

// EmbedHandler handles the message embedding request.
// @Summary Embed Secret Message into Audio File
// @Description Embeds a secret file into an MP3 cover audio file using the multiple-LSB steganography method. Supports encryption with extended Vigenère cipher, random starting position, and configurable LSB depth (1-4 bits). Returns the stego-audio file with PSNR quality measurement.
// @Tags Steganography
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param cover_audio formData file true "MP3 cover audio file to embed the secret message into"
// @Param secret_file formData file true "Secret file to be hidden (any file type and size)"
// @Param stego_key formData string true "Steganography key used for encryption and random seed generation (max 25 characters)"
// @Param n_lsb formData int true "Number of LSB bits to use for embedding (1, 2, 3, or 4)" Enums(1, 2, 3, 4)
// @Param use_encryption formData string true "Whether to encrypt the secret file using Vigenère cipher" Enums(true, false)
// @Param use_random_start formData string true "Whether to use random starting position for embedding" Enums(true, false)
// @Success 200 {file} file "Successfully embedded message into audio file. Returns stego-audio MP3 file."
// @Header 200 {string} Content-Disposition "attachment; filename=\"filename_stego.mp3\""
// @Header 200 {string} X-PSNR-Value "Peak Signal-to-Noise Ratio value (e.g., \"38.54\")"
// @Header 200 {string} Content-Type "application/octet-stream"
// @Failure 400 {object} models.ErrorResponse "Bad Request: Missing required parameters, invalid file format, or invalid parameter values"
// @Failure 413 {object} models.ErrorResponse "Payload Too Large: Secret file size exceeds embedding capacity"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error: Failed to process files or embed message"
// @Router /api/embed [post]
func EmbedHandler(c *gin.Context) {
	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Message: "Failed to parse multipart form",
			},
		})
		return
	}

	// Get cover audio file
	coverFile, coverHeader, err := c.Request.FormFile("cover_audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Message: "cover_audio file is required",
			},
		})
		return
	}
	defer coverFile.Close()

	coverData, err := io.ReadAll(coverFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Message: "Failed to read cover audio file",
			},
		})
		return
	}

	// Get secret file
	secretFile, secretHeader, err := c.Request.FormFile("secret_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Message: "secret_file is required",
			},
		})
		return
	}
	defer secretFile.Close()

	secretData, err := io.ReadAll(secretFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Message: "Failed to read secret file",
			},
		})
		return
	}

	// Parse parameters
	stegoKey := c.PostForm("stego_key")
	if stegoKey == "" || len(stegoKey) > 25 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Message: "stego_key is required and must be max 25 characters",
			},
		})
		return
	}

	nLsbStr := c.PostForm("n_lsb")
	nLsb, err := strconv.Atoi(nLsbStr)
	if err != nil || nLsb < 1 || nLsb > 4 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Message: "n_lsb must be an integer between 1 and 4",
			},
		})
		return
	}

	useEncryption := c.PostForm("use_encryption") == "true"
	useRandomStart := c.PostForm("use_random_start") == "true"

	// Create embed request
	embedReq := &models.EmbedRequest{
		CoverAudio:     coverData,
		SecretFile:     secretData,
		SecretFileName: secretHeader.Filename,
		StegoKey:       stegoKey,
		NLsb:           nLsb,
		UseEncryption:  useEncryption,
		UseRandomStart: useRandomStart,
	}

	// Check capacity
	capacity, err := controller.CalculateCapacity(coverData)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Message: "Invalid MP3 file: " + err.Error(),
			},
		})
		return
	}

	// Get capacity for selected n-LSB
	var maxCapacity int
	switch nLsb {
	case 1:
		maxCapacity = capacity.OneLSB
	case 2:
		maxCapacity = capacity.TwoLSB
	case 3:
		maxCapacity = capacity.ThreeLSB
	case 4:
		maxCapacity = capacity.FourLSB
	}

	// Calculate required space (secret file + metadata)
	metadata := controller.CreateMetadata(secretHeader.Filename, len(secretData), useEncryption, useRandomStart, nLsb)
	totalRequiredSize := len(secretData) + len(metadata)

	// Apply encryption if requested
	finalSecretData := secretData
	if useEncryption {
		finalSecretData = controller.VigenereCipher(secretData, stegoKey, true)
	}

	totalRequiredSize = len(finalSecretData) + len(metadata)

	if totalRequiredSize > maxCapacity {
		c.JSON(http.StatusRequestEntityTooLarge, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Message: "Secret file size exceeds the embedding capacity for the selected options.",
				Details: map[string]interface{}{
					"max_capacity_bytes": maxCapacity,
					"secret_file_bytes":  totalRequiredSize,
				},
			},
		})
		return
	}

	// Perform embedding
	stegoAudio, psnr, err := controller.EmbedMessage(embedReq, finalSecretData, metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Message: "Failed to embed message: " + err.Error(),
			},
		})
		return
	}

	// Set response headers
	filename := strings.TrimSuffix(coverHeader.Filename, filepath.Ext(coverHeader.Filename)) + "_stego.mp3"
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("X-PSNR-Value", fmt.Sprintf("%.2f", psnr))
	c.Header("Content-Type", "application/octet-stream")

	// Return stego audio
	c.Data(http.StatusOK, "application/octet-stream", stegoAudio)
}
