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
