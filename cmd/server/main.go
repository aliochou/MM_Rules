package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mm-rules/matchmaking/internal/allocation"
	"github.com/mm-rules/matchmaking/internal/api"
	"github.com/mm-rules/matchmaking/internal/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	if err := loadConfig(); err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	setupLogging()

	logger := logrus.New()
	logger.Info("Starting MM-Rules Matchmaking Server")

	// Initialize Redis storage
	redisAddr := viper.GetString("redis.addr")
	redisPassword := viper.GetString("redis.password")
	redisDB := viper.GetInt("redis.db")

	redisStorage := storage.NewRedisStorage(redisAddr, redisPassword, redisDB)
	defer redisStorage.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisStorage.Ping(ctx); err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	logger.Info("Connected to Redis")

	// Initialize allocator
	webhookURL := viper.GetString("allocation.webhook_url")
	var allocator allocation.Allocator = allocation.NewAllocator(webhookURL)

	// Initialize API handler
	handler := api.NewHandler(redisStorage, allocator, logger)

	// Setup router
	router := setupRouter(handler)

	// Get server configuration
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/mm-rules")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("allocation.webhook_url", "http://localhost:8081/allocate")
	viper.SetDefault("log.level", "info")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults
	}

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MM_RULES")

	return nil
}

func setupLogging() {
	level := viper.GetString("log.level")
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}

	logrus.SetLevel(logLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func setupRouter(handler *api.Handler) *gin.Engine {
	// Set Gin mode
	if viper.GetString("server.mode") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", handler.HealthCheck)

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := router.Group("/api/v1")
	{
		// Match requests
		api.POST("/match-request", handler.CreateMatchRequest)
		api.GET("/match-status/:request_id", handler.GetMatchStatus)

		// Game configuration
		api.POST("/rules/:game_id", handler.CreateGameConfig)

		// Matchmaking processing
		api.POST("/process-matchmaking/:game_id", handler.ProcessMatchmaking)
		api.POST("/allocate-sessions/:game_id", handler.AllocateSessions)

		// Statistics
		api.GET("/stats", handler.GetStats)
	}

	return router
} 