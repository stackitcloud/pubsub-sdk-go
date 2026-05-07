package pubsub

import "fmt"

// NetworkError indicates a failure to reach the StackIT API (e.g., DNS resolution, timeouts).
type NetworkError struct {
	Msg string
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("pubsub network error: %s: %v", e.Msg, e.Err)
}

func (e *NetworkError) Unwrap() error     { return e.Err }
func (e *NetworkError) IsTransient() bool { return true } // Network errors are generally retryable

func NewNetworkError(msg string, err error) *NetworkError {
	return &NetworkError{
		Msg: msg,
		Err: err,
	}
}
