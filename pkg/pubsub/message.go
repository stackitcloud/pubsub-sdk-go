package pubsub

import (
	"context"
	"fmt"
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

type PullMessages []PullMessage

// AckIDs extracts all AckIDs cleanly from a slice of PullMessages.
func (m PullMessages) AckIDs() []string {
	ids := make([]string, len(m))
	for i, msg := range m {
		ids[i] = msg.AckID
	}
	return ids
}

// DecodeString reverses the transparent base64 encoding, returning the cleartext string.
func (m *PullMessage) DecodeString() (string, error) {
	decoded, err := base64Decode(m.Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode message data: %w", err)
	}
	return string(decoded), nil
}

// DecodeStrings decodes an entire slice of messages.
// If any single message is corrupt, it returns the error immediately instead of swallowing it.
func (m PullMessages) DecodeStrings() ([]string, error) {
	strings := make([]string, len(m))
	for i, msg := range m {
		str, err := msg.DecodeString()
		if err != nil {
			return nil, fmt.Errorf("failed to decode message at index %d: %w", i, err)
		}
		strings[i] = str
	}
	return strings, nil
}

// Bytes exposes the underlying raw base64 data.
func (m *PullMessage) Bytes() []byte {
	return m.Data
}
