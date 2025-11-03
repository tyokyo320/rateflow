// Package currency provides currency-related value objects.
package currency

import (
	"fmt"
	"strings"
)

// Code represents a currency code (ISO 4217).
type Code string

// Supported currency codes
const (
	CNY Code = "CNY" // Chinese Yuan
	JPY Code = "JPY" // Japanese Yen
	USD Code = "USD" // US Dollar
	EUR Code = "EUR" // Euro
	GBP Code = "GBP" // British Pound
	HKD Code = "HKD" // Hong Kong Dollar
	KRW Code = "KRW" // South Korean Won
	SGD Code = "SGD" // Singapore Dollar
)

// validCodes contains all supported currency codes.
var validCodes = map[Code]bool{
	CNY: true,
	JPY: true,
	USD: true,
	EUR: true,
	GBP: true,
	HKD: true,
	KRW: true,
	SGD: true,
}

// NewCode creates a new Code from a string.
func NewCode(s string) (Code, error) {
	code := Code(strings.ToUpper(strings.TrimSpace(s)))
	if !code.IsValid() {
		return "", fmt.Errorf("invalid currency code: %s", s)
	}
	return code, nil
}

// IsValid checks if the currency code is valid.
func (c Code) IsValid() bool {
	return validCodes[c]
}

// String returns the string representation of the currency code.
func (c Code) String() string {
	return string(c)
}

// Equal checks if two currency codes are equal.
func (c Code) Equal(other Code) bool {
	return c == other
}

// AllCodes returns all supported currency codes.
func AllCodes() []Code {
	codes := make([]Code, 0, len(validCodes))
	for code := range validCodes {
		codes = append(codes, code)
	}
	return codes
}

// IsValidString checks if a string is a valid currency code.
func IsValidString(s string) bool {
	code := Code(strings.ToUpper(strings.TrimSpace(s)))
	return code.IsValid()
}
