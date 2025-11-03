// Package currency provides currency pair value objects.
package currency

import (
	"fmt"
	"strings"
)

// Pair represents a currency pair (e.g., CNY/JPY).
// Base currency is what you're converting from.
// Quote currency is what you're converting to.
// For example, in CNY/JPY, 1 CNY = X JPY.
type Pair struct {
	base  Code
	quote Code
}

// NewPair creates a new currency pair.
func NewPair(base, quote Code) (Pair, error) {
	if !base.IsValid() {
		return Pair{}, fmt.Errorf("invalid base currency: %s", base)
	}

	if !quote.IsValid() {
		return Pair{}, fmt.Errorf("invalid quote currency: %s", quote)
	}

	if base == quote {
		return Pair{}, fmt.Errorf("base and quote currencies must be different")
	}

	return Pair{base: base, quote: quote}, nil
}

// MustNewPair creates a new currency pair or panics on error.
// Use this only when you're certain the inputs are valid.
func MustNewPair(base, quote Code) Pair {
	pair, err := NewPair(base, quote)
	if err != nil {
		panic(err)
	}
	return pair
}

// ParsePair parses a currency pair string (e.g., "CNY/JPY", "CNY-JPY", "CNYJPY").
func ParsePair(s string) (Pair, error) {
	s = strings.TrimSpace(strings.ToUpper(s))

	// Try different separators
	var base, quote string

	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		if len(parts) != 2 {
			return Pair{}, fmt.Errorf("invalid pair format: %s", s)
		}
		base, quote = parts[0], parts[1]
	} else if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) != 2 {
			return Pair{}, fmt.Errorf("invalid pair format: %s", s)
		}
		base, quote = parts[0], parts[1]
	} else if len(s) == 6 {
		// Assume format like "CNYJPY"
		base, quote = s[:3], s[3:]
	} else {
		return Pair{}, fmt.Errorf("invalid pair format: %s", s)
	}

	baseCode, err := NewCode(base)
	if err != nil {
		return Pair{}, err
	}

	quoteCode, err := NewCode(quote)
	if err != nil {
		return Pair{}, err
	}

	return NewPair(baseCode, quoteCode)
}

// Base returns the base currency.
func (p Pair) Base() Code {
	return p.base
}

// Quote returns the quote currency.
func (p Pair) Quote() Code {
	return p.quote
}

// String returns the string representation (e.g., "CNY/JPY").
func (p Pair) String() string {
	return fmt.Sprintf("%s/%s", p.base, p.quote)
}

// Compact returns a compact string representation without separator (e.g., "CNYJPY").
func (p Pair) Compact() string {
	return string(p.base) + string(p.quote)
}

// Inverse returns the inverse pair (e.g., JPY/CNY for CNY/JPY).
func (p Pair) Inverse() Pair {
	return Pair{base: p.quote, quote: p.base}
}

// Equal checks if two pairs are equal.
func (p Pair) Equal(other Pair) bool {
	return p.base == other.base && p.quote == other.quote
}

// ConvertRate converts a rate from this pair to its inverse.
// For example, if CNY/JPY = 20, then JPY/CNY = 1/20 = 0.05.
func (p Pair) ConvertRate(rate float64) float64 {
	if rate == 0 {
		return 0
	}
	return 1.0 / rate
}

// CommonPairs returns commonly used currency pairs.
func CommonPairs() []Pair {
	return []Pair{
		MustNewPair(CNY, JPY), // Chinese Yuan to Japanese Yen
		MustNewPair(USD, JPY), // US Dollar to Japanese Yen
		MustNewPair(EUR, JPY), // Euro to Japanese Yen
		MustNewPair(USD, CNY), // US Dollar to Chinese Yuan
		MustNewPair(EUR, USD), // Euro to US Dollar
		MustNewPair(GBP, USD), // British Pound to US Dollar
	}
}
