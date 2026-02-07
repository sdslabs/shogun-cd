package secrets

import (
	"context"

	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/internal/encryption"
	"github.com/kunalvirwal/shogun-cd/internal/models"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type Service interface {
	Create(ctx context.Context, in []dto.SecretInput) error
	Upsert(ctx context.Context, in []dto.SecretInput) error
	Update(ctx context.Context, in *dto.SecretInput) error
	Resolve(ctx context.Context, name string) (string, error) // provides decrypted "value" of the secret
	FindMany(ctx context.Context, filter *dto.SecretFilter) ([]*dto.Secret, error)
	Delete(ctx context.Context, name string) error
}

type service struct {
	encryption encryption.Service
	store      store.SecretStore
	logger     utils.Logger
}

func NewSecretService(e encryption.Service, s store.Store, l utils.Logger) Service {
	return &service{
		encryption: e,
		store:      s.Secret,
		logger:     l,
	}
}

func (s *service) Create(ctx context.Context, in []dto.SecretInput) error {
	secrets := make([]models.Secret, len(in))

	for k, v := range in {
		encrypted, err := s.encryption.Encrypt(in[k].Value)
		if err != nil {
			return err
		}
		secrets[k] = models.Secret{
			Name:  v.Name,
			Value: encrypted,
		}
	}

	return s.store.Create(ctx, secrets)
}

func (s *service) Upsert(ctx context.Context, in []dto.SecretInput) error {
	secrets := make([]models.Secret, len(in))

	for k, v := range in {
		encrypted, err := s.encryption.Encrypt(in[k].Value)
		if err != nil {
			return err
		}
		secrets[k] = models.Secret{
			Name:  v.Name,
			Value: encrypted,
		}
	}

	return s.store.Upsert(ctx, secrets)
}

func (s *service) Update(ctx context.Context, in *dto.SecretInput) error {
	encrypted, err := s.encryption.Encrypt(in.Value)
	if err != nil {
		return err
	}

	return s.store.Update(ctx, &models.Secret{
		Name:  in.Name,
		Value: encrypted,
	})
}

func (s *service) Resolve(ctx context.Context, name string) (string, error) {
	secrets, err := s.store.GetSecret(ctx, name)
	if err != nil {
		return "", err
	}

	decrypted, err := s.encryption.Decrypt(secrets)
	if err != nil {
		return "", err
	}

	return decrypted, nil
}

func (s *service) FindMany(ctx context.Context, filter *dto.SecretFilter) ([]*dto.Secret, error) {
	return s.store.FindMany(ctx, filter)
}

func (s *service) Delete(ctx context.Context, name string) error {
	return s.store.Delete(ctx, name)
}
