package example

import (
	"context"
	"log"

	"dev.azure.com/schwarzit-wiking/schwarzit.stackit-pubsub/stackit-pubsub-go-sdk.git/pkg/pubsub"
	"github.com/google/uuid"
	"github.com/stackitcloud/stackit-sdk-go/core/auth"
	"github.com/stackitcloud/stackit-sdk-go/core/config"
)

func purge() {
	// Round Tripper Declaration
	rt, err := auth.DefaultAuth(&config.Configuration{
		ServiceAccountKey: "./service-account-key.json",
		TokenCustomUrl:    "https://service-account.api.stackit.cloud/token",
	})
	if err != nil {
		log.Printf("Error creating authentication token: %v", err)
	}

	//setup your TopicID and Subscription ID
	topicID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Declare publisher
	publisher := pubsub.NewPublisher(topicID,
		pubsub.WithHTTPRoundTripper(rt),
		pubsub.WithHost("pubsub.eu01.onstackit.cloud"),
	)

	err = publisher.Purge(context.Background())

}
