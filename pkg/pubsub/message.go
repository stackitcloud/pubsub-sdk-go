package pubsub

import (
	"context"
	"time"
)

type PullMessage struct {
	subscription     *Subscriber
	ID               uint64
	AckID            string
	Data             []byte
	CreateTime       time.Time
	DeliveryAttempts uint64
}

type PullMessages []PullMessage

func (m *PullMessage) Ack(ctx context.Context) error {
	return m.subscription.Ack(ctx, []string{m.AckID})
}

func (m *PullMessage) Nack(ctx context.Context) error {
	return m.subscription.Nack(ctx, []string{m.AckID})
}

func (m PullMessages) GetAckIDs() []string {
	ids := make([]string, len(m))
	for i, msg := range m {
		ids[i] = msg.AckID
	}
	return ids
}
