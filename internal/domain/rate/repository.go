// Package rate defines the repository interface for the rate aggregate.
package rate

import (
	"context"
	"time"

	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/pkg/genericrepo"
)

// Repository defines the persistence interface for Rate entities.
// This follows the repository pattern from DDD.
type Repository interface {
	// Embed the generic repository interface
	genericrepo.Repository[*Rate]

	// Domain-specific query methods

	// FindByPairAndDate finds a rate for a specific currency pair and date.
	FindByPairAndDate(ctx context.Context, pair currency.Pair, date time.Time) (*Rate, error)

	// FindLatest finds the most recent rate for a currency pair.
	FindLatest(ctx context.Context, pair currency.Pair) (*Rate, error)

	// FindByDateRange finds rates for a currency pair within a date range.
	FindByDateRange(ctx context.Context, pair currency.Pair, start, end time.Time) ([]*Rate, error)

	// FindByPairs finds the latest rates for multiple currency pairs.
	FindByPairs(ctx context.Context, pairs []currency.Pair) ([]*Rate, error)

	// ExistsByPairAndDate checks if a rate exists for a specific pair and date.
	ExistsByPairAndDate(ctx context.Context, pair currency.Pair, date time.Time) (bool, error)

	// DeleteOlderThan deletes rates older than the specified date.
	DeleteOlderThan(ctx context.Context, date time.Time) (int64, error)
}
