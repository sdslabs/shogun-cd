package webhooks

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

const (
	prefix        = "wh_"
	SlugEntropy   = 8
	SecretEntropy = 32
)

func (s *service) Create(input *WebhookInput) (*Webhook, error) {
	var slug string
	var secret string
	var err error

	secret, err = generateSecretHex(SecretEntropy)
	if err != nil {
		return nil, err
	}

	for {
		temp, _ := generateSecretHex(SlugEntropy)
		potentialSlug := prefix + temp

		s.mu.Lock()
		if _, exists := s.cache[potentialSlug]; !exists {
			slug = potentialSlug

			hook := &Webhook{
				Slug:      slug,
				Secret:    secret,
				Pipeline:  input.Pipeline,
				Alias:     input.Alias,
				CreatedBy: input.CreatedBy,
				IsActive:  true,
				CreatedAt: time.Now(),
			}

			s.cache[slug] = hook
			s.mu.Unlock()
			return hook, nil
		}

		s.mu.Unlock()
		s.logger.LogInfo("Slug collision, Retrying")
	}
}

func generateSecretHex(numByte int) (string, error) {
	bytes := make([]byte, numByte)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
