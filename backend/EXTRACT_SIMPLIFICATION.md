# Audio Steganography Backend - Extract Endpoint Simplification

## Summary of Changes

The extract endpoint has been **significantly simplified** based on your request. Now users only need to provide:
- **File**: The stego audio file (MP3 or WAV)  
- **Key**: The steganography key used during embedding

All other parameters are **automatically detected** from the embedded metadata.

## API Endpoints

### 1. `/api/v1/extract` (Simplified - Recommended)
**Input Required:**
- `stego_audio` (file): The audio file containing embedded data
- `stego_key` (string): The key used during embedding

**Auto-Detected:**
- LSB method (1-4 bits)
- Encryption usage (Vigenère cipher)
- Random start position usage
- Original filename

**Example cURL:**
```bash
curl -X POST http://localhost:8080/api/v1/extract \
  -F "stego_audio=@embedded_audio.mp3" \
  -F "stego_key=mySecretKey123" \
  -o extracted_file
```

### 2. `/api/v1/extract-manual` (Legacy - Advanced Users)
Still available for users who want explicit control over all parameters.

## Technical Implementation

### Auto-Detection Process

1. **Format Detection**: Automatically determines if file is MP3 or WAV
2. **Metadata Parsing**: Extracts embedding parameters from the stego metadata:
   - Tests LSB methods (1-4 bits) to find the correct one
   - Reads encryption flag from metadata
   - Reads random start flag from metadata
   - Recovers original filename

3. **Extraction**: Uses detected parameters to extract the hidden data

### Code Changes Made

1. **New Utils Function** (`service/utils.go`):
   - `parseMetadataWithNLsb()`: Auto-detects LSB method by testing different values

2. **Enhanced Steganography Service** (`service/steganography_service.go`):
   - `ExtractMessageAutoDetect()`: Main auto-detection extraction method

3. **Updated Handlers** (`handlers/service.go`):
   - `ExtractHandler()`: Simplified to only require file and key
   - `ExtractManualHandler()`: New legacy endpoint for manual parameter control

4. **Updated Interface** (`service/interfaces.go`):
   - Added auto-detection method to SteganographyService interface

## Benefits

✅ **User-Friendly**: Only file + key needed  
✅ **Backward Compatible**: Legacy endpoint still available  
✅ **Robust**: Handles both MP3 and WAV automatically  
✅ **Error Handling**: Clear messages if auto-detection fails  
✅ **Performance**: Efficient metadata parsing  

## Testing Status

- ✅ Code compiles successfully
- ✅ All unit tests pass
- ✅ Swagger documentation updated
- ✅ Both endpoints available in API

## Usage Notes

- The simplified endpoint works with files created by your embedding process
- If auto-detection fails, users can fall back to the manual endpoint
- All response headers include extraction details for debugging
- Both MP3 (ID3 PRIV tag) and WAV formats supported

The extraction process is now much more user-friendly while maintaining full functionality! 🎵