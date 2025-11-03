// Package httputil provides HTTP client utilities with retry and timeout support.
package httputil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client wraps http.Client with additional utilities.
type Client struct {
	client  *http.Client
	retries int
	timeout time.Duration
}

// Config holds configuration for the HTTP client.
type Config struct {
	Timeout time.Duration
	Retries int
}

// DefaultConfig returns the default HTTP client configuration.
func DefaultConfig() Config {
	return Config{
		Timeout: 30 * time.Second,
		Retries: 3,
	}
}

// NewClient creates a new HTTP client with the given configuration.
func NewClient(cfg Config) *Client {
	return &Client{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		retries: cfg.Retries,
		timeout: cfg.Timeout,
	}
}

// Get performs a GET request with retry logic.
func (c *Client) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.doWithRetry(req)
}

// Post performs a POST request with retry logic.
func (c *Client) Post(ctx context.Context, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.doWithRetry(req)
}

// Do executes an HTTP request with retry logic.
func (c *Client) Do(req *http.Request) ([]byte, error) {
	return c.doWithRetry(req)
}

func (c *Client) doWithRetry(req *http.Request) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed (attempt %d/%d): %w", attempt+1, c.retries+1, err)
			continue
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("read response body: %w", err)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return body, nil
		}

		// Retry on server errors (5xx)
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d (attempt %d/%d)", resp.StatusCode, attempt+1, c.retries+1)
			continue
		}

		// Don't retry on client errors (4xx)
		return nil, fmt.Errorf("client error: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retries+1, lastErr)
}

// GetJSON performs a GET request and expects JSON response.
func (c *Client) GetJSON(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Accept"] = "application/json"
	return c.Get(ctx, url, headers)
}

// SetTimeout updates the client timeout.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.client.Timeout = timeout
}

// SetRetries updates the retry count.
func (c *Client) SetRetries(retries int) {
	c.retries = retries
}
