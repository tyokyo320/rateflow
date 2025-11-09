package query

import (
	"context"
	"log/slog"
	"time"

	"github.com/tyokyo320/rateflow/internal/application/dto"
	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/domain/rate"
	"github.com/tyokyo320/rateflow/pkg/genericrepo"
)

// ListRatesQuery represents a query for listing rates with pagination.
type ListRatesQuery struct {
	Pair      currency.Pair
	Page      int
	PageSize  int
	StartDate *time.Time
	EndDate   *time.Time
}

// ListRatesResult contains the paginated list of rates.
type ListRatesResult struct {
	Items      []*dto.RateResponse    `json:"items"`
	Pagination genericrepo.Pagination `json:"pagination"`
}

// ListRatesHandler handles listing exchange rates.
type ListRatesHandler struct {
	rateRepo rate.Repository
	logger   *slog.Logger
}

// NewListRatesHandler creates a new handler.
func NewListRatesHandler(
	rateRepo rate.Repository,
	logger *slog.Logger,
) *ListRatesHandler {
	return &ListRatesHandler{
		rateRepo: rateRepo,
		logger:   logger,
	}
}

// Handle executes the query.
func (h *ListRatesHandler) Handle(ctx context.Context, query ListRatesQuery) (*ListRatesResult, error) {
	// If date range is specified, use FindByDateRange instead of generic query
	if query.StartDate != nil && query.EndDate != nil {
		return h.handleDateRangeQuery(ctx, query)
	}

	// Build query options
	opts := []genericrepo.QueryOption{
		genericrepo.WithFilter("base_currency", query.Pair.Base().String()),
		genericrepo.WithFilter("quote_currency", query.Pair.Quote().String()),
		genericrepo.WithOrderBy("effective_date DESC"),
		genericrepo.WithPagination(query.Page, query.PageSize),
	}

	// Get rates - try direct query first
	rates, err := h.rateRepo.FindAll(ctx, opts...)
	needsInversion := false
	directCount := int64(0)

	// Count direct results
	if err == nil && len(rates) > 0 {
		countOpts := []genericrepo.QueryOption{
			genericrepo.WithFilter("base_currency", query.Pair.Base().String()),
			genericrepo.WithFilter("quote_currency", query.Pair.Quote().String()),
		}
		directCount, _ = h.rateRepo.Count(ctx, countOpts...)
	}

	// Try inverse pair if: 1) error occurred, 2) no results, OR 3) very few results (< 10)
	// This handles cases where one direction has much more data than the other
	if err != nil || len(rates) == 0 || directCount < 10 {
		h.logger.Debug("trying inverse pair for list",
			"original_pair", query.Pair.String(),
			"inverse_pair", query.Pair.Inverse().String(),
			"direct_count", directCount,
		)

		inversePair := query.Pair.Inverse()
		inverseOpts := []genericrepo.QueryOption{
			genericrepo.WithFilter("base_currency", inversePair.Base().String()),
			genericrepo.WithFilter("quote_currency", inversePair.Quote().String()),
			genericrepo.WithOrderBy("effective_date DESC"),
			genericrepo.WithPagination(query.Page, query.PageSize),
		}

		inverseRates, inverseErr := h.rateRepo.FindAll(ctx, inverseOpts...)

		// Count inverse results
		inverseCount := int64(0)
		if inverseErr == nil && len(inverseRates) > 0 {
			inverseCountOpts := []genericrepo.QueryOption{
				genericrepo.WithFilter("base_currency", inversePair.Base().String()),
				genericrepo.WithFilter("quote_currency", inversePair.Quote().String()),
			}
			inverseCount, _ = h.rateRepo.Count(ctx, inverseCountOpts...)
		}

		// Use inverse data if it has more records
		if inverseErr == nil && inverseCount > directCount {
			h.logger.Debug("using inverse pair data",
				"direct_count", directCount,
				"inverse_count", inverseCount,
			)
			rates = inverseRates
			needsInversion = true
		} else if err != nil {
			// If direct query failed and inverse also failed, return error
			if inverseErr != nil {
				h.logger.Error("failed to list rates for both directions",
					"error", err,
					"inverse_error", inverseErr,
					"pair", query.Pair.String(),
				)
				return nil, err
			}
			// Direct failed but inverse succeeded
			rates = inverseRates
			needsInversion = true
		}
	}

	// Get total count (use inverse if needed)
	var total int64
	if needsInversion {
		inversePair := query.Pair.Inverse()
		countOpts := []genericrepo.QueryOption{
			genericrepo.WithFilter("base_currency", inversePair.Base().String()),
			genericrepo.WithFilter("quote_currency", inversePair.Quote().String()),
		}
		total, err = h.rateRepo.Count(ctx, countOpts...)
	} else {
		countOpts := []genericrepo.QueryOption{
			genericrepo.WithFilter("base_currency", query.Pair.Base().String()),
			genericrepo.WithFilter("quote_currency", query.Pair.Quote().String()),
		}
		total, err = h.rateRepo.Count(ctx, countOpts...)
	}

	if err != nil {
		h.logger.Error("failed to count rates", "error", err)
		return nil, err
	}

	// Convert to DTOs
	items := make([]*dto.RateResponse, 0, len(rates))
	for _, r := range rates {
		if needsInversion {
			items = append(items, h.toDTOInverted(r, query.Pair))
		} else {
			items = append(items, h.toDTO(r))
		}
	}

	// Build pagination info
	result := &ListRatesResult{
		Items: items,
		Pagination: genericrepo.Pagination{
			Page:     query.Page,
			PageSize: query.PageSize,
			Total:    total,
		},
	}
	result.Pagination.CalculateTotalPages()

	return result, nil
}

