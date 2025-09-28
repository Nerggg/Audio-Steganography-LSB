package models

type ExtractRequest struct {
	StegoAudio     []byte `json:"stego_audio"`
	StegoKey       string `json:"stego_key,omitempty"`
	OutputFilename string `json:"output_filename,omitempty"`
}

type ExtractResponse struct {
	SecretData   []byte `json:"secret_data"`
	Filename     string `json:"filename"`
	FileSize     int    `json:"file_size"`
	ExtractionOK bool   `json:"extraction_ok"`
}
