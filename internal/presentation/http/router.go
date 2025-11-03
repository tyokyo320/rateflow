package http

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/tyokyo320/rateflow/internal/presentation/http/handler"
	"github.com/tyokyo320/rateflow/internal/presentation/http/middleware"
)

// RouterConfig holds router configuration.
type RouterConfig struct {
	RateHandler *handler.RateHandler
	Logger      *slog.Logger
	Environment string // dev, staging, prod
}

// SetupRouter creates and configures the HTTP router.
func SetupRouter(cfg RouterConfig) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Environment == "prod" || cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.Environment == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create router
	router := gin.New()

	// Apply global middleware
	router.Use(middleware.Recovery(cfg.Logger))
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(cfg.Logger))
	router.Use(middleware.CORS())

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	router.GET("/health", cfg.RateHandler.Health)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Rate endpoints
		rates := v1.Group("/rates")
		{
			rates.GET("/latest", cfg.RateHandler.GetLatest)
			rates.GET("", cfg.RateHandler.GetByDate)
			rates.GET("/list", cfg.RateHandler.List)
		}
	}

	// Legacy API routes (for backward compatibility)
	api := router.Group("/api")
	{
		rates := api.Group("/rates")
		{
			rates.GET("/latest", cfg.RateHandler.GetLatest)
			rates.GET("", cfg.RateHandler.GetByDate)
			rates.GET("/list", cfg.RateHandler.List)
		}
	}

	return router
}
