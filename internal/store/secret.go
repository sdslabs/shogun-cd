package store

import (
	"context"

	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SecretStore interface {
	Create(ctx context.Context, secrets []models.Secret) error
	Upsert(ctx context.Context, secret []models.Secret) error
	Update(ctx context.Context, secret *models.Secret) error
	FindMany(ctx context.Context, filter *dto.SecretFilter) ([]*dto.Secret, error)
	GetSecret(ctx context.Context, name string) (string, error)
	Delete(ctx context.Context, name string) error
}

type secretStore struct {
	db *gorm.DB
}

func newSecretStore(db *gorm.DB) SecretStore {
	return &secretStore{
		db: db,
	}
}

func (s *secretStore) Create(ctx context.Context, secrets []models.Secret) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return gorm.G[[]models.Secret](s.db).
		Table(models.SecretTableName).
		Create(ctx, &secrets)
}

func (s *secretStore) Upsert(ctx context.Context, secret []models.Secret) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return gorm.G[[]models.Secret](s.db, clause.OnConflict{
		Columns: []clause.Column{{
			Name: models.SecretColName,
		}},
		DoUpdates: clause.AssignmentColumns([]string{
			models.SecretColValue,
			models.SecretColUpdatedAt,
		}),
	}).
		Table(models.SecretTableName).
		Create(ctx, &secret)
}

func (s *secretStore) Update(ctx context.Context, secret *models.Secret) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	r, err := gorm.G[dto.SecretInput](s.db).
		Table(models.SecretTableName).
		Where(&models.Secret{
			Name: secret.Name,
		}).
		Update(ctx, models.SecretColValue, secret.Value)

	if r == 0 {
		return ErrRecordNotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *secretStore) FindMany(ctx context.Context, filter *dto.SecretFilter) ([]*dto.Secret, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	secrets, err := gorm.G[*dto.Secret](s.db).
		Table(models.SecretTableName).
		Select(models.SecretColName).
		Where(filter).
		Find(ctx)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

// returns the value of secret as in the db
func (s *secretStore) GetSecret(ctx context.Context, name string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	secret, err := gorm.G[*models.Secret](s.db).
		Table(models.SecretTableName).
		Select(models.SecretColValue).
		Where(&models.Secret{
			Name: name,
		}).
		First(ctx)
	if err != nil {
		return "", err
	}

	return secret.Value, nil
}

func (s *secretStore) Delete(ctx context.Context, name string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	_, err := gorm.G[models.Secret](s.db).
		Where(&models.Secret{
			Name: name,
		}).
		Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}
