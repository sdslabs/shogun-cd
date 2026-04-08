package secrets

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kunalvirwal/shogun-cd/internal/encryption"
	"github.com/kunalvirwal/shogun-cd/internal/models"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type SecretService interface {
	SetSecrets(ctx context.Context, in []CreateParams) error
	FindMany(ctx context.Context, filter *FilterParams) ([]Secret, error)
	FetchSecret(ctx context.Context, name string) (string, error)
	ResolveSecrets(ctx context.Context, name string) (string, error) // provides decrypted "value" of the secret
	DeleteSecret(ctx context.Context, name string) error
}

type Secret struct {
	Name  string `json:"name"`
	Value string `json:"-"`
}

type service struct {
	encryption encryption.EncryptionService
	store      store.SecretStore
	logger     utils.Logger
}

func NewSecretService(e encryption.EncryptionService, s *store.Store, l utils.Logger) SecretService {
	return &service{
		encryption: e,
		store:      s.Secret,
		logger:     l,
	}
}

// updates the value of a secret, creates new if no such secret exists
func (s *service) SetSecrets(ctx context.Context, in []CreateParams) error {
	secrets := make([]models.Secret, len(in))

	for k, v := range in {
		encrypted, err := s.encryption.Encrypt(v.Value)
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

func (s *service) FindMany(ctx context.Context, filter *FilterParams) ([]Secret, error) {
	if filter == nil {
		filter = &FilterParams{}
	}
	data, err := s.store.FindMany(ctx, &models.Secret{
		Name: filter.Name,
	})
	if err != nil {
		return nil, err
	}

	secretArr := make([]Secret, len(data))
	for k, v := range data {
		secretArr[k] = Secret{
			Name:  v.Name,
			Value: v.Value,
		}
	}

	return secretArr, nil
}

func (s *service) FetchSecret(ctx context.Context, name string) (string, error) {
	secret, err := s.store.GetSecret(ctx, name)
	if err != nil {
		return "", err
	}

	decrypted, err := s.encryption.Decrypt(secret)
	if err != nil {
		return "", fmt.Errorf("%w : %w", encryption.ErrDecryption, err)
	}

	return decrypted, nil
}

// takes a string as input and replaces the {{SECRET_NAME}} with its decrypted value as per the database.
func (s *service) ResolveSecrets(ctx context.Context, input string) (string, error) {
	var secretRegex = regexp.MustCompile(`\{\{([A-Za-z_][A-Za-z0-9_]*)\}\}`)

	matches := secretRegex.FindAllStringSubmatch(input, -1)

	output := input

	for _, match := range matches {
		secretTag := match[0]
		secretName := match[1]

		decrypted, err := s.FetchSecret(ctx, secretName)
		if err != nil {
			return "", err
		}

		output = strings.ReplaceAll(output, secretTag, decrypted)
	}

	return output, nil
}

func (s *service) DeleteSecret(ctx context.Context, name string) error {
	return s.store.Delete(ctx, name)
}
