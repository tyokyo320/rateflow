package dto

import "time"

// RateResponse represents a rate in API responses.
type RateResponse struct {
	ID            string    `json:"id"`
	Pair          string    `json:"pair"`
	BaseCurrency  string    `json:"baseCurrency"`
	QuoteCurrency string    `json:"quoteCurrency"`
	Rate          float64   `json:"rate"`
	EffectiveDate time.Time `json:"effectiveDate"`
	Source        string    `json:"source"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// RateRequest represents a request for getting a specific rate.
type RateRequest struct {
	Pair string `json:"pair" binding:"required"`
	Date string `json:"date"` // Optional, format: YYYY-MM-DD
}

// ListRatesRequest represents a request for listing rates.
type ListRatesRequest struct {
	Pair     string `json:"pair"`
	Page     int    `json:"page" binding:"gte=1"`
	PageSize int    `json:"pageSize" binding:"gte=1,lte=100"`
}

// HistoryRequest represents a request for historical rates.
type HistoryRequest struct {
	Pair      string `json:"pair" binding:"required"`
	StartDate string `json:"startDate" binding:"required"`
	EndDate   string `json:"endDate" binding:"required"`
}
