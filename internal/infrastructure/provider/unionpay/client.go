package unionpay

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/domain/provider"
	"github.com/tyokyo320/rateflow/pkg/httputil"
	"github.com/tyokyo320/rateflow/pkg/timeutil"
)

const (
	baseURL = "https://m.unionpayintl.com/jfimg"
)

// Response represents the UnionPay API response structure.
type Response struct {
	ExchangeRateJSON []struct {
		TransCur string  `json:"transCur"`
		BaseCur  string  `json:"baseCur"`
		RateData float64 `json:"rateData"`
	} `json:"exchangeRateJson"`
	CurDate string `json:"curDate"`
}

// Client implements the UnionPay rate provider.
type Client struct {
	http   *httputil.Client
	logger *slog.Logger
}

// NewClient creates a new UnionPay provider client.
func NewClient(logger *slog.Logger) provider.Provider {
	return &Client{
		http:   httputil.NewClient(httputil.DefaultConfig()),
		logger: logger,
	}
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "unionpay"
}

// FetchRate fetches the exchange rate for a specific currency pair and date.
func (c *Client) FetchRate(ctx context.Context, pair currency.Pair, date time.Time) (float64, error) {
	// UnionPay only supports CNY/JPY
	if !c.supportsPair(pair) {
		return 0, provider.NewProviderError(
			c.Name(),
			fmt.Sprintf("unsupported currency pair: %s", pair.String()),
			nil,
		)
	}

	// Build URL with date
	dateStr := timeutil.FormatCompactDate(date)
	url := fmt.Sprintf("%s/%s.json", baseURL, dateStr)

	c.logger.Debug("fetching rate from unionpay",
		"url", url,
		"pair", pair.String(),
		"date", dateStr,
	)

	// Fetch data
	data, err := c.http.GetJSON(ctx, url, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
	})
	if err != nil {
		return 0, provider.NewProviderError(
			c.Name(),
			"failed to fetch data",
			err,
		)
	}

	// Parse response
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, provider.NewProviderError(
			c.Name(),
			"failed to parse response",
			err,
		)
	}

	// Find the rate for JPY/CNY
	for _, item := range resp.ExchangeRateJSON {
		if item.TransCur == "JPY" && item.BaseCur == "CNY" {
			c.logger.Info("rate fetched successfully",
				"pair", pair.String(),
				"rate", item.RateData,
				"date", dateStr,
			)
			return item.RateData, nil
		}
	}

	// Rate not found (weekends/holidays)
	c.logger.Warn("rate not found in response",
		"pair", pair.String(),
		"date", dateStr,
	)

	return 0, provider.NewProviderError(
		c.Name(),
		"rate not found in response (possibly weekend/holiday)",
		nil,
	)
}

// FetchLatest fetches the latest available exchange rate.
func (c *Client) FetchLatest(ctx context.Context, pair currency.Pair) (float64, error) {
	return c.FetchRate(ctx, pair, time.Now())
}

// SupportedPairs returns the list of supported currency pairs.
func (c *Client) SupportedPairs() []currency.Pair {
	return []currency.Pair{
		currency.MustNewPair(currency.CNY, currency.JPY),
	}
}

// SupportsMulti returns false as UnionPay doesn't support batch fetching.
func (c *Client) SupportsMulti() bool {
	return false
}

// FetchMulti is not supported by UnionPay.
func (c *Client) FetchMulti(ctx context.Context, pairs []currency.Pair, date time.Time) (map[string]float64, error) {
	return nil, provider.NewProviderError(
		c.Name(),
		"batch fetch not supported",
		nil,
	)
}

// supportsPair checks if the provider supports the given currency pair.
func (c *Client) supportsPair(pair currency.Pair) bool {
	supported := c.SupportedPairs()
	for _, p := range supported {
		if p.Equal(pair) {
			return true
		}
	}
	return false
}
