// Copyright 2026 Schwarz Digits Cloud GmbH & Co. KG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
