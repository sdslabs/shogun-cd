package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

const (
	GithubHMACHeader   = "X-Hub-Signature-256" // github uses hmac authentication on webhook payloads
	GitlabSecretHeader = "X-Gitlab-Token"      // gitlab uses simpler secret matching technique
)

type Provider string

const (
	Github Provider = "github"
	Gitlab Provider = "gitlab"
)

func (p Provider) ValidProvider() (Provider, bool) {
	valid := false
	switch p {
	case Github, Gitlab:
		valid = true
	}
	return p, valid
}

func (w *Webhook) VerifyPayload(headers http.Header, body []byte) (bool, error) {
	switch w.Provider {
	case Github:
		{
			val := headers.Get(GithubHMACHeader)
			if val == "" {
				return false, BadHeaderError
			}
			return HMACVerifier(body, val, w.Secret), nil
		}
	case Gitlab:
		{
			val := headers.Get(GitlabSecretHeader)
			if val == "" {
				return false, BadHeaderError
			}
			return gitlabVerifier(val, w.Secret), nil
		}
	default:
		{
			return false, InvalidProviderError
		}
	}
}

// Github hook payloads are verified using HMAC authentication
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

// Gitlab provides the secret in a header (no HMAC)
// [TODO] deprecate this as soon as gitlab releases HMAC support
func gitlabVerifier(header string, secret string) bool {
	return utils.SecureCompare(header, secret)
}
