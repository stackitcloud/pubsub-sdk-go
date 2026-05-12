package pubsub

import (
	"encoding/base64"
	"fmt"
)

func StringsToBase64(strings ...string) [][]byte {
	result := make([][]byte, len(strings))

	for i, string := range strings {
		encodedStrings := base64.StdEncoding.EncodeToString([]byte(string))
		result[i] = []byte(encodedStrings)
	}
	return result
}

func Base64ToStrings(encodedStrings ...string) ([]string, error) {
	result := make([]string, len(encodedStrings))

	for i, s := range encodedStrings {
		decodedBytes, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("failed to decode string at index %d: %w", i, err)
		}
		result[i] = string(decodedBytes)
	}

	return result, nil
}
