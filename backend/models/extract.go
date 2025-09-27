package models

type ExtractRequest struct {
	StegoAudio     []byte `json:"stego_audio"`
	NLsb           int    `json:"n_lsb"`
	StegoKey       string `json:"stego_key"`
	UseEncryption  bool   `json:"use_encryption"`
	UseRandomStart bool   `json:"use_random_start"`
	OutputFilename string `json:"output_filename"`
}

type ExtractResponse struct {
	SecretData   []byte `json:"secret_data"`
	Filename     string `json:"filename"`
	FileSize     int    `json:"file_size"`
	ExtractionOK bool   `json:"extraction_ok"`
}
