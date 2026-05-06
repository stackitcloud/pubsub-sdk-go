# STACKIT PubSub Go SDK

Welcome to the PubSub Dataplane API SDK.
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
go get [github.com/stackitcloud/pubsub-sdk-go.git](https://github.com/stackitcloud/pubsub-sdk-go.git)
```

### 2. Usage

To interact with the PubSub API, you need to create and configure a `Publisher` or a `Subscriber`.
The recommended way to handle authentication is by using a `RoundTripper` from the core STACKIT Go SDK, which automatically manages service account tokens.

Both clients share the same configuration options. You will need to provide the configured `RoundTripper` and the Host (Environment) URI.

```go
topicID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

publisher := pubsub.NewPublisher(topicID,
    pubsub.WithHTTPRoundTripper(yourRoundTripper),
    pubsub.WithHost(host),
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
    pubsub.WithHTTPRoundTripper(yourRoundTripper),
    pubsub.WithHost(host),
    pubsub.WithLogger(slogLogger),
)
```

> **Hint:** If your application already utilizes `logr` (for instance, via a backend like `zap`), you can skip `slog` and pass your `logr.Logger` instance directly to the configuration by using `pubsub.WithLogrLogger(yourLogrInstance)`.

### 4. Using Methods

Once you have configured your clients, you can send messages to a topic or consume them from a subscription.
To consume messages, you Pull them, process them, and `Ack` (acknowledge) them to remove them from the subscription.

```go
// Publish
topicID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
publisher := pubsub.NewPublisher(topicID, pubsub.WithHTTPRoundTripper(yourRoundTripper))

messages := [][]byte{
    []byte("Hello, PubSub!"),
}
messageIDs, err := publisher.Publish(ctx, messages)

// Pull
subscriptionID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
subscriber := pubsub.NewSubscriber(topicID, subscriptionID, pubsub.WithHTTPRoundTripper(yourRoundTripper))

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

### 6. Codeowners

Codeowners of this PubSub Dataplane SDK are Kevin Sussbauer, Paul Grossmann, Selina Fehn, Ferdinand Linnenberg, and Christoph Schumacher.