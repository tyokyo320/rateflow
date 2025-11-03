// Package rate provides the exchange rate aggregate root and related domain logic.
package rate

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tyokyo320/rateflow/internal/domain/currency"
)

// Source represents the data source of an exchange rate.
type Source string

const (
	SourceUnionPay     Source = "unionpay"
	SourceECB          Source = "ecb" // European Central Bank
	SourceOpenExchange Source = "openexchange"
	SourceManual       Source = "manual"
)

// Rate represents an exchange rate aggregate root.
// This is the core domain entity that encapsulates exchange rate business logic.
type Rate struct {
	id            string
	pair          currency.Pair
	value         float64
	effectiveDate time.Time
	source        Source
	createdAt     time.Time
	updatedAt     time.Time
}

// NewRate creates a new Rate with validation.
func NewRate(
	pair currency.Pair,
	value float64,
	effectiveDate time.Time,
	source Source,
) (*Rate, error) {
	rate := &Rate{
		id:            uuid.New().String(),
		pair:          pair,
		value:         value,
		effectiveDate: effectiveDate,
		source:        source,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
	}

	if err := rate.Validate(); err != nil {
		return nil, err
	}

	return rate, nil
}

// Reconstitute creates a Rate from persisted data (used by repository).
func Reconstitute(
	id string,
	pair currency.Pair,
	value float64,
	effectiveDate time.Time,
	source Source,
	createdAt, updatedAt time.Time,
) *Rate {
	return &Rate{
		id:            id,
		pair:          pair,
		value:         value,
		effectiveDate: effectiveDate,
		source:        source,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

// Validate performs domain validation on the rate.
func (r *Rate) Validate() error {
	if r.value <= 0 {
		return ErrInvalidRate{reason: "rate value must be positive"}
	}

	if r.effectiveDate.After(time.Now().Add(24 * time.Hour)) {
		return ErrInvalidRate{reason: "effective date cannot be more than 1 day in the future"}
	}

	if r.effectiveDate.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return ErrInvalidRate{reason: "effective date is too far in the past"}
	}

	if !r.isValidSource() {
		return ErrInvalidRate{reason: fmt.Sprintf("invalid source: %s", r.source)}
	}

	return nil
}

func (r *Rate) isValidSource() bool {
	validSources := []Source{SourceUnionPay, SourceECB, SourceOpenExchange, SourceManual}
	for _, valid := range validSources {
		if r.source == valid {
			return true
		}
	}
	return false
}

// UpdateValue updates the exchange rate value.
func (r *Rate) UpdateValue(newValue float64) error {
	if newValue <= 0 {
		return ErrInvalidRate{reason: "rate value must be positive"}
	}

	r.value = newValue
	r.updatedAt = time.Now()

	return nil
}

// IsStale checks if the rate data is considered stale.
func (r *Rate) IsStale(threshold time.Duration) bool {
	return time.Since(r.updatedAt) > threshold
}

// IsEffectiveOn checks if the rate is effective on the given date.
func (r *Rate) IsEffectiveOn(date time.Time) bool {
	effectiveDay := r.effectiveDate.Truncate(24 * time.Hour)
	checkDay := date.Truncate(24 * time.Hour)
	return effectiveDay.Equal(checkDay)
}

// Convert converts an amount using this exchange rate.
// For example, if rate is CNY/JPY = 20, then Convert(100) returns 2000 JPY.
func (r *Rate) Convert(amount float64) float64 {
	return amount * r.value
}

// ConvertInverse converts an amount using the inverse rate.
func (r *Rate) ConvertInverse(amount float64) float64 {
	if r.value == 0 {
		return 0
	}
	return amount / r.value
}

// GetID implements the genericrepo.Entity interface.
func (r *Rate) GetID() string {
	return r.id
}

// Getters
func (r *Rate) ID() string               { return r.id }
func (r *Rate) Pair() currency.Pair      { return r.pair }
func (r *Rate) Value() float64           { return r.value }
func (r *Rate) EffectiveDate() time.Time { return r.effectiveDate }
func (r *Rate) Source() Source           { return r.source }
func (r *Rate) CreatedAt() time.Time     { return r.createdAt }
func (r *Rate) UpdatedAt() time.Time     { return r.updatedAt }
