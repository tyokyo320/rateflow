package query

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tyokyo320/rateflow/internal/application/dto"
	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/domain/rate"
	"github.com/tyokyo320/rateflow/internal/infrastructure/persistence/redis"
)

// GetLatestRateQuery represents a query for the latest exchange rate.
type GetLatestRateQuery struct {
	Pair currency.Pair
}

// GetLatestRateHandler handles getting the latest exchange rate.
type GetLatestRateHandler struct {
	rateRepo rate.Repository
	cache    redis.CacheInterface
	logger   *slog.Logger
}

// NewGetLatestRateHandler creates a new handler.
func NewGetLatestRateHandler(
	rateRepo rate.Repository,
	cache redis.CacheInterface,
	logger *slog.Logger,
) *GetLatestRateHandler {
	return &GetLatestRateHandler{
		rateRepo: rateRepo,
		cache:    cache,
		logger:   logger,
	}
}

// Handle executes the query.
func (h *GetLatestRateHandler) Handle(ctx context.Context, query GetLatestRateQuery) (*dto.RateResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("latest:%s", query.Pair.String())
	var cached dto.RateResponse

	if err := h.cache.Get(ctx, cacheKey, &cached); err == nil {
		h.logger.Debug("cache hit", "key", cacheKey)
		return &cached, nil
	}

	// Cache miss - query database
	h.logger.Debug("cache miss", "key", cacheKey)

	r, err := h.rateRepo.FindLatest(ctx, query.Pair)
	if err != nil {
		h.logger.Error("failed to find latest rate",
			"error", err,
			"pair", query.Pair.String(),
		)
		return nil, err
	}

	// Convert to DTO
	result := h.toDTO(r)

	// Cache the result
	if err := h.cache.Set(ctx, cacheKey, result, 5*time.Minute); err != nil {
		h.logger.Warn("failed to cache result", "error", err)
	}

	return result, nil
}

func (h *GetLatestRateHandler) toDTO(r *rate.Rate) *dto.RateResponse {
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
