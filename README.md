# STACKIT PubSub Go SDK

Welcome to the PubSub Dataplane SDK.
It provides a convenient way to interact with STACKIT PubSub topics and subscriptions for publishing and consuming messages.
This guide will walk you through setting up clients, authenticating your requests, and performing common operations like publishing and pulling messages.

## Prerequisites

Before you begin, you will need the following:

* Go `1.25` or later.
* A STACKIT project.
* A Service Account with the necessary permissions for PubSub.
* The ID of the topic you want to publish to and/or the subscription you want to pull from.

### 1. Installation

To use the SDK in your project, install it using `go get`:

```bash
go get github.com/stackitcloud/pubsub-sdk-go
```

### 2. Usage

To interact with the PubSub Dataplane, you need to create and configure a `Publisher` or a `Subscriber`.
You will find an [example Folder](./example) in the root directory, with ready to copy examples for using the SDK even easier.

The recommended way to handle authentication is by using a `RoundTripper` from the core STACKIT Go SDK, which automatically manages service account tokens.
For initiating the Roundtripper you need a service Account, this you can get in the STACKIT Portal.
To authenticate against the STACKIT CLI please take a look at the [Dokumentation]{https://github.com/stackitcloud/stackit-cli/blob/main/docs/stackit_auth_get-access-token.md}.

```go
roundTripper, err := auth.DefaultAuth(&config.Configuration{
    ServiceAccountKeyPath: "./service-account-key.json",
})
if err != nil {
    log.Printf("Error creating authentication token: %v", err)
}

publisher := pubsub.NewPublisher(topicID,
    pubsub.WithHTTPRoundTripper(roundTripper),
)
```

### 3. Logging Configuration

The SDK supports structured logging to help you debug and monitor your application. You can inject a custom logger using either the Go standard library's `slog` or the `logr` interface.

By default, the SDK outputs standard operational infos and errors. Enable debug-level logging for more granular troubleshooting.

**Using `slog`:**

```go
import (
    "log/slog"
    "os"
)

// Configure an slog logger for debug output
opts := &slog.HandlerOptions{
    Level: slog.LevelDebug,
}
slogLogger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

// Pass the slog.Logger to the publisher or subscriber
topicID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
publisher := pubsub.NewPublisher(topicID,
    pubsub.WithHTTPRoundTripper(roundTripper),
    pubsub.WithLogger(slogLogger),
)
```

> **Hint:** If your application already utilizes `logr` (for instance, via a backend like `zap`), you can skip `slog` and pass your `logr.Logger` instance directly to the configuration by using `pubsub.WithLogrLogger(yourLogrInstance)`.

### 4. Using Methods

Once you have configured your clients, you can send messages to a topic or consume them from a subscription.
To consume messages, you Pull them, process them, and `Ack` (acknowledge) them.

```go
// Publish
topicID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
publisher := pubsub.NewPublisher(topicID, pubsub.WithHTTPRoundTripper(roundTripper))

messages := [][]byte{
    []byte("Hello, PubSub!"),
}
messageIDs, err := publisher.Publish(ctx, messages)

// Pull
subscriptionID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
subscriber := pubsub.NewSubscriber(topicID, subscriptionID, pubsub.WithHTTPRoundTripper(roundTripper))

pulledMessages, err := subscriber.Pull(ctx, pubsub.WithMaxMessages(10))

// Acknowledge
ackIDs := pulledMessages.GetAckIDs()
err = subscriber.Ack(ctx, ackIDs)
```

### 5. Error Handling

The SDK returns specific error types to help you handle different failure scenarios:
* `pubsub.APIError`: Represents an error returned by the PubSub API (e.g., 404 Not Found if a topic doesn't exist).
* `pubsub.NetworkError`: Represents a client-side network issue (e.g., a timeout).

```go
// Example of detailed error handling
_, err := publisher.Publish(ctx, messages)
if err != nil {
    var apiErr *pubsub.APIError
    if errors.As(err, &apiErr) {
        log.Printf("API Error [%d]: %s (Code: %s)\n",
            apiErr.StatusCode, apiErr.Msg, apiErr.Code)

        if apiErr.IsNotFound() {
            log.Println("Action required: The specified topic or subscription does not exist.")
        }
        return
    }

    var netErr *pubsub.NetworkError
    if errors.As(err, &netErr) {
        log.Printf("Network Error: %v\n", netErr.Unwrap())

        if netErr.IsTransient() {
            log.Println("This is a temporary issue. It is safe to trigger a retry.")
        }
        return
    }

    log.Fatalf("An unexpected error occurred: %v", err)
}

```