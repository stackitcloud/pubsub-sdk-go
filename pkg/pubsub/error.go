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
