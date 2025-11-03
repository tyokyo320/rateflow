// Package config provides configuration management for the application.
// It supports loading from JSON files and environment variables (env vars override file values).
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration.
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Logger   LoggerConfig   `json:"logger"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout"`
	Environment  string        `json:"environment"` // dev, staging, prod
}

// DatabaseConfig holds PostgreSQL configuration.
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslMode"`
	Timezone string `json:"timezone"`
	MaxConns int    `json:"maxConns"`
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// LoggerConfig holds logging configuration.
type LoggerConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // json, text
}

// Load loads configuration from file and environment variables.
// Environment variables take precedence over file values.
func Load() (*Config, error) {
	cfg := defaultConfig()

	// 1. Load from file if CONFIG_PATH is set
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		if err := loadFromFile(cfg, configPath); err != nil {
			return nil, fmt.Errorf("load config file: %w", err)
		}
	}

	// 2. Override with environment variables
	overrideFromEnv(cfg)

	// 3. Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			Environment:  "dev",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			SSLMode:  "disable",
			Timezone: "Asia/Tokyo",
			MaxConns: 25,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		Logger: LoggerConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// loadFromFile loads configuration from a JSON file.
func loadFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, cfg)
}

// overrideFromEnv overrides configuration with environment variables.
func overrideFromEnv(cfg *Config) {
	// Server
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("ENVIRONMENT"); v != "" {
		cfg.Server.Environment = v
	}

	// Database
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Database = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.Database.SSLMode = v
	}
	if v := os.Getenv("DB_TIMEZONE"); v != "" {
		cfg.Database.Timezone = v
	}
	if v := os.Getenv("DB_MAX_CONNS"); v != "" {
		if maxConns, err := strconv.Atoi(v); err == nil {
			cfg.Database.MaxConns = maxConns
		}
	}

	// Redis
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Redis.Port = port
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			cfg.Redis.DB = db
		}
	}

	// Logger
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logger.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.Logger.Format = v
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	return nil
}

// DSN returns the PostgreSQL connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode, c.Timezone,
	)
}

// RedisAddr returns the Redis address.
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDevelopment returns true if the environment is development.
func (c *ServerConfig) IsDevelopment() bool {
	return c.Environment == "dev" || c.Environment == "development"
}

// IsProduction returns true if the environment is production.
func (c *ServerConfig) IsProduction() bool {
	return c.Environment == "prod" || c.Environment == "production"
}
