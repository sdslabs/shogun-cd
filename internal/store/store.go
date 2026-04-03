package store

import (
	"errors"

	"github.com/kunalvirwal/shogun-cd/internal/config"
	"gorm.io/gorm"
)

type Store struct {
	User    UserStore
	Webhook WebhookStore
	Secret  SecretStore
}

var (
	ErrRecordNotFound = errors.New("Record Not Found")
)

func NewStore(cfg *config.Config, db *gorm.DB) *Store {
	return &Store{
		User:    newUserStore(db),
		Webhook: newWebhookStore(db),
		Secret:  newSecretStore(db),
	}
}
