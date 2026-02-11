package controllers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/api/http/apiutils"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/webhooks"
)

func (h *Handler) HandleWebhook(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	slug := c.Param("slug")

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.response.ServerError(c, fmt.Errorf("Failed to Read Request Body: %v", err))
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	_, err = h.webhook.Resolve(ctx, slug, c.Request.Header, bodyBytes)
	if err != nil {
		switch {
		case errors.Is(err, webhooks.ErrBadWebhookHeader):
			h.response.BadRequest(c, err.Error(), err)

		case errors.Is(err, webhooks.ErrInvalidJSON):
			h.response.BadRequest(c, webhooks.ErrInvalidJSON.Error(), err)

		case errors.Is(err, webhooks.ErrAuthFailed):
			h.response.Unauthorized(c, err.Error(), err)

		case errors.Is(err, webhooks.ErrHookInactive):
			h.response.Forbidden(c, err.Error(), err)

		case errors.Is(err, webhooks.ErrHookNotFound):
			h.response.NotFound(c, err.Error(), err)

		default:
			h.response.ServerError(c, err)
		}
		return
	}

	h.response.Success(c, "OK", nil)
}

func (h *Handler) CreateWebhook(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req dto.HookInput

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	req.CreatedBy = apiutils.GetUserEmail(c)
	data, err := h.webhook.Create(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, webhooks.ErrInvalidPipeline):
			h.response.BadRequest(c, err.Error(), err)

		default:
			h.response.ServerError(c, err)
		}
		return
	}

	response := struct {
		Webhook *webhooks.Webhook `json:"webhook"`
		Secret  string            `json:"secret"`
	}{
		Webhook: data,
		Secret:  data.Secret,
	} // secret NEVER to be sent to the user anywhere else except after creation

	h.logger.LogInfo("New webhook created, pipeline: %v slug: %v", data.Pipeline, data.Slug)

	h.response.Created(c, fmt.Sprintf("Webhook Created for Pipeline - %v", req.Pipeline), response)
}

func (h *Handler) DeleteWebhook(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	slug := c.Param("slug")

	err := h.webhook.Delete(ctx, slug)
	if err != nil {
		h.response.ServerError(c, err)
		return
	}

	h.response.Success(c, fmt.Sprintf("Webhook - %v Deleted", slug), nil)
}

func (h *Handler) ListWebhooks(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	pipeline := c.Param("pipeline")
	var filter webhooks.WebhookFilter
	if pipeline == "" {
		filter = nil
	} else {
		filter = webhooks.FilterByPipeline(pipeline)
	}

	data := h.webhook.Find(ctx, filter)

	h.response.Success(c, "All Webhooks", data)
}

func (h *Handler) DeactivateWebhook(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	slug := c.Param("slug")

	err := h.webhook.SetStatus(ctx, slug, false)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			h.response.NotFound(c, "Webhook Not Found", err)
			return

		default:
			h.response.ServerError(c, err)
			return
		}
	}

	h.response.Success(c, "Hook Deactivated", nil)
}

func (h *Handler) ActivateWebhook(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	slug := c.Param("slug")

	err := h.webhook.SetStatus(ctx, slug, true)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			h.response.NotFound(c, "Webhook Not Found", err)
			return

		default:
			h.response.ServerError(c, err)
			return
		}
	}

	h.response.Success(c, "Hook Activated", nil)
}
