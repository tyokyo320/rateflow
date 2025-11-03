package rate_test

import (
	"testing"
	"time"

	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/domain/rate"
)

func TestNewRate(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	now := time.Now()

	tests := []struct {
		name    string
		pair    currency.Pair
		value   float64
		date    time.Time
		source  rate.Source
		wantErr bool
	}{
		{
			name:    "valid rate",
			pair:    pair,
			value:   0.061234,
			date:    now,
			source:  rate.SourceUnionPay,
			wantErr: false,
		},
		{
			name:    "zero value",
			pair:    pair,
			value:   0,
			date:    now,
			source:  rate.SourceUnionPay,
			wantErr: true,
		},
		{
			name:    "negative value",
			pair:    pair,
			value:   -0.01,
			date:    now,
			source:  rate.SourceUnionPay,
			wantErr: true,
		},
		{
			name:    "future date (more than 1 day)",
			pair:    pair,
			value:   0.061234,
			date:    now.Add(48 * time.Hour),
			source:  rate.SourceUnionPay,
			wantErr: true,
		},
		{
			name:    "very old date",
			pair:    pair,
			value:   0.061234,
			date:    time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC),
			source:  rate.SourceUnionPay,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := rate.NewRate(tt.pair, tt.value, tt.date, tt.source)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewRate() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("NewRate() unexpected error = %v", err)
				}
				if r == nil {
					t.Fatal("NewRate() returned nil rate")
				}
				if r.Value() != tt.value {
					t.Errorf("Rate.Value() = %v, want %v", r.Value(), tt.value)
				}
				if !r.Pair().Equal(tt.pair) {
					t.Errorf("Rate.Pair() = %v, want %v", r.Pair(), tt.pair)
				}
				if r.Source() != tt.source {
					t.Errorf("Rate.Source() = %v, want %v", r.Source(), tt.source)
				}
			}
		})
	}
}

func TestRate_UpdateValue(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	r, _ := rate.NewRate(pair, 0.061234, time.Now(), rate.SourceUnionPay)

	tests := []struct {
		name     string
		newValue float64
		wantErr  bool
	}{
		{
			name:     "valid update",
			newValue: 0.062000,
			wantErr:  false,
		},
		{
			name:     "zero value",
			newValue: 0,
			wantErr:  true,
		},
		{
			name:     "negative value",
			newValue: -0.01,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.UpdateValue(tt.newValue)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateValue() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("UpdateValue() unexpected error = %v", err)
				}
				if r.Value() != tt.newValue {
					t.Errorf("Rate.Value() = %v, want %v", r.Value(), tt.newValue)
				}
			}
		})
	}
}

func TestRate_IsStale(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	r, _ := rate.NewRate(pair, 0.061234, time.Now(), rate.SourceUnionPay)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	if !r.IsStale(5 * time.Millisecond) {
		t.Error("Rate should be stale after 5ms threshold")
	}

	if r.IsStale(100 * time.Millisecond) {
		t.Error("Rate should not be stale with 100ms threshold")
	}
}

func TestRate_Convert(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	r, _ := rate.NewRate(pair, 20.0, time.Now(), rate.SourceUnionPay)

	tests := []struct {
		name     string
		amount   float64
		expected float64
	}{
		{
			name:     "convert 100 CNY",
			amount:   100,
			expected: 2000,
		},
		{
			name:     "convert 1 CNY",
			amount:   1,
			expected: 20,
		},
		{
			name:     "convert 0",
			amount:   0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.Convert(tt.amount)
			if result != tt.expected {
				t.Errorf("Convert(%v) = %v, want %v", tt.amount, result, tt.expected)
			}
		})
	}
}

func TestRate_ConvertInverse(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	r, _ := rate.NewRate(pair, 20.0, time.Now(), rate.SourceUnionPay)

	result := r.ConvertInverse(100)
	expected := 5.0

	if result != expected {
		t.Errorf("ConvertInverse(100) = %v, want %v", result, expected)
	}
}

func TestRate_IsEffectiveOn(t *testing.T) {
	date := time.Date(2025, 11, 2, 10, 30, 0, 0, time.UTC)
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	r, _ := rate.NewRate(pair, 0.061234, date, rate.SourceUnionPay)

	tests := []struct {
		name      string
		checkDate time.Time
		expected  bool
	}{
		{
			name:      "same day different time",
			checkDate: time.Date(2025, 11, 2, 15, 0, 0, 0, time.UTC),
			expected:  true,
		},
		{
			name:      "different day",
			checkDate: time.Date(2025, 11, 3, 10, 30, 0, 0, time.UTC),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.IsEffectiveOn(tt.checkDate)
			if result != tt.expected {
				t.Errorf("IsEffectiveOn() = %v, want %v", result, tt.expected)
			}
		})
	}
}
