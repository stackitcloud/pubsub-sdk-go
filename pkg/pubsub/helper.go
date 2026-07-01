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
