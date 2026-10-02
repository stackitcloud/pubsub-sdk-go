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
