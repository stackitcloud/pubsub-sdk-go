package pubsub_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stackitcloud/pubsub-sdk-go/pkg/pubsub"
)

var _ = Describe("publish a message", func() {
	Context("publish a message", func() {
		It("should publish the message and return a message ID", func(ctx context.Context) {
			publisher := pubsub.NewPublisher(topicId, pubsub.WithHTTPRoundTripper(rt))
			messages := [][]byte{
				[]byte("Hello, Stackit!"),
			}

			messageIDs, err := publisher.Publish(ctx, messages)

			Expect(err).ToNot(HaveOccurred())
			Expect(messageIDs).ToNot(BeNil())
			Expect(messageIDs).To(HaveLen(1))
		})
	})

	Context("Purging a topic", func() {
		It("should successfully remove all messages from the topic", func(ctx context.Context) {
			publisher := pubsub.NewPublisher(topicId, pubsub.WithHTTPRoundTripper(rt))
			subscriber := pubsub.NewSubscriber(topicId, subscriptionId, pubsub.WithHTTPRoundTripper(rt))

			messagesToPublish := [][]byte{
				[]byte("message-to-be-purged-1"),
				[]byte("message-to-be-purged-2"),
			}

			_, err := publisher.Publish(ctx, messagesToPublish)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				msgsBefore, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(10))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(msgsBefore).ToNot(BeEmpty(), "Topic should have messages before purge")
			}).WithTimeout(5 * time.Second).WithPolling(500 * time.Millisecond).Should(Succeed())

			err = publisher.Purge(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				msgsAfter, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(10))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(msgsAfter).To(BeEmpty(), "Topic should be empty after purge")
			}).WithTimeout(5 * time.Second).WithPolling(500 * time.Millisecond).Should(Succeed())
		})
	})
})
