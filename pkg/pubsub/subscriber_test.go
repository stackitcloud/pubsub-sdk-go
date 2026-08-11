package pubsub_test

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stackitcloud/pubsub-sdk-go/pkg/pubsub"
)

var _ = Describe("Acknowledge messages", func() {
	var AckIDs []string

	BeforeEach(func(ctx context.Context) {
		publisher := pubsub.NewPublisher(topicId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))

		_, err := publisher.PublishStrings(ctx, "test1")
		Expect(err).ToNot(HaveOccurred())

		subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
		pulledMessages, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(1))
		Expect(err).ToNot(HaveOccurred())
		Expect(pulledMessages).ToNot(BeEmpty())

		AckIDs = pulledMessages.AckIDs()
	})

	Context("acknowledging messages", func() {
		It("should acknowledge a single message", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
			ackId := AckIDs[0]

			err := subscriber.Ack(ctx, []string{ackId})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should acknowledge multiple messages", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))

			err := subscriber.Ack(ctx, AckIDs)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("not acknowledging messages", func() {
		It("should not acknowledge a single message", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
			nackId := AckIDs[0]

			err := subscriber.Nack(ctx, []string{nackId})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not Acknowledge multiple messages", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))

			err := subscriber.Nack(ctx, AckIDs)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("WithLongPullDuration validation", func() {
	DescribeTable("invalid durations return ConfigurationError",
		func(ctx context.Context, ms int32) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId)
			_, err := subscriber.Pull(ctx, pubsub.WithLongPullDuration(ms))
			Expect(err).To(HaveOccurred())
			var cfgErr *pubsub.ConfigurationError
			Expect(errors.As(err, &cfgErr)).To(BeTrue())
		},
		Entry("below minimum", int32(50)),
		Entry("above maximum", int32(6000)),
	)

	DescribeTable("valid durations do not return ConfigurationError",
		func(ms int32) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId)
			_, err := subscriber.Pull(context.Background(), pubsub.WithLongPullDuration(ms))
			if err != nil {
				var cfgErr *pubsub.ConfigurationError
				Expect(errors.As(err, &cfgErr)).To(BeFalse(), "expected no ConfigurationError for ms=%d", ms)
			}
		},
		Entry("minimum (100)", int32(100)),
		Entry("maximum (5000)", int32(5000)),
	)
})

var _ = Describe("Pull messages", func() {
	Context("pulling messages", func() {
		BeforeEach(func(ctx context.Context) {
			publisher := pubsub.NewPublisher(topicId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))

			_, err := publisher.PublishStrings(ctx, "test1", "test2", "test3")
			Expect(err).ToNot(HaveOccurred())
		})

		It("no error is occurring", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
			resp, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(32))
			Expect(resp).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
		It("should pull messages with long pull duration set", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
			resp, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(32), pubsub.WithLongPullDuration(500))
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).ToNot(BeNil())
		})
		It(
			"should pull only one Message, MaxMessages is set to 1 but more messages will be available",
			func(ctx context.Context) {
				subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
				resp, _ := subscriber.Pull(ctx, pubsub.WithMaxMessages(1))
				Expect(resp).To(HaveLen(1))
				Expect(resp).ToNot(HaveLen(2))
			},
		)
	})
})

var _ = Describe("PullJob", func() {
	BeforeEach(func(ctx context.Context) {
		publisher := pubsub.NewPublisher(topicId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
		// making sure everything is empty
		err := publisher.Purge(ctx)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(1 * time.Second) // wait for topic to be purged

		// publishing test messages
		_, err = publisher.PublishStrings(ctx, "testMessage")
		Expect(err).ToNot(HaveOccurred())
	})

	Context("using PullJobChan", func() {
		It("should receive a message from channel", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
			defer subscriber.Wait()

			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			jobChan, err := subscriber.PullJobChan(ctx,
				pubsub.WithPullLongPullDuration(100),
				pubsub.WithPullMaxMessages(1),
			)
			Expect(err).ToNot(HaveOccurred())

			var receivedMessages pubsub.PullMessages
			Eventually(jobChan, "10s").Should(Receive(&receivedMessages))
			Expect(receivedMessages).To(HaveLen(1))

			decodedString, err := receivedMessages[0].DecodeString()
			Expect(err).ToNot(HaveOccurred())
			Expect(decodedString).To(Equal("testMessage"))
			err = subscriber.Ack(ctx, receivedMessages.AckIDs())
			Expect(err).ToNot(HaveOccurred())

			cancel()
		})
	})

	Context("using PullJobCallback", func() {
		It("should invoke the callback with messages", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
			defer subscriber.Wait()
			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			var callbackInvoked atomic.Bool // check if callback was invoked
			callback := func(ctx context.Context, messages pubsub.PullMessages) {
				defer GinkgoRecover()
				Expect(messages).To(HaveLen(1))
				decoded, err := messages[0].DecodeString()
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal("testMessage"))
				err = subscriber.Ack(ctx, messages.AckIDs())
				Expect(err).ToNot(HaveOccurred())
				callbackInvoked.Store(true)
				cancel()
			}

			err := subscriber.PullJobCallback(
				ctx,
				callback,
				pubsub.WithPullMaxMessages(1),
				pubsub.WithPullLongPullDuration(100),
			)
			Expect(err).ToNot(HaveOccurred())

			<-ctx.Done()
			Expect(callbackInvoked.Load()).To(BeTrue(), "callback should have been invoked")
		})
	})
})
