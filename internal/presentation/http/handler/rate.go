package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/tyokyo320/rateflow/internal/application/query"
	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/pkg/timeutil"
)

// RateHandler handles rate-related HTTP requests.
type RateHandler struct {
	getLatestHandler *query.GetLatestRateHandler
	listRatesHandler *query.ListRatesHandler
	logger           *slog.Logger
}

// NewRateHandler creates a new rate handler.
func NewRateHandler(
	getLatestHandler *query.GetLatestRateHandler,
	listRatesHandler *query.ListRatesHandler,
	logger *slog.Logger,
) *RateHandler {
	return &RateHandler{
		getLatestHandler: getLatestHandler,
		listRatesHandler: listRatesHandler,
		logger:           logger,
	}
}

// GetLatest handles GET /api/rates/latest requests.
// @Summary Get latest exchange rate
// @Description Retrieves the most recent exchange rate for a given currency pair
// @Tags rates
// @Accept json
// @Produce json
// @Param pair query string true "Currency pair (e.g., CNY/JPY, CNYJPY, or CNY-JPY)"
// @Success 200 {object} map[string]interface{} "Success response with rate data"
// @Failure 400 {object} map[string]interface{} "Bad request error"
// @Failure 404 {object} map[string]interface{} "Rate not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/rates/latest [get]
// @Router /api/v1/rates/latest [get]
func (h *RateHandler) GetLatest(c *gin.Context) {
	pairStr := c.Query("pair")
	if pairStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "pair parameter is required",
			},
		})
		return
	}

	pair, err := currency.ParsePair(pairStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid currency pair format",
			},
		})
		return
	}

	result, err := h.getLatestHandler.Handle(c.Request.Context(), query.GetLatestRateQuery{
		Pair: pair,
	})
	if err != nil {
		h.logger.Error("failed to get latest rate", "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "rate not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetByDate handles GET /api/rates requests for a specific date.
// @Summary Get exchange rate for a specific date
// @Description Retrieves the exchange rate for a given currency pair on a specific date
// @Tags rates
// @Accept json
// @Produce json
// @Param pair query string true "Currency pair (e.g., CNY/JPY, CNYJPY, or CNY-JPY)"
// @Param date query string true "Date in YYYY-MM-DD format (e.g., 2025-01-15)"
// @Success 200 {object} map[string]interface{} "Success response with rate data"
// @Failure 400 {object} map[string]interface{} "Bad request error"
// @Failure 404 {object} map[string]interface{} "Rate not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/rates [get]
// @Router /api/v1/rates [get]
func (h *RateHandler) GetByDate(c *gin.Context) {
	pairStr := c.Query("pair")
	dateStr := c.Query("date")

	if pairStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "pair parameter is required",
			},
		})
		return
	}

	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "date parameter is required",
			},
		})
		return
	}

	pair, err := currency.ParsePair(pairStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid currency pair format",
			},
		})
		return
	}

	date, err := timeutil.ParseDate(dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid date format, use YYYY-MM-DD",
			},
		})
		return
	}

	result, err := h.getLatestHandler.Handle(c.Request.Context(), query.GetLatestRateQuery{
		Pair: pair,
	})
	if err != nil || !result.EffectiveDate.Truncate(24*3600000000000).Equal(date.Truncate(24*3600000000000)) {
		h.logger.Error("failed to get rate by date", "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "rate not found for the specified date",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// List handles GET /api/rates/list requests.
// @Summary List exchange rates with pagination
// @Description Retrieves a paginated list of exchange rates, optionally filtered by currency pair
// @Tags rates
// @Accept json
// @Produce json
// @Param pair query string false "Currency pair filter (e.g., CNY/JPY, CNYJPY, or CNY-JPY)"
// @Param page query int false "Page number (default: 1)" default(1)
// @Param pageSize query int false "Items per page (default: 20, max: 100)" default(20)
// @Success 200 {object} map[string]interface{} "Success response with paginated rate list and metadata"
// @Failure 400 {object} map[string]interface{} "Bad request error"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/rates/list [get]
// @Router /api/v1/rates/list [get]
func (h *RateHandler) List(c *gin.Context) {
	// Parse query parameters
	pairStr := c.Query("pair")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Parse currency pair if provided
	var pair currency.Pair
	if pairStr != "" {
		pair, err = currency.ParsePair(pairStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "invalid currency pair format",
				},
			})
			return
		}
	}

	// Execute query
	result, err := h.listRatesHandler.Handle(c.Request.Context(), query.ListRatesQuery{
		Pair:     pair,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		h.logger.Error("failed to list rates", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve rates",
			},
		})
		return
	}

	// Return response with metadata
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result.Items,
		"meta": gin.H{
			"page":       result.Pagination.Page,
			"pageSize":   result.Pagination.PageSize,
			"total":      result.Pagination.Total,
			"totalPages": result.Pagination.TotalPages,
		},
	})
}

// Health handles GET /health requests.
// @Summary Health check endpoint
// @Description Returns the health status of the API service
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Health status response"
// @Router /health [get]
func (h *RateHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status": "healthy",
		},
	})
}
