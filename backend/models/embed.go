package models

type EmbedRequest struct {
    CoverAudio     []byte
    SecretFile     []byte
    SecretFileName string
    StegoKey       string
    NLsb           int
    UseEncryption  bool
    UseRandomStart bool
}

type EmbedResponse struct {
    StegoAudio []byte
    PSNR       float64
}
