package controller

import (
    "bytes"
    "io"
    "errors"
    "github.com/hajimehoshi/go-mp3"
)

var ErrInvalidMP3 = errors.New("failed to decode audio data, not a valid MP3 file")

type CapacityResult struct {
    OneLSB   int `json:"1_lsb"`
    TwoLSB   int `json:"2_lsb"`
    ThreeLSB int `json:"3_lsb"`
    FourLSB  int `json:"4_lsb"`
}

func CalculateCapacity(audioData []byte) (*CapacityResult, error) {
    audioReader := bytes.NewReader(audioData)

    decoder, err := mp3.NewDecoder(audioReader)
    if err != nil {
        return nil, ErrInvalidMP3
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
        return nil, ErrInvalidMP3
    }

    capacities := &CapacityResult{
        OneLSB:   (totalSamples * 1) / 8,
        TwoLSB:   (totalSamples * 2) / 8,
        ThreeLSB: (totalSamples * 3) / 8,
        FourLSB:  (totalSamples * 4) / 8,
    }

    return capacities, nil
}
