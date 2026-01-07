package webhooks

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/internal/orchestrator"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type WebhookService interface {
	Create(input *dto.HookInput) (*Webhook, error)
	Resolve(slug string, headers http.Header, body []byte) (*Webhook, error)
	Delete(slug string) error
	Find(filter WebhookFilter) []*Webhook
}

var (
	ErrInvalidJSON      = errors.New("Bad JSON Payload")
	ErrBadWebhookHeader = errors.New("No Supported Auth header found")
	ErrAuthFailed       = errors.New("Payload Authentication Failed")
	ErrHookInactive     = errors.New("Webhook Inactive")
	ErrHookNotFound     = errors.New("Webhook Not Found")
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
	logger       utils.Logger
	mu           sync.RWMutex
}

func NewWebhookService(l utils.Logger, o orchestrator.Orchestrator) WebhookService {
	// [TODO] Bulk read and sync the registry with DB
	svc := &Service{
		registry:     make(map[string]*Webhook),
		orchestrator: o,
		logger:       l,
	}
	return svc
}
