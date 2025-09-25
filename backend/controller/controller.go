package controller

import (
    "bytes"
    "io"
    "errors"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/hajimehoshi/go-mp3"
)

// ErrInvalidMP3 is returned when the provided data cannot be decoded as an MP3 file.
var ErrInvalidMP3 = errors.New("failed to decode audio data, not a valid MP3 file")

// CapacityResult holds the calculated embedding capacity in bytes for different LSB modes.
type CapacityResult struct {
    OneLSB   int `json:"1_lsb"`
    TwoLSB   int `json:"2_lsb"`
    ThreeLSB int `json:"3_lsb"`
    FourLSB  int `json:"4_lsb"`
}

// CalculateCapacity decodes MP3 audio data to calculate the total number of available
// audio samples and determines the maximum secret message size (in bytes) that can be
// embedded using 1, 2, 3, and 4 LSBs.
func CalculateCapacity(audioData []byte) (*CapacityResult, error) {
    // 1. Buat reader dari data MP3 yang diinput.
    audioReader := bytes.NewReader(audioData)

    // 2. Dekode stream MP3 menjadi stream audio mentah (PCM).
    // Pustaka ini sesuai dengan spesifikasi yang mengizinkan penggunaan pustaka
    // pengolahan audio yang sudah ada.
    decoder, err := mp3.NewDecoder(audioReader)
    if err != nil {
        return nil, ErrInvalidMP3
    }

    // 3. Baca seluruh sampel PCM yang telah didekode.
    // Ini adalah bagian "audio data samples" yang disebutkan dalam spesifikasi.
    pcmData, err := io.ReadAll(decoder)
    if err != nil {
        // Menangani kemungkinan error saat membaca stream
        return nil, errors.New("could not read decoded audio stream: " + err.Error())
    }

    // Setiap sampel audio biasanya berukuran 2 byte (16-bit).
    // Jadi, jumlah total sampel adalah total byte PCM dibagi 2.
    totalSamples := len(pcmData) / 2

    // Jika pcmData ganjil, itu tidak diharapkan, tapi kita bisa tangani.
    if len(pcmData)%2 != 0 {
        totalSamples = (len(pcmData) - 1) / 2
    }

    if totalSamples == 0 {
        return nil, ErrInvalidMP3
    }

    // 4. Hitung kapasitas dalam byte berdasarkan rumus:
    // Kapasitas (byte) = (Total Sampel * Jumlah LSB) / 8 bit per byte
    capacities := &CapacityResult{
        OneLSB:   (totalSamples * 1) / 8,
        TwoLSB:   (totalSamples * 2) / 8,
        ThreeLSB: (totalSamples * 3) / 8,
        FourLSB:  (totalSamples * 4) / 8,
    }

    return capacities, nil
}

// @Summary ping example
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]
func Helloworld(g *gin.Context)  {
   g.JSON(http.StatusOK,"helloworld")
}

// @Summary ping example
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld2 [get]
func Helloworld2(g *gin.Context)  {
   g.JSON(http.StatusOK,"helloworld")
}
