package store

import (
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"gorm.io/gorm"
)

type Store struct {
	User    UserStore
	Webhook WebhookStore
}

func NewStore(cfg *config.Config, db *gorm.DB) (*Store, error) {
	c, err := newCrypto(cfg)
	if err != nil {
		return nil, err
	}

	return &Store{
		User:    newUserStore(db),
		Webhook: newWebhookStore(db, c),
	}, nil
}
