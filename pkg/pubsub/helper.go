package pubsub

import (
	"encoding/base64"
	"fmt"
)

func stringsToBase64(strings ...string) [][]byte {
	result := make([][]byte, len(strings))

	for i, string := range strings {
		encodedStrings := base64.StdEncoding.EncodeToString([]byte(string))
		result[i] = []byte(encodedStrings)
	}
	return result
}

func base64ToStrings(encodedStrings ...[]byte) ([]string, error) {
	result := make([]string, len(encodedStrings))

	for i, s := range encodedStrings {
		decodedBytes := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
		n, err := base64.StdEncoding.Decode(decodedBytes, s)
		if err != nil {
			return nil, fmt.Errorf("failed to decode string at index %d: %w", i, err)
		}
		result[i] = string(decodedBytes[:n])
	}

	return result, nil
}
