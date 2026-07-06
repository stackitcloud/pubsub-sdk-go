#!/usr/bin/env bash
set -e

PROJECT_ID="0ec85b07-ecb2-4253-9ba8-25ae06db1b7a"
REGION="eu01"
BASE_URL="https://pubsub.api.qa.stackit.cloud/v1alpha"
PUBLISHER_MAIL="pubsub-dataplane-sdk-44cqm3i8@sa.stackit.cloud"

echo "Fetching fresh access token..."
TOKEN=$(stackit auth activate-service-account --only-print-access-token --service-account-key-path "$SA_KEY_PATH" | tr -d '\r\n ')
echo "SERVICE_ACCOUNT_TOKEN=$TOKEN" >> $GITHUB_ENV

if [ -z "$TOKEN" ] || [ ${#TOKEN} -lt 20 ]; then
  echo "Error: Retrieved token is empty or too short."
  exit 1
fi

#TOPIC
echo "Creating Topic via curl (Targeting: $REGION)..."
TOPICRESPONSE=$(curl -sk -w "\n%{http_code}" -X POST "${BASE_URL}/projects/${PROJECT_ID}/regions/${REGION}/topics" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"displayName\": \"ci-topic-$(date +%s)\"}")

HTTP_STATUS=$(echo "$TOPICRESPONSE" | tail -n 1)
TOPICBODY=$(echo "$TOPICRESPONSE" | sed '$d')
echo "Response Body: $TOPICBODY"

if [ "$HTTP_STATUS" -ne 202 ] && [ "$HTTP_STATUS" -ne 200 ]; then
  echo "API Error (HTTP $HTTP_STATUS)"
  echo "Response Topic Body: $TOPICBODY"
  exit 1
fi

TOPIC_ID=$(echo "$TOPICBODY" | jq -r '.id')
echo "Success! Topic ID: $TOPIC_ID"
echo "TOPIC_ID=$TOPIC_ID" >> $GITHUB_ENV

echo "Waiting for topic to become active..."
for i in {1..50}; do
  STATUS=$(curl -sk -H "Authorization: Bearer $TOKEN" "${BASE_URL}/projects/${PROJECT_ID}/regions/${REGION}/topics/${TOPIC_ID}" | jq -r '.state')
  echo "Current topic status: $STATUS"
  if [ "$STATUS" == "active" ]; then
    break
  fi
  if [ "$i" -eq 50 ]; then
    echo "Topic did not become active in time."
    exit 1
  fi
  sleep 5
done

#SUBSCRIPTION
echo "Creating Subscription via curl (Targeting: $REGION)..."
SUBRESPONSE=$(curl -sk -w "\n%{http_code}" -X POST "${BASE_URL}/projects/${PROJECT_ID}/regions/${REGION}/topics/${TOPIC_ID}/subscriptions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"displayName\": \"ci-topic-$(date +%s)\"}")

HTTP_STATUS=$(echo "$SUBRESPONSE" | tail -n 1)
SUBBODY=$(echo "$SUBRESPONSE" | sed '$d')
echo "Response Body: $SUBBODY"

if [ "$HTTP_STATUS" -ne 202 ] && [ "$HTTP_STATUS" -ne 200 ]; then
  echo "API Error (HTTP $HTTP_STATUS)"
  echo "Response SUBSCRIPTION Body: $SUBBODY"
  exit 1
fi

SUBSCRIPTION_ID=$(echo "$SUBBODY" | jq -r '.id')
echo "Success! Subscription ID: $SUBSCRIPTION_ID"
echo "SUBSCRIPTION_ID=$SUBSCRIPTION_ID" >> $GITHUB_ENV

echo "Waiting for subscription to become active..."
for i in {1..50}; do
  STATUS=$(curl -sk -H "Authorization: Bearer $TOKEN" "${BASE_URL}/projects/${PROJECT_ID}/regions/${REGION}/topics/${TOPIC_ID}/subscriptions/${SUBSCRIPTION_ID}" | jq -r '.state')
  echo "Current subscription status: $STATUS"
  if [ "$STATUS" == "active" ]; then
    break
  fi
  if [ "$i" -eq 50 ]; then
    echo "Subscription did not become active in time."
    exit 1
  fi
  sleep 5
done


#ACCESS
echo "Granting Publisher Access via curl (Targeting: $REGION)..."
PUBLISHER_RESPONSE=$(curl -sk -w "\n%{http_code}" -X PATCH "${BASE_URL}/projects/${PROJECT_ID}/regions/${REGION}/topics/${TOPIC_ID}/publishers" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"emailAddress\": \"$PUBLISHER_MAIL\"}")

HTTP_STATUS=$(echo "$PUBLISHER_RESPONSE" | tail -n 1)
PUBLISHER_BODY=$(echo "$PUBLISHER_RESPONSE" | sed '$d')

if [ "$HTTP_STATUS" -ne 202 ] && [ "$HTTP_STATUS" -ne 200 ]; then
  echo "API Error Granting Publisher Access (HTTP $HTTP_STATUS)"
  echo "Response Granting Publisher Access Body: $PUBLISHER_BODY"
  exit 1
fi

echo "Granting Subscriber Access via curl (Targeting: $REGION)..."
SUBSCRIBER_RESPONSE=$(curl -sk -w "\n%{http_code}" -X PATCH "${BASE_URL}/projects/${PROJECT_ID}/regions/${REGION}/topics/${TOPIC_ID}/subscriptions/${SUBSCRIPTION_ID}/subscribers" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"emailAddress\": \"$PUBLISHER_MAIL\"}")

HTTP_STATUS=$(echo "$SUBSCRIBER_RESPONSE" | tail -n 1)
SUBSCRIBER_BODY=$(echo "$SUBSCRIBER_RESPONSE" | sed '$d')

if [ "$HTTP_STATUS" -ne 202 ] && [ "$HTTP_STATUS" -ne 200 ]; then
  echo "API Error Granting Subscriber Access (HTTP $HTTP_STATUS)"
  echo "Response Granting Subscriber Access Body: $SUBSCRIBER_BODY"
  exit 1
fi

echo "Waiting for access permissions to propagate..."
sleep 5