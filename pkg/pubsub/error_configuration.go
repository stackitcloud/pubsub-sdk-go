package pubsub

import "fmt"

// ConfigurationError indicates invalid client setup or credentials.
type ConfigurationError struct {
	Msg string
	Err error
}

func (e *ConfigurationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("pubsub configuration error: %s: %v", e.Msg, e.Err)
	}
	return fmt.Sprintf("pubsub configuration error: %s", e.Msg)
}

func (e *ConfigurationError) Unwrap() error     { return e.Err }
func (e *ConfigurationError) IsTransient() bool { return false }

func NewConfigurationError(msg string, err error) *ConfigurationError {
	return &ConfigurationError{
		Msg: msg,
		Err: err,
	}
}
