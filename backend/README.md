# Audio Steganography LSB API

A high-performance REST API for audio steganography using Least Significant Bit (LSB) encoding, built with Go and documented with Swagger.

## 🚀 Features

- **Swagger Documentation**: Complete API documentation with interactive UI
- **Multiple LSB Support**: 1-4 bit LSB encoding for flexible capacity/quality trade-offs
- **Comprehensive Endpoints**:
  - `/api/v1/health` - Health check and system status
  - `/api/v1/capacity` - Calculate embedding capacity for audio files
  - `/api/v1/embed` - Embed secret data into audio files
  - `/api/v1/extract` - Extract hidden data from stego audio files
- **Best Practice Implementation**:
  - Graceful shutdown
  - CORS middleware with security headers
  - Request tracing and logging
  - File size limits and validation
  - Structured error responses
- **Interactive Documentation**: Swagger UI with live API testing
- **Production Ready**: Proper middleware, error handling, and security headers

## 🏗️ Architecture

This API follows Go best practices with a clean layered architecture, dependency injection, and Swagger documentation:

```
backend/
├── main.go                 # Application entrypoint with service initialization
├── handlers/
│   └── service.go        # HTTP handlers with dependency injection
├── service/
│   ├── interfaces.go     # Service layer interfaces for dependency injection
│   └── controller.go     # Business logic implementations
├── models/
│   ├── capacity.go       # Data models for capacity calculations
│   ├── embed.go          # Data models for embedding operations
│   ├── extract.go        # Data models for extraction operations
│   └── error.go          # Error response models
└── docs/                 # Generated Swagger documentation
    ├── docs.go           # Generated Go documentation
    ├── swagger.json      # Generated JSON specification
    └── swagger.yaml      # Generated YAML specification
```

### Key Principles

1. **Dependency Injection**: Services are injected into handlers for better testability
2. **Interface-Driven Design**: Clean service interfaces separate concerns
3. **Layered Architecture**: Handlers → Services → Data Models
4. **Documentation-Driven**: Swagger annotations in code generate docs
5. **Type Safety**: Go structs ensure consistent API responses
6. **Best Practices**: Security, error handling, and observability built-in

### Service Layer Pattern

The service layer follows Go best practices:

- **Interfaces First**: `interfaces.go` defines service contracts
- **Dependency Injection**: Services are injected into handlers in `main.go`
- **Single Responsibility**: Each service handles one domain (steganography, cryptography, audio)
- **Testability**: Mock services can be easily created for testing
- **Modularity**: Services can be independently developed and tested

**Communication Flow:**
```
HTTP Request → Handlers (HTTP logic) → Services (Business logic) → Models (Data) → HTTP Response
```

## 🎵 **Steganography Features**

### **✅ Implemented Features**

- **LSB Steganography**: Supports 1-4 bit LSB embedding and extraction
- **MP3 Audio Support**: Decodes MP3 files to PCM for processing
- **Capacity Calculation**: Accurately calculates embedding capacity with metadata overhead
- **Metadata Embedding**: Stores file information (name, size, flags) with secret data
- **Encryption Support**: XOR-based Vigenère cipher for data protection
- **Random Start Position**: Pseudo-random embedding position based on stego key
- **PSNR Calculation**: Measures audio quality after embedding
- **Comprehensive Validation**: Input validation, file size limits, error handling
- **Security Best Practices**: Key length requirements, proper error messages

### **Core Algorithms**

1. **Embedding Process**:
   ```
   MP3 → PCM → Create Metadata → Encrypt (optional) → LSB Embed → Calculate PSNR
   ```

2. **Extraction Process**:
   ```
   MP3 → PCM → Extract LSB → Parse Metadata → Decrypt (optional) → Return Secret
   ```

3. **Supported Parameters**:
   - LSB bits: 1-4
   - Audio format: MP3 (decoded to 16-bit PCM)
   - Max file sizes: 100MB audio, 50MB secret
   - Encryption: XOR-based with repeating key
   - Random positioning: Seed-based pseudo-random start

### **⚠️ Current Limitations**

- **No MP3 Re-encoding**: Returns raw PCM data instead of MP3 (technical limitation)
- **16-bit PCM Assumption**: Assumes 16-bit audio samples
- **No Audio Metadata**: Duration, bitrate calculated from headers (not implemented)

## 🛠️ Setup & Installation

### Prerequisites

