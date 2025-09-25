package service

import (
    "io"
    "errors"
    "net/http"
    "github.com/gin-gonic/gin"
   "github.com/Nerggg/Audio-Steganography-LSB/backend/controller"
)

var ErrInvalidMP3 = errors.New("failed to decode audio data, not a valid MP3 file")

// CapacityHandler handles the capacity calculation request.
// @Summary Calculate Audio Embedding Capacity
// @Description Calculates the maximum size of a secret file (in bytes) that can be embedded into an uploaded MP3 file using the multiple-LSB method. The capacity is returned for 1, 2, 3, and 4 LSBs.
// @Tags Steganography
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "MP3 audio file to calculate capacity for."
// @Success 200 {object} controller.CapacityResult "Successfully calculated embedding capacity."
// @Failure 400 {object} map[string]string "Bad Request: No file uploaded, file is not an MP3, or file is corrupted."
// @Failure 500 {object} map[string]string "Internal Server Error: Failed to process the file."
// @Router /api/capacity [post]
func CalculateCapacityHandler(c *gin.Context)  {
    // Ambil berkas dari form request
    fileHeader, err := c.FormFile("audio")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file not provided."})
        return
    }

    // Buka berkas yang diunggah
    file, err := fileHeader.Open()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file."})
        return
    }
    defer file.Close()

    // Baca konten berkas ke dalam byte slice
    audioData, err := io.ReadAll(file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content."})
        return
    }

    // Panggil fungsi inti untuk menghitung kapasitas
    capacities, err := controller.CalculateCapacity(audioData)
    if err != nil {
        if err == ErrInvalidMP3 {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    // Kirim hasil kapasitas sebagai JSON
    c.JSON(http.StatusOK, capacities)
}
