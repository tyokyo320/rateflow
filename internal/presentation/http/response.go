package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a unified API response format.
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Error represents an error in API responses.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Meta represents metadata in API responses (for pagination, etc.).
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"pageSize,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"totalPages,omitempty"`
}

// SuccessResponse returns a success response.
func SuccessResponse(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessResponseWithMeta returns a success response with metadata.
func SuccessResponseWithMeta(c *gin.Context, data any, meta *Meta) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// CreatedResponse returns a created response.
func CreatedResponse(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// ErrorResponse returns an error response.
func ErrorResponse(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorResponseWithDetails returns an error response with details.
func ErrorResponseWithDetails(c *gin.Context, statusCode int, code, message string, details any) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// BadRequestError returns a 400 bad request error.
func BadRequestError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// UnauthorizedError returns a 401 unauthorized error.
func UnauthorizedError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// ForbiddenError returns a 403 forbidden error.
func ForbiddenError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFoundError returns a 404 not found error.
func NotFoundError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", message)
}

// InternalServerError returns a 500 internal server error.
func InternalServerError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// ValidationError returns a 422 validation error.
func ValidationError(c *gin.Context, details any) {
	ErrorResponseWithDetails(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed", details)
}
