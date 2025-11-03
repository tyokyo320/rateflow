package postgres

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/tyokyo320/rateflow/internal/infrastructure/config"
)

// NewConnection creates a new PostgreSQL database connection.
func NewConnection(cfg config.DatabaseConfig, log *slog.Logger) (*gorm.DB, error) {
	// Use silent logger to avoid GORM's verbose output
	gormLogger := logger.Default.LogMode(logger.Silent)

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.MaxConns / 2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migrate tables
	if err := db.AutoMigrate(&RateModel{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	log.Info("database connected",
		"host", cfg.Host,
		"database", cfg.Database,
		"max_conns", cfg.MaxConns,
	)

	return db, nil
}