func (h *ListRatesHandler) toDTO(r *rate.Rate) *dto.RateResponse {
	return &dto.RateResponse{
		ID:            r.ID(),
		Pair:          r.Pair().String(),
		BaseCurrency:  r.Pair().Base().String(),
		QuoteCurrency: r.Pair().Quote().String(),
		Rate:          r.Value(),
		EffectiveDate: r.EffectiveDate(),
		Source:        string(r.Source()),
		CreatedAt:     r.CreatedAt(),
		UpdatedAt:     r.UpdatedAt(),
	}
}

// toDTOInverted converts a rate from inverse pair to the requested pair.
func (h *ListRatesHandler) toDTOInverted(r *rate.Rate, requestedPair currency.Pair) *dto.RateResponse {
	invertedRate := r.Pair().ConvertRate(r.Value())

	return &dto.RateResponse{
		ID:            r.ID(),
		Pair:          requestedPair.String(),
		BaseCurrency:  requestedPair.Base().String(),
		QuoteCurrency: requestedPair.Quote().String(),
		Rate:          invertedRate,
		EffectiveDate: r.EffectiveDate(),
		Source:        string(r.Source()),
		CreatedAt:     r.CreatedAt(),
		UpdatedAt:     r.UpdatedAt(),
	}
}

// handleDateRangeQuery handles queries with specific date ranges.
func (h *ListRatesHandler) handleDateRangeQuery(ctx context.Context, query ListRatesQuery) (*ListRatesResult, error) {
	// Try direct pair first
	rates, err := h.rateRepo.FindByDateRange(ctx, query.Pair, *query.StartDate, *query.EndDate)
	needsInversion := false

	// Count direct results
	directCount := int64(len(rates))

	// Try inverse pair if no results or very few results
	if err != nil || len(rates) == 0 || directCount < 10 {
		h.logger.Debug("trying inverse pair for date range query",
			"original_pair", query.Pair.String(),
			"inverse_pair", query.Pair.Inverse().String(),
			"direct_count", directCount,
		)

		inversePair := query.Pair.Inverse()
		inverseRates, inverseErr := h.rateRepo.FindByDateRange(ctx, inversePair, *query.StartDate, *query.EndDate)
		inverseCount := int64(len(inverseRates))

		// Use inverse data if it has more records
		if inverseErr == nil && inverseCount > directCount {
			h.logger.Debug("using inverse pair data for date range",
				"direct_count", directCount,
				"inverse_count", inverseCount,
			)
			rates = inverseRates
			needsInversion = true
		} else if err != nil {
			// If direct query failed and inverse also failed, return error
			if inverseErr != nil {
				h.logger.Error("failed to query date range for both directions",
					"error", err,
					"inverse_error", inverseErr,
					"pair", query.Pair.String(),
				)
				return nil, err
			}
			// Direct failed but inverse succeeded
			rates = inverseRates
			needsInversion = true
		}
	}

	// Sort by date descending (most recent first)
	// Since FindByDateRange returns in ascending order
	for i, j := 0, len(rates)-1; i < j; i, j = i+1, j-1 {
		rates[i], rates[j] = rates[j], rates[i]
	}

	// Apply pagination manually
	total := int64(len(rates))
	startIdx := (query.Page - 1) * query.PageSize
	endIdx := startIdx + query.PageSize

	if startIdx >= len(rates) {
		rates = []*rate.Rate{}
	} else {
		if endIdx > len(rates) {
			endIdx = len(rates)
		}
		rates = rates[startIdx:endIdx]
	}

	// Convert to DTOs
	items := make([]*dto.RateResponse, 0, len(rates))
	for _, r := range rates {
		if needsInversion {
			items = append(items, h.toDTOInverted(r, query.Pair))
		} else {
			items = append(items, h.toDTO(r))
		}
	}

	// Build pagination info
	result := &ListRatesResult{
		Items: items,
		Pagination: genericrepo.Pagination{
			Page:     query.Page,
			PageSize: query.PageSize,
			Total:    total,
		},
	}
	result.Pagination.CalculateTotalPages()

	return result, nil
}
