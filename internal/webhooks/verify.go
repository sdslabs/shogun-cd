package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	ShogunSecretHeader = "X-Shogun-Token" // contains the actual secret (for requests made using curl command)
	ShogunHMACHeader   = "X-Shogun-HMAC"  // contains the HMAC signature (for cli requests)
)

func (w *Webhook) VerifyPayload(headers http.Header, body []byte) (bool, error) {

	if token := headers.Get(ShogunSecretHeader); token != "" {
		return subtle.ConstantTimeCompare([]byte(token), []byte(w.Secret)) == 1, nil
	}

	if signature := headers.Get(ShogunHMACHeader); signature != "" {
		return HMACVerifier(body, signature, w.Secret), nil
	}

	return false, ErrBadWebhookHeader
}

func HMACVerifier(body []byte, signature string, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}
	signature = strings.ToLower(signature)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	const prefix = "sha256="
	expected := prefix + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}
