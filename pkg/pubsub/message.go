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

func (m *PullMessage) AsString() (string, error) {
	str, err := base64ToStrings(m.Data)
	if err != nil {
		return "", err
	}
	return str[0], nil
}

func (m *PullMessage) AsBytes() []byte {
	return m.Data
}

type PullMessages []PullMessage

func (m PullMessages) AsStrings() ([]string, error) {
	strings := make([]string, 0, len(m))
	for _, msg := range m {
		str, _ := msg.AsString()
		strings = append(strings, str)
	}
	return strings, nil
}
