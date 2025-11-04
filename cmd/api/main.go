package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tyokyo320/rateflow/internal/application/query"
	"github.com/tyokyo320/rateflow/internal/infrastructure/config"
	"github.com/tyokyo320/rateflow/internal/infrastructure/logger"
	"github.com/tyokyo320/rateflow/internal/infrastructure/persistence/postgres"
	redisCache "github.com/tyokyo320/rateflow/internal/infrastructure/persistence/redis"
	httpHandler "github.com/tyokyo320/rateflow/internal/presentation/http"
	"github.com/tyokyo320/rateflow/internal/presentation/http/handler"

	_ "github.com/tyokyo320/rateflow/docs" // Import generated swagger docs
)

const (
	serviceName    = "rateflow-api"
	serviceVersion = "1.3.1"
)

// @title RateFlow API
// @version 1.3.1
// @description Multi-currency exchange rate tracking service powered by UnionPay
// @description
// @description This API provides access to historical and current exchange rates
// @description for 1,920+ currency pairs from UnionPay (12 base × 160+ target currencies).
//
// @contact.name API Support
// @contact.url https://github.com/tyokyo320/rateflow
// @contact.email support@example.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /
// @schemes http https
//
// @tag.name rates
// @tag.description Exchange rate operations
// @tag.name health
// @tag.description Health check operations

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.Logger)
	log = logger.WithContext(log, serviceName, serviceVersion)

	log.Info("starting API server",
		slog.String("version", serviceVersion),
		slog.String("environment", cfg.Server.Environment),
		slog.Int("port", cfg.Server.Port),
	)

	// Initialize database
	db, err := postgres.NewConnection(cfg.Database, log)
	if err != nil {
		log.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("failed to get database connection", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Error("failed to close database", "error", err)
		}
	}()

	// Initialize Redis cache
	cache := redisCache.NewCache(cfg.Redis, log)
	defer func() {
		if err := cache.Close(); err != nil {
			log.Error("failed to close redis", "error", err)
		}
	}()

	// Test Redis connection
	ctx := context.Background()
	if err := cache.Ping(ctx); err != nil {
		log.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	rateRepo := postgres.NewRateRepository(db, log)

	// Initialize query handlers
	getLatestHandler := query.NewGetLatestRateHandler(rateRepo, cache, log)
	listRatesHandler := query.NewListRatesHandler(rateRepo, log)

	// Initialize HTTP handlers
	rateHandler := handler.NewRateHandler(getLatestHandler, listRatesHandler, log)

	// Setup router
	router := httpHandler.SetupRouter(httpHandler.RouterConfig{
		RateHandler: rateHandler,
		Logger:      log,
		Environment: cfg.Server.Environment,
	})

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server exited")
}
