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
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/google/uuid"

	"github.com/stackitcloud/pubsub-sdk-go/pkg/pubsub/api"
)

type Publisher struct {
	dataplane  *api.ClientWithResponses
	TopicID    uuid.UUID
	logger     logr.Logger
	topicURL   url.URL
	httpClient *http.Client
}

// NewPublisher instantiates a new Publisher.
func NewPublisher(topicID uuid.UUID, opts ...Option) *Publisher {
	cfg := &clientConfig{
		httpClient: &http.Client{},
		host:       "pubsub.eu01.onstackit.cloud",
		logger:     logr.FromSlogHandler(slog.Default().Handler()),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	topicURL := url.URL{Scheme: "https", Host: fmt.Sprintf("%s.%s", topicID.String(), cfg.host)}

	// SAFETY: The error here can never be non nil, as WithHTTPClient always returns a nil error.
	dataplane, _ := api.NewClientWithResponses(
		topicURL.String(),
		api.WithHTTPClient(cfg.httpClient),
	)

	publisher := &Publisher{
		TopicID:    topicID,
		topicURL:   topicURL,
		httpClient: cfg.httpClient,
		logger:     cfg.logger.WithValues("topic_id", topicID),
		dataplane:  dataplane,
	}

	return publisher
}

// PublishStrings acts as a lightweight adapter converting string slice to byte slices,
// deferring all encoding safely to the core Publish method.
func (p *Publisher) PublishStrings(ctx context.Context, messages ...string) ([]uint64, error) {
	byteMessages := make([][]byte, len(messages))
	for i, msg := range messages {
		byteMessages[i] = []byte(msg)
	}
	return p.Publish(ctx, byteMessages)
}

// Publish processes raw bytes, encodes them transparently using bytesToBase64,
// and transmits them out via the API client.
func (p *Publisher) Publish(ctx context.Context, messages [][]byte) ([]uint64, error) {
	encodedMessages := bytesToBase64(messages...)

	messagesToPublish := make([]api.PublishMessage, len(encodedMessages))
	for i, msg := range encodedMessages {
		messagesToPublish[i] = api.PublishMessage{
			Data: msg,
		}
	}

	reqBody := api.PublishMessagesJSONRequestBody{
		Messages: messagesToPublish,
	}

	p.logger.V(4).Info("publishing messages", "count", len(messages))
	resp, err := p.dataplane.PublishMessagesWithResponse(ctx, reqBody)
	if err != nil {
		return nil, NewNetworkError("failed to execute publish messages request", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, NewAPIError(resp.StatusCode(), resp.Body)
	}

	p.logger.V(4).Info(
		"published messages",
		"count", len(messages),
		"message_ids", resp.JSON200.MessageIds,
		"status_code", resp.StatusCode(),
	)
	return resp.JSON200.MessageIds, nil
}

// Purge removes all messages currently stored in the topic.
func (p *Publisher) Purge(ctx context.Context) error {
	p.logger.V(4).Info("purging topic")

	resp, err := p.dataplane.PurgeTopicWithResponse(ctx)
	if err != nil {
		return NewNetworkError("failed to execute purge topic request", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return NewAPIError(resp.StatusCode(), resp.Body)
	}

	p.logger.V(4).Info(
		"topic purged successfully",
		"status_code", resp.StatusCode(),
	)
	return nil
}
