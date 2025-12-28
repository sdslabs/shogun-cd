package utils

import "crypto/subtle"

func SecureCompare(given, actual string) bool {
	return subtle.ConstantTimeCompare([]byte(given), []byte(actual)) == 1
}
