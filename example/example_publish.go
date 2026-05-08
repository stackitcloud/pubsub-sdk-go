package example

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/stackitcloud/pubsub-sdk-go/pkg/pubsub"
	"github.com/stackitcloud/stackit-sdk-go/core/auth"
	"github.com/stackitcloud/stackit-sdk-go/core/config"
)

//nolint:all
func publish() {

	// Authentication with STACKIT SDK - Returns a Round-Tripper
	rt, err := auth.DefaultAuth(&config.Configuration{
		ServiceAccountKeyPath: "./service-account-key.json",
	})
	if err != nil {
		log.Printf("Error creating authentication token: %v", err)
	}

	// Setup your Topic ID
	//TODO: ID needs to be replaced here
	topicID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Declare publisher
	publisher := pubsub.NewPublisher(topicID,
		pubsub.WithHTTPRoundTripper(rt),
	)

	// Create a message to publish to the topic and encode it to base64 format
	message := pubsub.StringsToBase64("Hello PubSub from example", "This is another message")

	// Publish the messages to the topic using the publisher client
	_, err = publisher.Publish(
		context.Background(),
		message,
	)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
	}

	log.Print("Successfully published messages")
}
