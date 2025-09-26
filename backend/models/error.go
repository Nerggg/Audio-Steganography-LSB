package models

import (
    "errors"
)

var ErrInvalidMP3 = errors.New("failed to decode audio data, not a valid MP3 file")

type ErrorResponse struct {
    Success bool        `json:"success"`
    Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
