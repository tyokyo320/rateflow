package query

import (
	"context"
	"log/slog"

	"github.com/tyokyo320/rateflow/internal/application/dto"
	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/domain/rate"
	"github.com/tyokyo320/rateflow/pkg/genericrepo"
)

// ListRatesQuery represents a query for listing rates with pagination.
type ListRatesQuery struct {
	Pair     currency.Pair
	Page     int
	PageSize int
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
	// Build query options
	opts := []genericrepo.QueryOption{
		genericrepo.WithFilter("base_currency", query.Pair.Base().String()),
		genericrepo.WithFilter("quote_currency", query.Pair.Quote().String()),
		genericrepo.WithOrderBy("effective_date DESC"),
		genericrepo.WithPagination(query.Page, query.PageSize),
	}

	// Get rates
	rates, err := h.rateRepo.FindAll(ctx, opts...)
	if err != nil {
		h.logger.Error("failed to list rates",
			"error", err,
			"pair", query.Pair.String(),
		)
		return nil, err
	}

	// Get total count
	countOpts := []genericrepo.QueryOption{
		genericrepo.WithFilter("base_currency", query.Pair.Base().String()),
		genericrepo.WithFilter("quote_currency", query.Pair.Quote().String()),
	}
	total, err := h.rateRepo.Count(ctx, countOpts...)
	if err != nil {
		h.logger.Error("failed to count rates", "error", err)
		return nil, err
	}

	// Convert to DTOs
	items := make([]*dto.RateResponse, 0, len(rates))
	for _, r := range rates {
		items = append(items, h.toDTO(r))
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
