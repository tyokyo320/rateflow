package unionpay

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
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
		// Check if it's a 404 error which might indicate historical data not available
		if strings.Contains(err.Error(), "404") {
			c.logger.Warn("UnionPay API returned 404 - historical data may not be available for this date",
				"pair", pair.String(),
				"date", dateStr,
				"url", url,
			)
			return 0, provider.NewProviderError(
				c.Name(),
				fmt.Sprintf("data not available for %s (404 - possibly too old or API unavailable)", dateStr),
				err,
			)
		}
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

	// Get base and quote currency codes
	baseCur := string(pair.Base())
	quoteCur := string(pair.Quote())

	// Find the rate in response
	// UnionPay format: transCur (quote) / baseCur (base)
	for _, item := range resp.ExchangeRateJSON {
		if item.TransCur == quoteCur && item.BaseCur == baseCur {
			c.logger.Info("rate fetched successfully",
				"pair", pair.String(),
				"rate", item.RateData,
				"date", dateStr,
			)
			return item.RateData, nil
		}
	}

	// Rate not found (weekends/holidays or unsupported pair)
	c.logger.Warn("rate not found in response",
		"pair", pair.String(),
		"date", dateStr,
	)

	return 0, provider.NewProviderError(
		c.Name(),
		fmt.Sprintf("rate not found for %s (possibly weekend/holiday or unsupported pair)", pair.String()),
		nil,
	)
}

// FetchLatest fetches the latest available exchange rate.
func (c *Client) FetchLatest(ctx context.Context, pair currency.Pair) (float64, error) {
	return c.FetchRate(ctx, pair, time.Now())
}

// SupportedPairs returns the list of supported currency pairs.
// UnionPay supports 12 base currencies and 160+ target currencies.
// We return the most commonly used pairs here.
func (c *Client) SupportedPairs() []currency.Pair {
	// List of major supported pairs
	// UnionPay actually supports many more (12 base × 160+ target currencies)
	return []currency.Pair{
		// CNY pairs
		currency.MustNewPair(currency.CNY, currency.JPY),
		currency.MustNewPair(currency.CNY, currency.USD),
		currency.MustNewPair(currency.CNY, currency.EUR),
		currency.MustNewPair(currency.CNY, currency.GBP),
		currency.MustNewPair(currency.CNY, currency.HKD),

		// JPY pairs
		currency.MustNewPair(currency.JPY, currency.USD),
		currency.MustNewPair(currency.JPY, currency.EUR),
		currency.MustNewPair(currency.JPY, currency.CNY),

		// USD pairs
		currency.MustNewPair(currency.USD, currency.JPY),
		currency.MustNewPair(currency.USD, currency.EUR),
		currency.MustNewPair(currency.USD, currency.CNY),

		// Other major pairs
		currency.MustNewPair(currency.EUR, currency.USD),
		currency.MustNewPair(currency.EUR, currency.JPY),
		currency.MustNewPair(currency.GBP, currency.USD),
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
