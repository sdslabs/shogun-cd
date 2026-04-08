package controllers

import (
	"github.com/kunalvirwal/shogun-cd/api/http/response"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/secrets"
	"github.com/kunalvirwal/shogun-cd/internal/users"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"github.com/kunalvirwal/shogun-cd/internal/webhooks"
)

type Handler struct {
	logger   utils.Logger
	config   *config.Config
	response response.Responder
	webhook  webhooks.WebhookService
	user     users.Service
	secret   secrets.SecretService
}

func NewHandler(l utils.Logger, cfg *config.Config, responder response.Responder, wh webhooks.WebhookService, u users.Service, secrets secrets.SecretService) *Handler {
	return &Handler{
		logger:   l,
		config:   cfg,
		response: responder,
		webhook:  wh,
		user:     u,
		secret:   secrets,
	}
}
