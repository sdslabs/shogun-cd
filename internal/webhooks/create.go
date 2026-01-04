package webhooks

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/kunalvirwal/shogun-cd/api/dto"
)

const (
	prefix        = "shogun_"
	SlugEntropy   = 8
	SecretEntropy = 32
)

func (s *Service) Create(input *dto.HookInput) (*Webhook, error) {
	var secret string
	var err error

	provider, ok := Provider(input.Provider).ValidProvider()
	if !ok {
		return nil, InvalidProviderError
	}

	secret, err = generateSecretHex(SecretEntropy) // hex: composed of 0-9 and a-f
	if err != nil {
		return nil, err
	}

	hook := &Webhook{
		Slug:      "",
		Provider:  provider,
		Secret:    secret,
		Pipeline:  input.Pipeline,
		Alias:     input.Alias,
		CreatedBy: input.CreatedBy,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		temp, _ := generateSecretHex(SlugEntropy)
		newSlug := prefix + temp
		if _, exists := s.registry[newSlug]; !exists {
			hook.Slug = newSlug
			s.registry[newSlug] = hook
			return hook, nil
		}
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
