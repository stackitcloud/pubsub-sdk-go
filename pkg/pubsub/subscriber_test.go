package pubsub_test

import (
	"context"
	"encoding/base64"
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
		messagesToPublish := pubsub.StringsToBase64("test1")

		_, err := publisher.Publish(ctx, messagesToPublish)
		Expect(err).ToNot(HaveOccurred())

		subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
		pulledMessages, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(1))
		Expect(err).ToNot(HaveOccurred())
		Expect(pulledMessages).ToNot(BeEmpty())

		AckIDs = pulledMessages.GetAckIDs()
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

var _ = Describe("Pull messages", func() {
	Context("pulling messages", func() {
		BeforeEach(func(ctx context.Context) {
			publisher := pubsub.NewPublisher(topicId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
			messagesToPublish := pubsub.StringsToBase64("test1", "test2", "test3")

			_, err := publisher.Publish(ctx, messagesToPublish)
			Expect(err).ToNot(HaveOccurred())
		})

		It("no error is occurring", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
			resp, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(128))
			Expect(resp).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
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
		// making sure everything is empty, and acking everything, stopping when len = 0
		subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))

		Eventually(func(g Gomega) bool {
			msgs, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(13))
			g.Expect(err).ToNot(HaveOccurred())

			if len(msgs) == 0 {
				return true
			}

			err = subscriber.Ack(ctx, msgs.GetAckIDs())
			g.Expect(err).ToNot(HaveOccurred())

			return false
		}).WithTimeout(10 * time.Second).WithPolling(500 * time.Millisecond).Should(BeTrue())

		// publishing test messages
		publisher := pubsub.NewPublisher(topicId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
		messagesToPublish := pubsub.StringsToBase64("testMessage", "testMessage2")

		_, err := publisher.Publish(ctx, messagesToPublish)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("using PullJobChan", func() {
		It("should receive a message from channel", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))

			var receivedMessages pubsub.PullMessages
			Eventually(func(g Gomega) {
				msgs, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(1))
				g.Expect(err).ToNot(HaveOccurred())
				if len(msgs) > 0 {
					receivedMessages = msgs
				}
				g.Expect(receivedMessages).ToNot(BeEmpty())
			}).WithContext(ctx).Should(Succeed())

			Expect(receivedMessages).To(HaveLen(1))

			decodedData, err := base64.StdEncoding.DecodeString(string(receivedMessages[0].Data))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(decodedData)).To(Or(Equal("testMessage"), Equal("testMessage2")))
			err = subscriber.Ack(ctx, receivedMessages.GetAckIDs())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("using PullJobCallback", func() {
		It("should invoke the callback with messages", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt), pubsub.WithHost(environment))
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			var callbackInvoked atomic.Bool // check if callback was invoked
			callback := func(ctx context.Context, messages pubsub.PullMessages) {
				defer GinkgoRecover()
				Expect(messages).To(HaveLen(1))
				decodedData, err := base64.StdEncoding.DecodeString(string(messages[0].Data))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(decodedData)).To(Or(Equal("testMessage"), Equal("testMessage2")))
				err = subscriber.Ack(ctx, messages.GetAckIDs())
				Expect(err).ToNot(HaveOccurred())
				callbackInvoked.Store(true)
				cancel()
			}

			err := subscriber.PullJobCallback(
				ctx,
				callback,
				pubsub.WithInterval(100*time.Millisecond),
				pubsub.WithPullMaxMessages(1),
			)
			Expect(err).ToNot(HaveOccurred())

			<-ctx.Done()
			Expect(callbackInvoked.Load()).To(BeTrue(), "callback should have been invoked")
		})
	})
})
