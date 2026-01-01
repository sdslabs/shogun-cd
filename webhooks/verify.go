package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// this HMAC verification works for github, which sends the hash in the header X-Hub-Signature-256
func (s *service) VerifyPayload(body []byte, header string, secret string) bool {
	if header == "" || secret == "" {
		return false
	}
	header = strings.ToLower(header)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	const prefix = "sha256="
	expected := prefix + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(header), []byte(expected))
}
