package example

import (
	"context"
	"encoding/base64"
	"log"

	"dev.azure.com/schwarzit-wiking/schwarzit.stackit-pubsub/stackit-pubsub-go-sdk.git/pkg/pubsub"
	"github.com/google/uuid"
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
	message := []byte("Hello PubSub from example")
	encodedMessage := base64.StdEncoding.EncodeToString(message)
	messagesToPublish := [][]byte{
		[]byte(encodedMessage),
	}

	// Publish the messages to the topic using the publisher client
	_, err = publisher.Publish(
		context.Background(),
		messagesToPublish,
	)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
	}

	log.Print("Successfully published messages")
}
