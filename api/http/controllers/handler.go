package controllers

import (
	"github.com/kunalvirwal/shogun-cd/api/http/response"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"github.com/kunalvirwal/shogun-cd/internal/webhooks"
)

type Handler struct {
	logger   utils.Logger
	config   *config.Config
	response response.Responder
	webhook  webhooks.WebhookService
	store    *store.Store
}

func NewHandler(l utils.Logger, cfg *config.Config, responder response.Responder, wh webhooks.WebhookService, store *store.Store) *Handler {
	return &Handler{
		logger:   l,
		config:   cfg,
		response: responder,
		webhook:  wh,
		store:    store,
	}
}
