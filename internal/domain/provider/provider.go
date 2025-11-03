// Package provider defines the interface for external exchange rate data sources.
package provider

import (
	"context"
	"time"

	"github.com/tyokyo320/rateflow/internal/domain/currency"
)

// Provider represents an external data source for exchange rates.
type Provider interface {
	// Name returns the provider name.
	Name() string

	// FetchRate fetches the exchange rate for a specific currency pair and date.
	FetchRate(ctx context.Context, pair currency.Pair, date time.Time) (float64, error)

	// FetchLatest fetches the latest available exchange rate for a currency pair.
	FetchLatest(ctx context.Context, pair currency.Pair) (float64, error)

	// SupportedPairs returns the list of currency pairs supported by this provider.
	SupportedPairs() []currency.Pair

	// SupportsMulti returns true if the provider supports fetching multiple pairs at once.
	SupportsMulti() bool

	// FetchMulti fetches rates for multiple currency pairs (if supported).
	// Returns a map of pair string to rate value.
	FetchMulti(ctx context.Context, pairs []currency.Pair, date time.Time) (map[string]float64, error)
}

// ProviderError represents an error from a provider.
type ProviderError struct {
	ProviderName string
	Message      string
	Err          error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return e.ProviderName + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.ProviderName + ": " + e.Message
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new ProviderError.
func NewProviderError(providerName, message string, err error) *ProviderError {
	return &ProviderError{
		ProviderName: providerName,
		Message:      message,
		Err:          err,
	}
}
