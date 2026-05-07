package example

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/stackitcloud/pubsub-sdk-go.git/pkg/pubsub"
	"github.com/stackitcloud/stackit-sdk-go/core/auth"
	"github.com/stackitcloud/stackit-sdk-go/core/config"
)

func publish() {

	// Round Tripper Declaration
	rt, err := auth.DefaultAuth(&config.Configuration{
		ServiceAccountKey: "./service-account-key.json",
		TokenCustomUrl:    "https://service-account.api.stackit.cloud/token",
	})
	if err != nil {
		log.Printf("Error creating authentication token: %v", err)
	}

	//setup your Topic ID
	topicID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Declare publisher
	publisher := pubsub.NewPublisher(topicID,
		pubsub.WithHTTPRoundTripper(rt),
		pubsub.WithHost("pubsub.eu01.onstackit.cloud"),
	)

	// Create a message to publish to the topic and encode it to base64 format
	message := pubsub.ConvertToBase64("Hello PubSub from example")

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
