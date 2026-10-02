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

package pubsub

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the StackIT Pub/Sub API.
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Code       string `json:"code"`
	Msg        string `json:"message"`
	Err        error  `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("pubsub api error (status: %d, code: %s): %s",
		e.StatusCode, e.Code, e.Msg)
}

func (e *APIError) Unwrap() error { return e.Err }

func (e *APIError) IsTransient() bool {
	// 429 Too Many Requests and 5xx Server Errors are typically retryable.
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

// IsNotFound returns true if the API returned a 404 (e.g., Topic or Subscription doesn't exist).
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsConflict returns true if the API returned a 409 (e.g., resource modified by another client).
func (e *APIError) IsConflict() bool {
	return e.StatusCode == http.StatusConflict
}

// IsUnauthorized returns true if the API returned a 401.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsBadRequest returns true if the API returned a 400.
func (e *APIError) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

// IsRequestEntityTooLarge returns true if the API returned a 413.
func (e *APIError) IsRequestEntityTooLarge() bool {
	return e.StatusCode == http.StatusRequestEntityTooLarge
}

// APIErrorOption allows for customizing the APIError during initialization.
type APIErrorOption func(*APIError)

// WithUnderlyingError attaches a root cause error to the APIError.
func WithUnderlyingError(err error) APIErrorOption {
	return func(e *APIError) {
		e.Err = err
	}
}

// NewAPIError constructs an APIError from the HTTP response details.
func NewAPIError(statusCode int, body []byte, opts ...APIErrorOption) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
	}

	// Apply any provided options
	for _, opt := range opts {
		opt(apiErr)
	}

	var payload struct {
		Message string `json:"message"`
		Error   string `json:"error,omitempty"`
	}

	err := json.Unmarshal(body, &payload)
	if err != nil || (payload.Message == "" && payload.Error == "") {
		apiErr.Msg = string(body)
		if apiErr.Msg == "" {
			apiErr.Msg = "api request failed with no response body"
		}
		apiErr.Code = "unknown_error"
		return apiErr
	}

	if payload.Message != "" {
		apiErr.Msg = payload.Message
	} else {
		apiErr.Msg = "api request failed"
	}

	if payload.Error != "" {
		apiErr.Code = payload.Error
	} else {
		apiErr.Code = "api_error"
	}

	return apiErr
}
