package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-logr/logr"
	"github.com/google/uuid"

	"github.com/stackitcloud/pubsub-sdk-go/pkg/pubsub/api"
)

type Subscriber struct {
	SubscriptionID uuid.UUID
	logger         logr.Logger
	dataplane      *api.ClientWithResponses
	topicURL       url.URL
	httpClient     *http.Client
	wg             sync.WaitGroup
}

// NewSubscriber instantiates a new Subscriber. It returns an error if the underlying
// API dataplane client fails to initialize.
func NewSubscriber(topicID uuid.UUID, subscriptionID uuid.UUID, opts ...Option) *Subscriber {
	cfg := &clientConfig{
		httpClient: http.DefaultClient,
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

	subscriber := &Subscriber{
		SubscriptionID: subscriptionID,
		topicURL:       topicURL,
		httpClient:     cfg.httpClient,
		logger:         cfg.logger.WithValues("subscription_id", subscriptionID),
		dataplane:      dataplane,
	}

	return subscriber
}

func (s *Subscriber) Ack(ctx context.Context, ids []string) error {
	reqBody := api.AckMessagesFromTopicRequest{
		AckIds: ids,
	}

	s.logger.V(4).Info("acknowledging messages", "count", len(ids))

	resp, err := s.dataplane.AckMessagesWithResponse(ctx, s.SubscriptionID, reqBody)
	if err != nil {
		return NewNetworkError("failed to execute ack messages request", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return NewAPIError(resp.StatusCode(), resp.Body)
	}

	s.logger.V(4).Info("acknowledged messages", "count", len(ids))
	return nil
}

func (s *Subscriber) Nack(ctx context.Context, ids []string) error {
	reqBody := api.NackMessagesFromTopicRequest{
		NackIds: ids,
	}

	s.logger.V(4).Info("nacking messages", "count", len(ids))

	resp, err := s.dataplane.NackMessagesWithResponse(ctx, s.SubscriptionID, reqBody)
	if err != nil {
		return NewNetworkError("failed to execute nack messages request", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return NewAPIError(resp.StatusCode(), resp.Body)
	}

	s.logger.V(4).Info("nacked messages", "count", len(ids))
	return nil
}

func toSDKMessages(m []api.Message, subscription *Subscriber) PullMessages {
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
	maxMessages      int32
	longPullDuration *int32
}

type PullOption func(*pullOptions)

func WithMaxMessages(maximum int32) PullOption {
	return func(opts *pullOptions) {
		opts.maxMessages = maximum
	}
}

func WithLongPullDuration(milliseconds int32) PullOption {
	return func(opts *pullOptions) {
		opts.longPullDuration = &milliseconds
	}
}

func (s *Subscriber) Pull(ctx context.Context, opts ...PullOption) (PullMessages, error) {
	cfg := &pullOptions{
		maxMessages: 32,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	var longPullDuration *int32
	if cfg.longPullDuration != nil && *cfg.longPullDuration != 0 {
		ms := *cfg.longPullDuration
		if ms < 100 || ms > 5000 {
			return nil, &ConfigurationError{
				Msg: fmt.Sprintf("long_pull_duration must be 0 (default) or between 100–5000, got %d", ms),
			}
		}
		longPullDuration = &ms
	}

	reqBody := api.PullMessagesParams{
		PubSubMaxMessages:      &cfg.maxMessages,
		PubSubLongPullDuration: longPullDuration,
	}

	s.logger.V(4).Info("pulling messages", "max_messages", int(cfg.maxMessages))

	resp, err := s.dataplane.PullMessagesWithResponse(ctx, s.SubscriptionID, &reqBody)
	if err != nil {
		return nil, NewNetworkError("failed to execute pull messages request", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, NewAPIError(resp.StatusCode(), resp.Body)
	}

	messages := toSDKMessages(resp.JSON200.Messages, s)

	s.logger.V(4).Info(
		"pulled messages",
		"count", len(resp.JSON200.Messages),
		"ack_ids", messages.AckIDs(),
	)
	return messages, nil
}

func (s *Subscriber) Wait() {
	s.wg.Wait()
}
