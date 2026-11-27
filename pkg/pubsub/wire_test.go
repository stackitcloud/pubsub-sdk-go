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

package pubsub_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/onsi/gomega"

	"github.com/stackitcloud/pubsub-sdk-go/pkg/pubsub"
)

type stubRoundTripper func(*http.Request) *http.Response

func (f stubRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r), nil
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

// Plain Go test so it runs without the Ginkgo suite's environment (BeforeSuite).
func TestWire(t *testing.T) {
	payload := []byte("Hello, Stackit!\x00\xff")
	wire := base64.StdEncoding.EncodeToString(payload)

	t.Run("base64-encodes published payloads exactly once", func(t *testing.T) {
		g := gomega.NewWithT(t)
		var requestBody []byte
		rt := stubRoundTripper(func(r *http.Request) *http.Response {
			requestBody, _ = io.ReadAll(r.Body)
			return jsonResponse(http.StatusNoContent, "")
		})
		publisher := pubsub.NewPublisher(uuid.New(), pubsub.WithHTTPRoundTripper(rt))

		g.Expect(publisher.Publish(context.Background(), [][]byte{payload})).To(gomega.Succeed())

		g.Expect(requestBody).To(gomega.MatchJSON(`{"messages":[{"data":"` + wire + `"}]}`))
	})

	t.Run("returns the original bytes for pulled messages", func(t *testing.T) {
		g := gomega.NewWithT(t)
		rt := stubRoundTripper(func(*http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"messages":[{"ackId":"a","createTime":"2026-01-01T00:00:00Z","data":"`+wire+`","deliveryAttempts":1}]}`)
		})
		subscriber := pubsub.NewSubscriber(uuid.New(), uuid.New(), pubsub.WithHTTPRoundTripper(rt))

		messages, err := subscriber.Pull(context.Background())

		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(messages).To(gomega.HaveLen(1))
		g.Expect(messages[0].Bytes()).To(gomega.Equal(payload))
		str, err := messages[0].DecodeString()
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(str).To(gomega.Equal(string(payload)))
	})
}
