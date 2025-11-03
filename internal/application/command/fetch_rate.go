package command

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/domain/provider"
	"github.com/tyokyo320/rateflow/internal/domain/rate"
	"github.com/tyokyo320/rateflow/internal/infrastructure/persistence/redis"
)

// FetchRateCommand represents a command to fetch and store an exchange rate.
type FetchRateCommand struct {
	Pair currency.Pair
	Date time.Time
}

// FetchRateHandler handles the fetch rate command.
type FetchRateHandler struct {
	rateRepo rate.Repository
	provider provider.Provider
	cache    *redis.Cache
	logger   *slog.Logger
}

// NewFetchRateHandler creates a new fetch rate command handler.
func NewFetchRateHandler(
	rateRepo rate.Repository,
	provider provider.Provider,
	cache *redis.Cache,
	logger *slog.Logger,
) *FetchRateHandler {
	return &FetchRateHandler{
		rateRepo: rateRepo,
		provider: provider,
		cache:    cache,
		logger:   logger,
	}
}

// Handle executes the fetch rate command.
func (h *FetchRateHandler) Handle(ctx context.Context, cmd FetchRateCommand) error {
	h.logger.Info("fetching rate",
		"pair", cmd.Pair.String(),
		"date", cmd.Date.Format("2006-01-02"),
		"provider", h.provider.Name(),
	)

	// Check if rate already exists
	exists, err := h.rateRepo.ExistsByPairAndDate(ctx, cmd.Pair, cmd.Date)
	if err != nil {
		h.logger.Error("failed to check if rate exists", "error", err)
		return fmt.Errorf("check rate existence: %w", err)
	}

	if exists {
		h.logger.Info("rate already exists, skipping",
			"pair", cmd.Pair.String(),
			"date", cmd.Date.Format("2006-01-02"),
		)
		return nil
	}

	// Fetch rate from provider
	rateValue, err := h.provider.FetchRate(ctx, cmd.Pair, cmd.Date)
	if err != nil {
		h.logger.Error("failed to fetch rate from provider",
			"error", err,
			"pair", cmd.Pair.String(),
			"date", cmd.Date.Format("2006-01-02"),
		)
		return fmt.Errorf("fetch rate from provider: %w", err)
	}

	// Create rate entity
	r, err := rate.NewRate(
		cmd.Pair,
		rateValue,
		cmd.Date,
		rate.Source(h.provider.Name()),
	)
	if err != nil {
		h.logger.Error("failed to create rate entity", "error", err)
		return fmt.Errorf("create rate entity: %w", err)
	}

	// Save to repository
	if err := h.rateRepo.Create(ctx, r); err != nil {
		h.logger.Error("failed to save rate", "error", err)
		return fmt.Errorf("save rate: %w", err)
	}

	h.logger.Info("rate fetched and saved successfully",
		"id", r.ID(),
		"pair", r.Pair().String(),
		"rate", r.Value(),
		"date", r.EffectiveDate().Format("2006-01-02"),
	)

	// Invalidate cache for this pair
	cacheKey := fmt.Sprintf("latest:%s", cmd.Pair.String())
	if err := h.cache.Delete(ctx, cacheKey); err != nil {
		h.logger.Warn("failed to invalidate cache", "error", err, "key", cacheKey)
	}

	return nil
}

// FetchRateResult contains the result of fetching a rate.
type FetchRateResult struct {
	RateID string
	Pair   string
	Value  float64
	Date   time.Time
}
