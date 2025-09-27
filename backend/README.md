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

## 🚀 Production Deployment

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