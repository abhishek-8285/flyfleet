package client

import (
	"net/http"
	"time"
)

type Option func(*Client)

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

func WithRetry(maxAttempts int, baseDelay, maxDelay time.Duration) Option {
	return func(c *Client) {
		c.retry = &RetryPolicy{
			MaxAttempts: maxAttempts,
			BaseDelay:   baseDelay,
			MaxDelay:    maxDelay,
		}
	}
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}