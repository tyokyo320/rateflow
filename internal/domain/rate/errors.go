// Package rate provides domain errors for the rate aggregate.
package rate

import "fmt"

// ErrInvalidRate represents a domain validation error for rates.
type ErrInvalidRate struct {
	reason string
}

func (e ErrInvalidRate) Error() string {
	return fmt.Sprintf("invalid rate: %s", e.reason)
}

// ErrRateNotFound indicates that a rate was not found.
type ErrRateNotFound struct {
	ID string
}

func (e ErrRateNotFound) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("rate not found: %s", e.ID)
	}
	return "rate not found"
}

// ErrDuplicateRate indicates that a rate already exists for the given criteria.
type ErrDuplicateRate struct {
	Pair string
	Date string
}

func (e ErrDuplicateRate) Error() string {
	return fmt.Sprintf("rate already exists for %s on %s", e.Pair, e.Date)
}

// ErrStaleRate indicates that a rate is too old.
type ErrStaleRate struct {
	Age string
}

func (e ErrStaleRate) Error() string {
	return fmt.Sprintf("rate is stale: %s old", e.Age)
}
