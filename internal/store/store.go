package store

import (
	"errors"
	"gorm.io/gorm"
)

type Store struct {
	User     UserStore
	Webhook  WebhookStore
	Secret   SecretStore
	Pipeline PipelineStore
}

var (
	ErrRecordNotFound = errors.New("record not found")
)

func NewStore(db *gorm.DB) *Store {
	return &Store{
		User:     newUserStore(db),
		Webhook:  newWebhookStore(db),
		Secret:   newSecretStore(db),
		Pipeline: newPipelineStore(db),
	}
}
