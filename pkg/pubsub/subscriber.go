package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/google/uuid"

	"github.com/stackitcloud/pubsub-sdk-go.git/pkg/pubsub/api"
)

type Subscriber struct {
	SubscriptionId uuid.UUID
	logger         logr.Logger
	dataplane      *api.ClientWithResponses
	topicUrl       url.URL
	httpClient     *http.Client
}

func NewSubscriber(topicId uuid.UUID, subscriptionId uuid.UUID, opts ...Option) *Subscriber {
	cfg := &clientConfig{
		httpClient: http.DefaultClient,
		host:       "pubsub.eu01.onstackit.cloud",
		logger:     logr.FromSlogHandler(slog.Default().Handler()),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	topicUrl := url.URL{Scheme: "https", Host: fmt.Sprintf("%s.%s", topicId.String(), cfg.host)}

	subscriber := &Subscriber{
		SubscriptionId: subscriptionId,
		topicUrl:       topicUrl,
		httpClient:     cfg.httpClient,
		logger:         cfg.logger.WithValues("subscription_id", topicId),
	}

	subscriber.dataplane, _ = api.NewClientWithResponses(
		subscriber.topicUrl.String(),
		api.WithHTTPClient(subscriber.httpClient),
	)

	return subscriber
}

func (s *Subscriber) Ack(ctx context.Context, ids []string) error {
	reqBody := api.AckMessagesFromTopicRequest{
		AckIds: ids,
	}

	s.logger.V(4).Info("acknowledging messages",
		"count", len(ids),
	)

	resp, err := s.dataplane.AckMessagesWithResponse(ctx, s.SubscriptionId, reqBody)
	if err != nil {
		return NewNetworkError(
			"failed to execute ack messages request",
			err,
		)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return NewAPIError(resp.StatusCode(), resp.Body)
	}

	s.logger.V(4).Info("acknowledged messages",
		"count", len(ids),
	)
	return nil
}

func (s *Subscriber) Nack(ctx context.Context, ids []string) error {
	reqBody := api.NackMessagesFromTopicRequest{
		NackIds: ids,
	}

	s.logger.V(4).Info("nacking messages",
		"count", len(ids),
	)

	resp, err := s.dataplane.NackMessagesWithResponse(ctx, s.SubscriptionId, reqBody)
	if err != nil {
		return NewNetworkError(
			"failed to execute nack messages request",
			err,
		)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return NewAPIError(resp.StatusCode(), resp.Body)
	}

	s.logger.V(4).Info("nacked messages",
		"count", len(ids),
	)
	return nil
}

func toSdkMessages(m []api.Message, subscription *Subscriber) PullMessages {
	sdkMessages := make(PullMessages, len(m))
	for i, msg := range m {
		sdkMessages[i] = PullMessage{
			subscription:     subscription,
			AckID:            msg.AckId,
			CreateTime:       msg.CreateTime,
			Data:             msg.Data,
			DeliveryAttempts: msg.DeliveryAttempts,
			ID:               msg.Id,
		}
	}
	return sdkMessages
}

type pullOptions struct {
	maxMessages int32
}

type PullOption func(*pullOptions)

func WithMaxMessages(maximum int32) PullOption {
	return func(opts *pullOptions) {
		opts.maxMessages = maximum
	}
}

func (s *Subscriber) Pull(ctx context.Context, opts ...PullOption) (PullMessages, error) {
	cfg := &pullOptions{
		maxMessages: 64,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	reqBody := api.PullMessagesParams{
		PubSubMaxMessages: &cfg.maxMessages,
	}

	s.logger.V(4).Info("pulling messages",
		"max_messages", int(cfg.maxMessages),
	)

	resp, err := s.dataplane.PullMessagesWithResponse(ctx, s.SubscriptionId, &reqBody)
	if err != nil {
		return nil, &NetworkError{
			Msg: "failed to execute pull messages request",
			Err: err,
		}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, NewAPIError(resp.StatusCode(), resp.Body)
	}

	messages := toSdkMessages(resp.JSON200.Messages, s)

	s.logger.V(4).Info(
		"pulled messages",
		"count", len(resp.JSON200.Messages),
		"ack_ids", messages.GetAckIDs(),
	)
	return messages, nil
}
