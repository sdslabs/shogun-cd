package store

import (
	"context"
	"errors"

	"github.com/kunalvirwal/shogun-cd/internal/models"

	"gorm.io/gorm"
)

type WebhookStore interface {
	Create(ctx context.Context, in *models.Webhook) error
	Delete(ctx context.Context, slug string) error
	SetStatus(ctx context.Context, slug string, isActive bool) error
	Load(ctx context.Context) ([]*models.Webhook, error)
}

var (
	ErrSlugCollision = errors.New("Slug Collision")
	ErrHookNotFound  = errors.New("Webhook Not Found")
	ErrEncryption    = errors.New("Error Encrypting Data")
	ErrDecryption    = errors.New("Error Decrypting Data")
)

type webhookStore struct {
	db     *gorm.DB
	crypto Crypto
}

func newWebhookStore(db *gorm.DB, c Crypto) WebhookStore {
	return &webhookStore{
		db:     db,
		crypto: c,
	}
}

func (w *webhookStore) Create(ctx context.Context, in *models.Webhook) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if in.Secret != "" {
		secret := in.Secret
		encrypted, err := w.crypto.Encrypt(secret)
		if err != nil {
			return ErrEncryption
		}
		in.Secret = encrypted
	}

	if err := gorm.G[models.Webhook](w.db).
		Create(ctx, in); err != nil {
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return ErrSlugCollision

		default:
			return err
		}
	}

	return nil
}

func (w *webhookStore) Delete(ctx context.Context, slug string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if _, err := gorm.G[models.Webhook](w.db).
		Where(&models.Webhook{
			Slug: slug,
		}).
		Delete(ctx); err != nil {
		return err
	}

	return nil
}

func (w *webhookStore) SetStatus(ctx context.Context, slug string, isActive bool) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	rows, err := gorm.G[models.Webhook](w.db).
		Where(&models.Webhook{
			Slug: slug,
		}).
		Updates(ctx, models.Webhook{
			IsActive: &isActive,
		})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrHookNotFound
	}

	return nil
}

func (w *webhookStore) Load(ctx context.Context) ([]*models.Webhook, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	hooks, err := gorm.G[*models.Webhook](w.db).Find(ctx)
	if err != nil {
		return nil, err
	}

	var validHooks []*models.Webhook

	for _, h := range hooks {
		if h.Secret == "" {
			continue
		}
		decrypted, err := w.crypto.Decrypt(h.Secret)
		if err != nil {
			continue
		}
		h.Secret = decrypted
		validHooks = append(validHooks, h)
	}

	return validHooks, nil
}
