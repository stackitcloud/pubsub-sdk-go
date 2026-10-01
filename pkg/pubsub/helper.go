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
	"encoding/base64"
)

// bytesToBase64 encodes multiple raw byte slices into base64 byte slice.
func bytesToBase64(messages ...[]byte) [][]byte {
	result := make([][]byte, len(messages))
	for i, msg := range messages {
		dst := make([]byte, base64.StdEncoding.EncodedLen(len(msg)))
		base64.StdEncoding.Encode(dst, msg)
		result[i] = dst
	}
	return result
}

// base64Decode safely decodes a single base64-encoded byte slice.
func base64Decode(src []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst[:n], nil
}