- Go 1.21 or higher
- `swag` tool for Swagger generation: `go install github.com/swaggo/swag/cmd/swag@latest`

### Quick Start

1. **Clone and navigate to backend:**
   ```powershell
   cd backend/
   ```

2. **Install dependencies:**
   ```powershell
   go mod tidy
   ```

3. **Build and run:**
   ```powershell
   .\dev.ps1 build
   .\dev.ps1 run
   ```

4. **Access the API:**
   - Health Check: http://localhost:8080/api/v1/health
   - Swagger UI: http://localhost:8080/swagger/index.html

### Development Scripts

The included PowerShell scripts streamline development:

```powershell
# Build application
.\dev.ps1 build

# Run development server
.\dev.ps1 run

# Run API tests
.\dev.ps1 test

# Regenerate Swagger documentation
.\dev.ps1 docs

# Clean build artifacts
.\dev.ps1 clean

# Open Swagger documentation
.\dev.ps1 swagger
```

## 📖 API Documentation

### Health Check
```http
GET /api/v1/health
```
Returns system health and status information.

### Calculate Capacity
```http
POST /api/v1/capacity
Content-Type: multipart/form-data

audio: [MP3 file]
```
Calculates maximum embedding capacity for different LSB levels.

### Embed Data
```http
POST /api/v1/embed
Content-Type: multipart/form-data

audio: [MP3 file]
secret: [File to hide]
lsb: [1-4]
use_encryption: [true/false] (optional)
use_random_start: [true/false] (optional) 
stego_key: [string] (optional)
output_filename: [string] (optional)
```
Embeds secret data into audio file using LSB steganography.

**Response Headers:**
- `X-PSNR-Value`: Audio quality metric
- `X-Embedding-Method`: LSB method used
- `X-Secret-Size`: Size of embedded data
- `Content-Disposition`: Download filename

### Extract Data
```http
POST /api/v1/extract
Content-Type: multipart/form-data

stego_audio: [MP3 file with hidden data]
lsb: [1-4]
output_filename: [string] (optional)
```
Extracts hidden data from stego audio file.

**Response Headers:**
- `X-Extraction-Method`: LSB method used
- `X-Secret-Size`: Size of extracted data
- `Content-Disposition`: Download filename

## 🔧 Configuration

Environment variables (`.env` file):

```env
# Server Configuration
GIN_MODE=debug          # debug, release, test
PORT=8080              # Server port

# CORS Configuration  
CORS_ORIGINS=http://localhost:3000,http://localhost:5173
```

## 🧪 Testing

### Automated Tests
```powershell
.\dev.ps1 test
```

### Manual Testing
```powershell
# Start server
.\dev.ps1 run

# In another terminal, run tests
.\test-api.ps1
```

### Test with curl
```bash
# Health check
curl -X GET http://localhost:8080/api/v1/health

# Calculate capacity
curl -X POST http://localhost:8080/api/v1/capacity \
  -F "audio=@sample.mp3"

# Embed data
curl -X POST http://localhost:8080/api/v1/embed \
  -F "audio=@cover.mp3" \
  -F "secret=@secret.txt" \
  -F "lsb=2" \
  -o stego_audio.mp3

# Extract data  
curl -X POST http://localhost:8080/api/v1/extract \
  -F "stego_audio=@stego_audio.mp3" \
  -F "lsb=2" \
  -o extracted_secret.txt
```

## 🔒 Security Features

- **Input Validation**: File type and size limits
- **Security Headers**: OWASP recommended headers
- **CORS Protection**: Configurable origin restrictions  
- **Request Tracing**: Unique request IDs for debugging
- **Error Sanitization**: No sensitive data in error responses
- **File Size Limits**: 100MB for audio, 50MB for secrets

## 📊 Monitoring & Observability

- **Health Endpoint**: System status and dependencies
- **Request Logging**: Structured logs with timing
- **Error Tracking**: Categorized error responses with codes
- **Performance Headers**: Processing time in responses
- **Graceful Shutdown**: Proper cleanup on termination

## 🔍 **Comprehensive Debugging & Logging**

The application includes **extensive logging** throughout all layers for easy debugging and monitoring:

### **Log Levels & Format**

