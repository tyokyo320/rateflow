package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Logger returns a middleware that logs HTTP requests.
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status and error
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		// Log with structured fields
		fields := []any{
			slog.String("request_id", requestID),
			slog.String("client_ip", clientIP),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("user_agent", userAgent),
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields = append(fields, slog.String("error", c.Errors.String()))
		}

		// Log based on status code
		if status >= 500 {
			logger.Error("server error", fields...)
		} else if status >= 400 {
			logger.Warn("client error", fields...)
		} else {
			logger.Info("request completed", fields...)
		}
	}
}

// RequestID returns a middleware that ensures each request has a unique ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already set (e.g., from a header)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}
