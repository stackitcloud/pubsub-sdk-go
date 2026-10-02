// Copyright 2026 Schwarz Digits Cloud GmbH & Co. KG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
func purge() {
	// Authentication with STACKIT SDK - Returns a Round-Tripper
	rt, err := auth.DefaultAuth(&config.Configuration{
		ServiceAccountKeyPath: "./service-account-key.json",
	})
	if err != nil {
		log.Fatalf("Error creating authentication token: %v", err)
	}

	// Setup your TopicID and Subscription ID
	//TODO: ID needs to be replaced here
	topicID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Declare publisher
	publisher := pubsub.NewPublisher(topicID,
		pubsub.WithHTTPRoundTripper(rt),
	)

	err = publisher.Purge(context.Background())
	if err != nil {
		log.Fatalf("Error purging topic: %v", err)
	}
}