```bash
# Log format: [LEVEL] [Request-ID] Component: Message
2025/09/27 15:08:05 [INFO] Audio Steganography API Server Starting...
2025/09/27 15:08:05 [DEBUG] [req_1695820085123] EmbedHandler: Starting embed request from 127.0.0.1
2025/09/27 15:08:05 [ERROR] [req_1695820085123] EmbedMessage: Failed to decode MP3: invalid format
```

### **Service Layer Logging**

**Capacity Calculation:**
```bash
[DEBUG] CalculateCapacity: Starting capacity calculation for audio data (size: 2048576 bytes)
[DEBUG] CalculateCapacity: Successfully decoded MP3 to PCM (size: 1920000 bytes)
[WARN]  CalculateCapacity: Odd PCM data length detected, adjusted samples to 959999
[INFO]  CalculateCapacity: Completed successfully (total_samples: 960000, capacities: 1LSB=119940, 2LSB=239880, duration: 45ms)
```

**Embed Operation:**
```bash
[DEBUG] EmbedMessage: Starting embed operation (audio_size: 2048576 bytes, secret_size: 1024 bytes, nLSB: 2, encryption: true)
[DEBUG] EmbedMessage: Successfully decoded MP3 to PCM (pcm_size: 1920000 bytes)  
[DEBUG] EmbedMessage: Capacity check (total_samples: 960000, data_size: 1084 bytes, max_capacity: 239880 bytes)
[DEBUG] EmbedMessage: Applying Vigenère cipher encryption with key length: 8
[DEBUG] EmbedMessage: Data encrypted (original: 1024 bytes, encrypted: 1024 bytes)
[DEBUG] EmbedMessage: Final data to embed: 1084 bytes (metadata: 60 + data: 1024)
[DEBUG] EmbedMessage: Converted to bit array: 8672 bits
[DEBUG] EmbedMessage: Using random start position: 1247
[DEBUG] EmbedMessage: Starting LSB embedding...
[DEBUG] EmbedMessage: LSB embedding completed successfully
[INFO]  EmbedMessage: PSNR calculated: 42.35 dB
[DEBUG] EmbedMessage: Encoding to MP3 format...
[DEBUG] EncodeToMP3: Starting MP3 encoding (pcm_size: 1920000 bytes, sample_rate: 44100)
[DEBUG] EncodeToMP3: WAV intermediate created (size: 1920044 bytes)
[DEBUG] convertWAVToMP3WithFFmpeg: Starting ffmpeg conversion (input: 1920044 bytes)
[WARN]  EncodeToMP3: ffmpeg conversion failed, falling back to WAV: executable not found
[INFO]  EmbedMessage: Embed operation completed successfully (output_size: 1920044 bytes, psnr: 42.35 dB, duration: 156ms)
```

**Extract Operation:**
```bash
[DEBUG] ExtractMessage: Starting extraction operation (audio_size: 1920044 bytes, use_encryption: true, use_random_start: true)
[DEBUG] ExtractMessage: Decoding as MP3 format
[DEBUG] ExtractMessage: Successfully decoded MP3 to PCM (pcm_size: 1920000 bytes)
[DEBUG] ExtractMessage: Total samples available: 960000
[INFO]  ExtractMessage: Extraction completed successfully (extracted: 1024 bytes, filename: 'secret.txt', duration: 89ms)
```

### **Handler Layer Logging**

**HTTP Request Tracking:**
```bash
[INFO]  [req_1695820085123] EmbedHandler: Starting embed request from 192.168.1.100
[DEBUG] [req_1695820085123] EmbedHandler: Audio file 'cover.mp3' (size: 2048576 bytes)
[DEBUG] [req_1695820085123] EmbedHandler: Secret file 'document.pdf' (size: 1024 bytes)
[ERROR] [req_1695820085123] EmbedHandler: Invalid LSB parameter 'abc', expected 1-4
[INFO]  [req_1695820085123] CalculateCapacityHandler: Starting capacity calculation request from 10.0.0.15
```

### **Server Lifecycle Logging**

**Startup Sequence:**
```bash
[INFO]  Audio Steganography API Server Starting...
[WARN]  No .env file found, using environment variables
[INFO]  Gin mode set to: release
[DEBUG] Gin router created
[DEBUG] Middleware configured  
[DEBUG] Initializing services...
[INFO]  All services initialized successfully
[INFO]  Handlers initialized with dependency injection
[INFO]  Audio Steganography API Server is ready to accept connections
[INFO]  Starting HTTP server on port 8080
[INFO]  Swagger documentation available at: http://localhost:8080/swagger/index.html
[INFO]  Health check endpoint: http://localhost:8080/api/v1/health
```

