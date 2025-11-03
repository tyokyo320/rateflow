package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID if available
				requestID, _ := c.Get("request_id")

				// Log the panic with stack trace
				logger.Error("panic recovered",
					slog.Any("error", err),
					slog.String("request_id", fmt.Sprintf("%v", requestID)),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("stack", string(debug.Stack())),
				)

				// Return error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"code":  "INTERNAL_ERROR",
				})

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}
