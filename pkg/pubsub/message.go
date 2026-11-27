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
	"time"
)

type PullMessage struct {
	subscription     *Subscriber
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

// DecodeString returns the message payload as a string.
// The error is always nil and only kept for API compatibility.
func (m *PullMessage) DecodeString() (string, error) {
	return string(m.Data), nil
}

// DecodeStrings returns the payloads of all messages as strings.
// The error is always nil and only kept for API compatibility.
func (m PullMessages) DecodeStrings() ([]string, error) {
	strings := make([]string, len(m))
	for i, msg := range m {
		strings[i] = string(msg.Data)
	}
	return strings, nil
}

// Bytes returns the original message payload.
func (m *PullMessage) Bytes() []byte {
	return m.Data
}