**Shutdown Sequence:**
```bash
[INFO] Received shutdown signal, shutting down server gracefully...
[INFO] Server gracefully stopped
```

### **Error Debugging**

**Detailed Error Context:**
```bash
[ERROR] [req_1695820085124] CalculateCapacity: Failed to create MP3 decoder: invalid header
[ERROR] [req_1695820085125] EmbedMessage: Insufficient capacity - need 50000 bytes, only 25000 bytes available
[ERROR] [req_1695820085126] ExtractMessage: Unsupported audio format - not MP3 or WAV
[ERROR] [req_1695820085127] EmbedMessage: Failed to encode to MP3: ffmpeg process failed
```

### **Log Level Configuration**

Set the `LOG_LEVEL` environment variable to control verbosity:

```bash
# Show all logs (default for development)
export LOG_LEVEL=DEBUG

# Production logging (recommended)
export LOG_LEVEL=INFO  

# Only warnings and errors
export LOG_LEVEL=WARN

# Only critical errors
export LOG_LEVEL=ERROR
```

### **Request Tracing**

Each HTTP request gets a unique ID for complete request tracing:
- **Request ID Format**: `req_<timestamp_nanoseconds>`
- **Cross-Layer Tracking**: Same ID used in handlers and services
- **Easy Debugging**: Grep logs by request ID to see complete request lifecycle

### **Performance Monitoring**

All operations include timing information:
- **Duration Tracking**: Start-to-finish timing for all major operations
- **Performance Headers**: `X-Processing-Time` header in responses
- **Bottleneck Identification**: Detailed timing at each step

This comprehensive logging makes debugging, monitoring, and troubleshooting extremely easy!

## � **API Usage Examples**

### 1. **Calculate Capacity**
```bash
curl -X POST http://localhost:8080/api/v1/capacity \
  -F "audio=@cover.mp3"
```

**Response:**
```json
{
  "capacities": {
    "OneLSB": 12580,
    "TwoLSB": 25160,  
    "ThreeLSB": 37740,
    "FourLSB": 50320
  },
  "file_info": {
    "filename": "cover.mp3",
    "size_bytes": 2048576
  }
}
```

### 2. **Embed Secret Data**
```bash
curl -X POST http://localhost:8080/api/v1/embed \
  -F "audio=@cover.mp3" \
  -F "secret=@secret.txt" \
  -F "lsb=2" \
  -F "use_encryption=true" \
  -F "stego_key=mysecretkey" \
  -F "use_random_start=true" \
  --output stego.mp3
```

### 3. **Extract Secret Data**
```bash
curl -X POST http://localhost:8080/api/v1/extract \
  -F "stego_audio=@stego.mp3" \
  -F "lsb=2" \
  -F "use_encryption=true" \
  -F "stego_key=mysecretkey" \
  -F "use_random_start=true" \
  --output extracted_secret.txt
```

### 4. **Health Check**
```bash
curl http://localhost:8080/api/v1/health
```

## �🚀 Production Deployment

### Build for Production
```powershell
$env:GIN_MODE="release"
.\dev.ps1 build
```

### Docker Support
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o audio-steganography-api .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/audio-steganography-api .
EXPOSE 8080
CMD ["./audio-steganography-api"]
```

### Environment Setup
```env
GIN_MODE=release
PORT=8080
CORS_ORIGINS=https://yourdomain.com
```

## 🔄 API Evolution

The API follows semantic versioning and maintains backward compatibility:

1. **Update Handler Code**: Modify handlers in `handlers/service.go`
2. **Update Swagger Annotations**: Add/modify @Summary, @Description, etc.
3. **Regenerate Docs**: Run `.\dev.ps1 docs` or `swag init`
4. **Test**: Run `.\dev.ps1 test`
5. **Deploy**: Build and deploy new version

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Make changes following the architecture patterns
4. Update OpenAPI spec if needed
5. Run tests: `.\dev.ps1 test`
6. Commit changes: `git commit -m 'Add amazing feature'`
7. Push branch: `git push origin feature/amazing-feature`
8. Submit pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/api/v1/health  
- **Issues**: Please report bugs and feature requests via GitHub issues

---

Built with ❤️ using Go, OpenAPI 3.0, and oapi-codegen