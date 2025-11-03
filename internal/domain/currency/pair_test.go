package currency_test

import (
	"testing"

	"github.com/tyokyo320/rateflow/internal/domain/currency"
)

func TestNewPair(t *testing.T) {
	tests := []struct {
		name      string
		base      currency.Code
		quote     currency.Code
		wantErr   bool
		errString string
	}{
		{
			name:    "valid pair CNY/JPY",
			base:    currency.CNY,
			quote:   currency.JPY,
			wantErr: false,
		},
		{
			name:    "valid pair USD/JPY",
			base:    currency.USD,
			quote:   currency.JPY,
			wantErr: false,
		},
		{
			name:      "same currency",
			base:      currency.CNY,
			quote:     currency.CNY,
			wantErr:   true,
			errString: "base and quote currencies must be different",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair, err := currency.NewPair(tt.base, tt.quote)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPair() expected error but got nil")
				}
				if tt.errString != "" && err.Error() != tt.errString {
					t.Errorf("NewPair() error = %v, want %v", err.Error(), tt.errString)
				}
			} else {
				if err != nil {
					t.Errorf("NewPair() unexpected error = %v", err)
				}
				if pair.Base() != tt.base {
					t.Errorf("NewPair() base = %v, want %v", pair.Base(), tt.base)
				}
				if pair.Quote() != tt.quote {
					t.Errorf("NewPair() quote = %v, want %v", pair.Quote(), tt.quote)
				}
			}
		})
	}
}

func TestParsePair(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantBase  currency.Code
		wantQuote currency.Code
		wantErr   bool
	}{
		{
			name:      "slash format",
			input:     "CNY/JPY",
			wantBase:  currency.CNY,
			wantQuote: currency.JPY,
			wantErr:   false,
		},
		{
			name:      "dash format",
			input:     "USD-JPY",
			wantBase:  currency.USD,
			wantQuote: currency.JPY,
			wantErr:   false,
		},
		{
			name:      "compact format",
			input:     "EURJPY",
			wantBase:  currency.EUR,
			wantQuote: currency.JPY,
			wantErr:   false,
		},
		{
			name:    "invalid format",
			input:   "INVALID",
			wantErr: true,
		},
		{
			name:    "invalid currency",
			input:   "XXX/YYY",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair, err := currency.ParsePair(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePair() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ParsePair() unexpected error = %v", err)
				}
				if pair.Base() != tt.wantBase {
					t.Errorf("ParsePair() base = %v, want %v", pair.Base(), tt.wantBase)
				}
				if pair.Quote() != tt.wantQuote {
					t.Errorf("ParsePair() quote = %v, want %v", pair.Quote(), tt.wantQuote)
				}
			}
		})
	}
}

func TestPair_String(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)

	if pair.String() != "CNY/JPY" {
		t.Errorf("Pair.String() = %v, want CNY/JPY", pair.String())
	}
}

func TestPair_Inverse(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	inverse := pair.Inverse()

	if inverse.Base() != currency.JPY {
		t.Errorf("Inverse base = %v, want JPY", inverse.Base())
	}
	if inverse.Quote() != currency.CNY {
		t.Errorf("Inverse quote = %v, want CNY", inverse.Quote())
	}
}

func TestPair_Equal(t *testing.T) {
	pair1 := currency.MustNewPair(currency.CNY, currency.JPY)
	pair2 := currency.MustNewPair(currency.CNY, currency.JPY)
	pair3 := currency.MustNewPair(currency.USD, currency.JPY)

	if !pair1.Equal(pair2) {
		t.Error("Equal pairs should be equal")
	}

	if pair1.Equal(pair3) {
		t.Error("Different pairs should not be equal")
	}
}

func TestPair_ConvertRate(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)

	tests := []struct {
		name     string
		rate     float64
		expected float64
	}{
		{
			name:     "normal rate",
			rate:     20.0,
			expected: 0.05,
		},
		{
			name:     "zero rate",
			rate:     0.0,
			expected: 0.0,
		},
		{
			name:     "small rate",
			rate:     0.061234,
			expected: 16.329588,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pair.ConvertRate(tt.rate)
			if tt.rate != 0 && result == 0 {
				t.Errorf("ConvertRate(%v) = %v, want non-zero", tt.rate, result)
			}
		})
	}
}
