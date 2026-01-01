package webhooks

import (
	"sync"
	"time"

	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

const (
	WebhookNotFound = "Webhook Not Found"
	WebhookInactive = "Webhook Inactive"
)

type Webhook struct {
	ID        string    `json:"-"`    // db primary key
	Slug      string    `json:"slug"` // [TODO] indexed in db
	Secret    string    `json:"-"`    // HMAC secret
	Pipeline  string    `json:"pipeline_name"`
	Alias     string    `json:"alias"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	// LastUsedAt *time.Time `json:"last_used"` // not a part of in memory map - [TODO] update batched data in regular intervals to the DB
	IsActive bool `json:"is_active"`
}

// input DTO
type WebhookInput struct {
	Pipeline  string
	Alias     string
	CreatedBy string
}

type WebhookService interface {
	Create(input *WebhookInput) (*Webhook, error)
	Resolve(slug string) (*Webhook, error)
	Delete(slug string) error
	VerifyPayload(body []byte, header string, secret string) bool
	Find(filter func(*Webhook) bool) []*Webhook
}

type service struct {
	cache  map[string]*Webhook
	logger utils.Logger
	mu     sync.RWMutex
}

func NewWebhookService(l utils.Logger) WebhookService {
	// [TODO] Bulk read and sync the cache with DB
	svc := &service{
		cache:  make(map[string]*Webhook),
		logger: l,
	}
	return svc
}
