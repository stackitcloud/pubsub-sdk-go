package example

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/stackitcloud/pubsub-sdk-go.git/pkg/pubsub"
	"github.com/stackitcloud/stackit-sdk-go/core/auth"
	"github.com/stackitcloud/stackit-sdk-go/core/config"
)

//nolint:all
func pull() {

	// Authentication with STACKIT SDK - Returns a Round-Tripper
	rt, err := auth.DefaultAuth(&config.Configuration{
		ServiceAccountKeyPath: "./service-account-key.json",
	})
	if err != nil {
		log.Printf("Error creating authentication token: %v", err)
	}

	// Setup your TopicID and Subscription ID
	//TODO: ID's needs to be replaced here
	topicID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	subscriptionID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Declare subscriber
	subscriber := pubsub.NewSubscriber(topicID,
		subscriptionID,
		pubsub.WithHTTPRoundTripper(rt),
	)

	// Pull messages via subscription
	pulledMessages, _ := subscriber.Pull(context.Background(), pubsub.WithMaxMessages(10))

	log.Printf("Successfully pulled message: %v", pulledMessages)

	// Get your AckIDs and acknowledge them
	ackIDs := pulledMessages.GetAckIDs()
	err = subscriber.Ack(context.Background(), ackIDs)

	// Get your NackIDs and not acknowledge them
	nackIDs := pulledMessages.GetAckIDs()
	err = subscriber.Nack(context.Background(), nackIDs)
}
