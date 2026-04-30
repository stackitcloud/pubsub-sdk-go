package pubsub_test

import (
	"context"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"dev.azure.com/schwarzit-wiking/schwarzit.stackit-pubsub/stackit-pubsub-go-sdk.git/pkg/pubsub"
)

var _ = Describe("Acknowledge messages", func() {
	var AckIDs []string

	BeforeEach(func(ctx context.Context) {
		publisher := pubsub.NewPublisher(topicId, pubsub.WithHTTPRoundTripper(rt))
		messagesToPublish := [][]byte{
			[]byte("test1"),
		}
		_, err := publisher.Publish(ctx, messagesToPublish)
		Expect(err).ToNot(HaveOccurred())

		subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))
		pulledMessages, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(1))
		Expect(err).ToNot(HaveOccurred())
		Expect(pulledMessages).ToNot(BeEmpty())

		AckIDs = pulledMessages.GetAckIDs()
	})

	Context("acknowledging messages", func() {
		It("should acknowledge a single message", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))
			ackId := AckIDs[0]

			err := subscriber.Ack(ctx, []string{ackId})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should acknowledge multiple messages", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))

			err := subscriber.Ack(ctx, AckIDs)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("not acknowledging messages", func() {
		It("should not acknowledge a single message", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))
			nackId := AckIDs[0]

			err := subscriber.Nack(ctx, []string{nackId})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not Acknowledge multiple messages", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))

			err := subscriber.Nack(ctx, AckIDs)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("Pull messages", func() {
	Context("pulling messages", func() {
		BeforeEach(func(ctx context.Context) {
			publisher := pubsub.NewPublisher(topicId, pubsub.WithHTTPRoundTripper(rt))
			messagesToPublish := [][]byte{
				[]byte("test1"),
				[]byte("test2"),
				[]byte("test3"),
			}
			_, err := publisher.Publish(ctx, messagesToPublish)
			Expect(err).ToNot(HaveOccurred())
		})

		It("no error is occurring", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))
			resp, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(128))
			Expect(resp).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
		It(
			"should pull only one Message, MaxMessages is set to 1 but more messages will be available",
			func(ctx context.Context) {
				subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))
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
		subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))

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
		publisher := pubsub.NewPublisher(topicId, pubsub.WithHTTPRoundTripper(rt))
		messagesToPublish := [][]byte{
			[]byte("testMessage"),
			[]byte("testMessage2"),
		}
		_, err := publisher.Publish(ctx, messagesToPublish)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("using PullJobChan", func() {
		It("should receive a message from channel", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))

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
			Expect(string(receivedMessages[0].Data)).To(Or(Equal("testMessage"), Equal("testMessage2")))
			err := subscriber.Ack(ctx, receivedMessages.GetAckIDs())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("using PullJobCallback", func() {
		It("should invoke the callback with messages", func(ctx context.Context) {
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			var callbackInvoked atomic.Bool // check if callback was invoked
			callback := func(ctx context.Context, messages pubsub.PullMessages) {
				defer GinkgoRecover()
				Expect(messages).To(HaveLen(1))
				Expect(string(messages[0].Data)).To(Or(Equal("testMessage"), Equal("testMessage2")))
				err := subscriber.Ack(ctx, messages.GetAckIDs())
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
