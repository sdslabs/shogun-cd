package middlewares

import (
	"github.com/kunalvirwal/shogun-cd/api/http/response"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type Manager struct {
	logger   utils.Logger
	config   *config.Config
	response response.Responder
}

func NewManager(l utils.Logger, cfg *config.Config, responder response.Responder) *Manager {
	return &Manager{
		logger:   l,
		config:   cfg,
		response: responder,
	}
}
