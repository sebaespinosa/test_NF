package main

// @title Irrigation Analytics API
// @version 0.0.1
// @description API for managing irrigation analytics within an agricultural platform.
// @BasePath /

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sebaespinosa/test_NF/config"
	"github.com/sebaespinosa/test_NF/controller"
	"github.com/sebaespinosa/test_NF/internal/database"
	"github.com/sebaespinosa/test_NF/internal/logging"
	"github.com/sebaespinosa/test_NF/internal/middleware"
	"github.com/sebaespinosa/test_NF/internal/observability"
	"github.com/sebaespinosa/test_NF/repository"
	"github.com/sebaespinosa/test_NF/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := logging.New(cfg.Server.Env)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	logger.Info(
		"application starting",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.String("env", cfg.Server.Env),
	)

	// Initialize Jaeger tracing
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdown, err := observability.InitJaeger(ctx, &cfg.Jaeger, &cfg.Service)
	if err != nil {
		logger.Fatal("failed to initialize jaeger", zap.Error(err))
	}
	defer func() {
		if shutdown != nil {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			if err := shutdown(shutdownCtx); err != nil {
				logger.Error("failed to shutdown tracer", zap.Error(err))
			}
		}
	}()

	// Initialize database
	db, err := database.Initialize(&cfg.Database)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}

	// Initialize repositories
	healthRepo := repository.NewHealthRepository(db)

	// Initialize services
	healthService := service.NewHealthService(healthRepo, logger, cfg.Service.Version)

	// Initialize controllers
	healthController := controller.NewHealthController(healthService)

	// Setup Gin router
	router := gin.Default()

	// Apply observability middleware
	router.Use(middleware.TraceMiddleware(logger))

	// Register routes
	router.GET("/health", healthController.GetHealth)

	// Swagger docs
	router.StaticFile("/docs/swagger.json", "./documentation/swagger.json")
	swaggerURL := ginSwagger.URL("/docs/swagger.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL))

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("server starting", zap.Uint16("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down server")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}

	logger.Info("server stopped")
}
