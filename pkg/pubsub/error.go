package pubsub

type SDKError interface {
	// Error returns a string representation of the error.
	Error() string
	// Unwrap returns the underlying cause of the error, if any.
	Unwrap() error
	// IsTransient returns true if the error is temporary and can be retried
	IsTransient() bool
}

// Compile time assertions to ensure that SDKError implements the required interfaces.
var (
	_ SDKError = (*APIError)(nil)
	_ SDKError = (*ConfigurationError)(nil)
	_ SDKError = (*NetworkError)(nil)
)
