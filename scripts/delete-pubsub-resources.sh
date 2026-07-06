#!/bin/bash

# Configuration
BASE_URL="https://pubsub.api.qa.stackit.cloud/v1alpha"
PROJECT_ID="0ec85b07-ecb2-4253-9ba8-25ae06db1b7a"
REGION="eu01"

if [ -z "$TOPIC_ID" ]; then
    echo "No TOPIC_ID found, skipping cleanup."
    exit 0
fi

# We use force=true query parameter on the topic deletion to perform a cascading delete of the topic
# and all of its active subscriptions.
echo "Cleaning up Topic: $TOPIC_ID with force=true"
DELETE_RESPONSE=$(curl -sk -w "\n%{http_code}" -X DELETE \
  --url "${BASE_URL}/projects/${PROJECT_ID}/regions/${REGION}/topics/${TOPIC_ID}?force=true" \
  --header "Authorization: Bearer ${STACKIT_SERVICE_ACCOUNT_TOKEN}")

HTTP_STATUS=$(echo "$DELETE_RESPONSE" | tail -n 1)
DELETE_BODY=$(echo "$DELETE_RESPONSE" | sed '$d')

if [ "$HTTP_STATUS" -ne 202 ]; then
  echo "::warning file=scripts/delete-pubsub-resources.sh::Failed to delete topic $TOPIC_ID (HTTP $HTTP_STATUS) - Response: $DELETE_BODY"
else
  echo "Successfully deleted topic $TOPIC_ID (HTTP $HTTP_STATUS)"
fi
