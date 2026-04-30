package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/google/uuid"

	"dev.azure.com/schwarzit-wiking/schwarzit.stackit-pubsub/stackit-pubsub-go-sdk.git/pkg/pubsub/api"
)

type Publisher struct {
	dataplane  *api.ClientWithResponses
	TopicId    uuid.UUID
	logger     logr.Logger
	topicUrl   url.URL
	httpClient *http.Client
}

func NewPublisher(topicId uuid.UUID, opts ...Option) *Publisher {
	cfg := &clientConfig{
		httpClient: http.DefaultClient,
		host:       "pubsub.eu01.onstackit.cloud",
		logger:     logr.FromSlogHandler(slog.Default().Handler()),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	topicUrl := url.URL{Scheme: "https", Host: fmt.Sprintf("%s.%s", topicId.String(), cfg.host)}

	publisher := &Publisher{
		TopicId:    topicId,
		topicUrl:   topicUrl,
		httpClient: cfg.httpClient,
		logger:     cfg.logger.WithValues("topic_id", topicId),
	}

	publisher.dataplane, _ = api.NewClientWithResponses(
		publisher.topicUrl.String(),
		api.WithHTTPClient(publisher.httpClient),
	)

	return publisher
}

func (p *Publisher) Publish(ctx context.Context, messages [][]byte) ([]uint64, error) {
	messagesToPublish := make([]api.PublishMessage, len(messages))
	for i, msg := range messages {
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
		return nil, &NetworkError{
			Msg: "failed to execute publish messages request",
			Err: err,
		}
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
// This is a destructive operation and cannot be undone.
func (p *Publisher) Purge(ctx context.Context) error {
	p.logger.V(4).Info("purging topic")

	resp, err := p.dataplane.PurgeTopicWithResponse(ctx)
	if err != nil {
		return &NetworkError{
			Msg: "failed to execute purge topic request",
			Err: err,
		}
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
