package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	docs "github.com/Nerggg/Audio-Steganography-LSB/backend/docs"
	"github.com/Nerggg/Audio-Steganography-LSB/backend/handlers"
	"github.com/Nerggg/Audio-Steganography-LSB/backend/service"
)

// @BasePath /api/v1

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Set gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.New()

	// Configure best-practice middleware
	setupMiddleware(r)

	// Initialize services with dependency injection
	steganographyService := service.NewSteganographyService()
	cryptographyService := service.NewCryptographyService()
	audioService := service.NewAudioService()

	// Initialize handlers with injected services
	h := handlers.NewHandlers(steganographyService, cryptographyService, audioService)

	// Set up Swagger documentation
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register API routes with dependency-injected handlers
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", h.HealthHandler)
		v1.POST("/capacity", h.CalculateCapacityHandler)
		v1.POST("/embed", h.EmbedHandler)
		v1.POST("/extract", h.ExtractHandler)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with best practices
	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}

// setupMiddleware configures all necessary middleware following best practices
func setupMiddleware(r *gin.Engine) {
	// Recovery middleware recovers from any panics and writes a 500
	r.Use(gin.Recovery())

	// Logger middleware with custom format
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// CORS middleware with secure configuration
	corsConfig := cors.Config{
		AllowOrigins: getAllowedOrigins(),
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"X-API-Key",
			"X-Trace-Id",
		},
		ExposeHeaders: []string{
			"Content-Disposition",
			"X-PSNR-Value",
			"X-Embedding-Method",
			"X-Extraction-Method",
			"X-Secret-Size",
			"X-Processing-Time",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// Security headers middleware
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})

	// Request ID middleware for tracing
	r.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Trace-Id")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Trace-Id", requestID)
		c.Set("trace_id", requestID)
		c.Next()
	})

	// File size limit middleware for multipart requests
	r.Use(func(c *gin.Context) {
		if c.ContentType() == "multipart/form-data" {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 100*1024*1024) // 100MB limit
		}
		c.Next()
	})
}

// getAllowedOrigins returns allowed CORS origins based on environment
func getAllowedOrigins() []string {
	if origins := os.Getenv("CORS_ORIGINS"); origins != "" {
		return strings.Split(origins, ",")
	}

	// Default origins for development
	return []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:5173",
	}
}

// generateRequestID generates a simple request ID for tracing
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
