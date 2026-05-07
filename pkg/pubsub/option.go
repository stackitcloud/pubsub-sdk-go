package pubsub

import (
	"log/slog"
	"net/http"

	"github.com/go-logr/logr"
)

// clientConfig holds all shared configurable properties for pub/sub clients.
type clientConfig struct {
	httpClient *http.Client
	host       string
	logger     logr.Logger
}

// Option defines the functional option signature for configuring clients.
type Option func(*clientConfig)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = client
	}
}

// WithHTTPRoundTripper sets a custom transport for the default HTTP client.
func WithHTTPRoundTripper(rt http.RoundTripper) Option {
	return func(c *clientConfig) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Transport = rt
	}
}

// WithHost sets a custom host for the data plane API.
func WithHost(host string) Option {
	return func(c *clientConfig) {
		c.host = host
	}
}

// WithLogger sets the logger using a standard slog.Logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *clientConfig) {
		c.logger = logr.FromSlogHandler(logger.Handler())
	}
}

// WithLogrLogger sets the logger using a go-logr instance.
func WithLogrLogger(logger logr.Logger) Option {
	return func(c *clientConfig) {
		c.logger = logger
	}
}
