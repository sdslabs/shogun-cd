package controllers

import "github.com/kunalvirwal/shogun-cd/internal/utils"

type Handler struct {
	logger utils.Logger
}

func NewHandler(l utils.Logger) *Handler {
	return &Handler{
		logger: l,
	}
}
