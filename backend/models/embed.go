package models

// SteganographyMethod represents the type of steganography method to use
type SteganographyMethod string

const (
	MethodLSB    SteganographyMethod = "lsb"
	MethodParity SteganographyMethod = "parity"
)

// IsValid checks if the steganography method is valid
func (sm SteganographyMethod) IsValid() bool {
	return sm == MethodLSB || sm == MethodParity
}

// String returns the string representation of the method
func (sm SteganographyMethod) String() string {
	return string(sm)
}

// GetSupportedMethods returns a list of supported steganography methods
func GetSupportedMethods() []SteganographyMethod {
	return []SteganographyMethod{MethodLSB, MethodParity}
}

type EmbedRequest struct {
	CoverAudio     []byte
	SecretFile     []byte
	SecretFileName string
	StegoKey       string
	Method         SteganographyMethod // "lsb" or "parity"
	NLsb           int                 // Only used for LSB method (1-4)
	UseEncryption  bool
	UseRandomStart bool
}

type EmbedResponse struct {
	StegoAudio []byte
	PSNR       float64
}
