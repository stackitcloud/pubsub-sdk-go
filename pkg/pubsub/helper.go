package pubsub

import "encoding/base64"

func StringsToBase64(strings ...string) [][]byte {
	result := make([][]byte, len(strings))

	for i, string := range strings {
		encodedStrings := base64.StdEncoding.EncodeToString([]byte(string))
		result[i] = []byte(encodedStrings)
	}
	return result
}
