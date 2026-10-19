package webhooks

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/kunalvirwal/shogun-cd/internal/encryption"
	"github.com/kunalvirwal/shogun-cd/internal/orchestrator"
	"github.com/kunalvirwal/shogun-cd/internal/store"

	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type WebhookService interface {
	Create(ctx context.Context, input *CreateParams) (*Webhook, error)
	Resolve(ctx context.Context, slug string, headers http.Header, body []byte) (*Webhook, error)
	Delete(ctx context.Context, slug string) error
	Find(ctx context.Context, filter WebhookFilter) []*Webhook
	SetStatus(ctx context.Context, slug string, isActive bool) error
	Load() error
}

var (
	ErrInvalidJSON      = errors.New("bad json payload")
	ErrBadWebhookHeader = errors.New("no supported auth header found")
	ErrInvalidPipeline  = errors.New("pipeline does not exist")
	ErrAuthFailed       = errors.New("payload authentication failed")
	ErrHookInactive     = errors.New("webhook inactive")
	ErrHookNotFound     = errors.New("webhook not found")
)

type Webhook struct {
	ID        string    `json:"-"`    // db primary key
	Slug      string    `json:"slug"` // [TODO] indexed in db
	Secret    string    `json:"-"`    // hook secret
	Pipeline  string    `json:"pipeline_name"`
	Alias     string    `json:"alias"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	// LastUsedAt *time.Time `json:"last_used"` // not a part of in memory map - [TODO] update batched data in regular intervals to the DB
	IsActive bool `json:"is_active"`
}

type Service struct {
	registry     map[string]*Webhook
	orchestrator orchestrator.Orchestrator
	store        store.WebhookStore
	encryption   encryption.EncryptionService
	logger       utils.Logger
	mu           sync.RWMutex
}

func NewWebhookService(l utils.Logger, o orchestrator.Orchestrator, s *store.Store, e encryption.EncryptionService) WebhookService {
	svc := &Service{
		registry:     make(map[string]*Webhook),
		orchestrator: o,
		store:        s.Webhook,
		logger:       l,
		encryption:   e,
	}
	return svc
}
