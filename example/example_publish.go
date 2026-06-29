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
		log.Fatalf("Error creating authentication token: %v", err)
	}

	// Setup your Topic ID
	//TODO: ID needs to be replaced here
	topicID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Declare publisher
	publisher := pubsub.NewPublisher(topicID,
		pubsub.WithHTTPRoundTripper(rt),
	)

	// Publish the messages to the topic using the publisher client
	_, err = publisher.PublishStrings(
		context.Background(),
		"Hello PubSub from example", "This is another message",
	)
	if err != nil {
		log.Fatalf("Error publishing message: %v", err)
	}

	log.Print("Successfully published messages")
}
