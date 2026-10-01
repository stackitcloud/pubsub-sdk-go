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
