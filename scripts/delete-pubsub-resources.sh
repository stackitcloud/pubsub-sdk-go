#!/bin/bash

# Configuration
BASE_URL="https://pubsub.api.eu01.qa/v1alpha"
PROJECT_ID="0ec85b07-ecb2-4253-9ba8-25ae06db1b7a"
REGION="eu01"

if [ -z "$TOPIC_ID" ]; then
    echo "No TOPIC_ID found, skipping cleanup."
    exit 0
fi

echo "Cleaning up Subscription: $SUBSCRIPTION_ID"
curl -s --request DELETE \
  --url "${BASE_URL}/projects/${PROJECT_ID}/regions/${REGION}/topics/${TOPIC_ID}/subscriptions/${SUBSCRIPTION_ID}" \
  --header "Authorization: Bearer ${STACKIT_SERVICE_ACCOUNT_TOKEN}"

echo "Cleaning up Topic: $TOPIC_ID"
curl -s --request DELETE \
  --url "${BASE_URL}/projects/${PROJECT_ID}/regions/${REGION}/topics/${TOPIC_ID}" \
  --header "Authorization: Bearer ${STACKIT_SERVICE_ACCOUNT_TOKEN}"

echo "Cleanup complete."