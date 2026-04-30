package pubsub_test

import (
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"dev.azure.com/schwarzit-wiking/schwarzit.stackit-pubsub/stackit-pubsub-go-sdk.git/pkg/pubsub"
)

var _ = Describe("SDKError Interfaces", func() {
	Context("ConfigurationError", func() {
		ErrTestSentinel := pubsub.NewConfigurationError("test sentinel config error", nil)

		It("acts as a sentinel error", func() {
			err := ErrTestSentinel

			Expect(err.Error()).To(Equal("pubsub configuration error: test sentinel config error"))
			Expect(err.IsTransient()).To(BeFalse())
			Expect(err.Unwrap()).To(Succeed())

			Expect(errors.Is(err, ErrTestSentinel)).To(BeTrue())

			wrapped := fmt.Errorf("wrapper: %w", err)
			Expect(errors.Is(wrapped, ErrTestSentinel)).To(BeTrue())
			Expect(wrapped).To(MatchError(ErrTestSentinel))
		})

		It("correctly wraps and formats underlying errors", func() {
			rootCause := errors.New("missing credentials")
			err := pubsub.NewConfigurationError("init failed", rootCause)

			Expect(err.Error()).To(Equal("pubsub configuration error: init failed: missing credentials"))
			Expect(err.Unwrap()).To(Equal(rootCause))

			// Verify we can find the root cause via errors.Is
			Expect(errors.Is(err, rootCause)).To(BeTrue())
		})
	})

	Context("NetworkError", func() {
		It("wraps the error and is always transient", func() {
			rootCause := errors.New("dial tcp: i/o timeout")
			err := pubsub.NewNetworkError("failed to reach dataplane", rootCause)

			Expect(err.Error()).To(Equal("pubsub network error: failed to reach dataplane: dial tcp: i/o timeout"))
			Expect(err.IsTransient()).To(BeTrue(), "network errors should always be retryable")
			Expect(err.Unwrap()).To(Equal(rootCause))
			Expect(errors.Is(err, rootCause)).To(BeTrue())
		})

		It("wraps an underlying error using functional options and allows type extraction", func() {
			rootCause := errors.New("connection reset by peer")
			netErr := pubsub.NewNetworkError("dataplane unreachable", rootCause)

			// A StatusBadGateway is not triggered by a "connection reset by peer" error, but this is just a test
			apiErr := pubsub.NewAPIError(
				http.StatusBadGateway,
				[]byte(`{"message": "bad gateway"}`),
				pubsub.WithUnderlyingError(netErr),
			)

			Expect(errors.Is(apiErr, netErr)).To(BeTrue())
			Expect(errors.Is(apiErr, rootCause)).To(BeTrue())

			var extractedNetErr *pubsub.NetworkError
			Expect(errors.As(apiErr, &extractedNetErr)).To(BeTrue())

			Expect(extractedNetErr.Msg).To(Equal("dataplane unreachable"))
			Expect(extractedNetErr.IsTransient()).To(BeTrue())
		})
	})

	Context("APIError", func() {
		It("parses standard JSON error responses", func() {
			body := []byte(`{"message": "topic does not exist", "error": "topic_not_found"}`)
			err := pubsub.NewAPIError(http.StatusNotFound, body)

			// Format and properties
			Expect(err.Error()).To(Equal("pubsub api error (status: 404, code: topic_not_found): topic does not exist"))
			Expect(err.Code).To(Equal("topic_not_found"))
			Expect(err.Msg).To(Equal("topic does not exist"))

			// Status checkers
			Expect(err.IsNotFound()).To(BeTrue())
			Expect(err.IsConflict()).To(BeFalse())
			Expect(err.IsTransient()).To(BeFalse(), "404 is not transient")
		})

		It("identifies transient server and rate limit errors", func() {
			err500 := pubsub.NewAPIError(http.StatusInternalServerError, []byte(`{"message": "server on fire"}`))
			Expect(err500.IsTransient()).To(BeTrue(), "500 should be transient")

			err429 := pubsub.NewAPIError(http.StatusTooManyRequests, []byte(`{"message": "slow down"}`))
			Expect(err429.IsTransient()).To(BeTrue(), "429 should be transient")
		})

		It("falls back cleanly on malformed JSON or plain text", func() {
			body := []byte("gateway timeout html or plain text")
			err := pubsub.NewAPIError(http.StatusGatewayTimeout, body)

			Expect(err.Code).To(Equal("unknown_error"))
			Expect(err.Msg).To(Equal("gateway timeout html or plain text"))
			Expect(err.IsTransient()).To(BeTrue()) // 504 is >= 500
		})

		It("falls back if JSON is valid but required fields are missing", func() {
			body := []byte(`{"unrelated_field": "foo"}`)
			err := pubsub.NewAPIError(http.StatusBadRequest, body)

			Expect(err.Code).To(Equal("unknown_error"))
			Expect(err.Msg).To(Equal(`{"unrelated_field": "foo"}`))
			Expect(err.IsBadRequest()).To(BeTrue())
		})

		It("handles completely empty response bodies", func() {
			err := pubsub.NewAPIError(http.StatusInternalServerError, []byte{})

			Expect(err.Code).To(Equal("unknown_error"))
			Expect(err.Msg).To(Equal("api request failed with no response body"))
		})

		It("identifies unauthorized, conflict, and payload too large errors", func() {
			err401 := pubsub.NewAPIError(http.StatusUnauthorized, []byte(`{"message": "unauthorized"}`))
			Expect(err401.IsUnauthorized()).To(BeTrue())

			err409 := pubsub.NewAPIError(http.StatusConflict, []byte(`{"message": "resource conflict"}`))
			Expect(err409.IsConflict()).To(BeTrue())

			err413 := pubsub.NewAPIError(http.StatusRequestEntityTooLarge, []byte(`{"message": "message too big"}`))
			Expect(err413.IsRequestEntityTooLarge()).To(BeTrue())
		})

		It("correctly unwraps an underlying error", func() {
			rootCause := errors.New("custom wrapped error")
			err := pubsub.NewAPIError(
				http.StatusInternalServerError,
				[]byte(`{"message": "internal error"}`),
				pubsub.WithUnderlyingError(rootCause),
			)

			Expect(err.Unwrap()).To(Equal(rootCause))
			Expect(errors.Is(err, rootCause)).To(BeTrue())
		})

		It("handles valid JSON responses that only contain the 'error' field", func() {
			body := []byte(`{"error": "invalid_parameter"}`)
			err := pubsub.NewAPIError(http.StatusBadRequest, body)

			Expect(err.Code).To(Equal("invalid_parameter"))
			Expect(err.Msg).To(Equal("api request failed"))
		})

		It("handles valid JSON responses that only contain the 'message' field", func() {
			body := []byte(`{"message": "something went wrong"}`)
			err := pubsub.NewAPIError(http.StatusBadRequest, body)

			Expect(err.Msg).To(Equal("something went wrong"))
			Expect(err.Code).To(Equal("api_error"))
		})
	})
})
