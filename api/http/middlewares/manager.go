package middlewares

import (
	"github.com/kunalvirwal/shogun-cd/api/http/response"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"github.com/kunalvirwal/shogun-cd/internal/webhooks"
)

type Manager struct {
	logger   utils.Logger
	config   config.API
	response response.Responder
	webhook  webhooks.WebhookService
}

func NewManager(l utils.Logger, cfg config.API, responder response.Responder, wh webhooks.WebhookService) *Manager {
	return &Manager{
		logger:   l,
		config:   cfg,
		response: responder,
		webhook:  wh,
	}
}
