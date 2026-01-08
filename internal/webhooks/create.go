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

	if !s.orchestrator.PipelineExists(input.Pipeline) {
		return nil, ErrInvalidPipeline
	}

	secret, err = generateSecretHex(SecretEntropy) // hex: composed of 0-9 and a-f
	if err != nil {
		return nil, err
	}

	hook := &Webhook{
		Slug:      "",
		Secret:    secret,
		Pipeline:  input.Pipeline,
		Alias:     input.Alias,
		CreatedBy: input.CreatedBy,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	for {
		temp, _ := generateSecretHex(SlugEntropy)
		newSlug := prefix + temp

		s.mu.Lock()
		if _, exists := s.registry[newSlug]; exists {
			s.mu.Unlock()
			s.logger.LogInfo("Slug collision, Retrying")
			continue
		}
		hook.Slug = newSlug
		s.registry[newSlug] = hook

		s.mu.Unlock()
		return hook, nil
	}
}

func generateSecretHex(numByte int) (string, error) {
	bytes := make([]byte, numByte)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
