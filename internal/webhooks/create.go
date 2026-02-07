package webhooks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/internal/models"
	"github.com/kunalvirwal/shogun-cd/internal/store"
)

const (
	prefix        = "shogun_"
	SlugEntropy   = 8
	SecretEntropy = 32
)

func (s *Service) Create(ctx context.Context, input *dto.HookInput) (*Webhook, error) {
	var secret string
	var err error

	if !s.orchestrator.PipelineExists(input.Pipeline) {
		return nil, ErrInvalidPipeline
	}

	secret, err = generateSecretHex(SecretEntropy) // hex: composed of 0-9 and a-f
	if err != nil {
		return nil, err
	}

	//[TODO] enforce input field lengths for webhook using validator tags

	encrypted, err := s.encryption.Encrypt(secret)
	if err != nil {
		return nil, err
	}

	dbHook := &models.Webhook{
		Secret:    encrypted,
		Pipeline:  input.Pipeline,
		Alias:     input.Alias,
		CreatedBy: input.CreatedBy,
	}
	hook := &Webhook{
		Secret:    secret,
		Pipeline:  dbHook.Pipeline,
		Alias:     dbHook.Alias,
		CreatedBy: dbHook.CreatedBy,
	}

	for {
		temp, _ := generateSecretHex(SlugEntropy)
		newSlug := prefix + temp

		dbHook.Slug = newSlug
		err := s.store.Create(ctx, dbHook)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrSlugCollision):
				s.logger.LogError(store.ErrSlugCollision)
				continue
			default:
				return nil, err
			}
		}

		hook.ID = dbHook.ID
		hook.Slug = dbHook.Slug
		hook.CreatedAt = dbHook.CreatedAt
		hook.IsActive = *dbHook.IsActive

		s.mu.Lock()
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
