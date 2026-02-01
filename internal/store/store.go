package store

import (
	"gorm.io/gorm"
)

type Store struct {
	User    UserStore
	Webhook WebhookStore
}

func NewStore(db *gorm.DB) *Store {
	return &Store{
		User:    newUserStore(db),
		Webhook: newWebhookStore(db),
	}
}
