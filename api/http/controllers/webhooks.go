package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/api/http/apiutils"
	"github.com/kunalvirwal/shogun-cd/internal/webhooks"
)

func (h *Handler) HandleWebhook(c *gin.Context) {
	slug := c.Param("slug")

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.response.ServerError(c, fmt.Errorf("Failed to Read Request Body: %v", err))
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	hook, err := h.webhook.Resolve(slug, c.Request.Header, bodyBytes)
	if err != nil {
		switch {
		case errors.Is(err, webhooks.InvalidProviderError) || errors.Is(err, webhooks.BadHeaderError):
			h.response.BadRequest(c, err.Error(), err)

		case errors.Is(err, webhooks.AuthFailedError):
			h.response.Unauthorized(c, err.Error(), err)

		case errors.Is(err, webhooks.HookInactiveError):
			h.response.Forbidden(c, err.Error(), err)

		case errors.Is(err, webhooks.HookNotFoundError):
			h.response.NotFound(c, err.Error(), err)

		default:
			h.response.ServerError(c, err)
		}
		return
	}

	h.response.Success(c, "OK", hook)
}

func (h *Handler) CreateWebhook(c *gin.Context) {
	var req dto.HookInput

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	req.CreatedBy = apiutils.GetUserEmail(c)
	data, err := h.webhook.Create(&req)
	if err != nil {
		switch {
		case errors.Is(err, webhooks.InvalidProviderError):
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
	slug := c.Param("slug")

	err := h.webhook.Delete(slug)
	if err != nil {
		switch {
		case errors.Is(err, webhooks.HookNotFoundError):
			h.response.NotFound(c, err.Error(), err)
		default:
			h.response.ServerError(c, err)
		}
		return
	}

	h.response.Success(c, fmt.Sprintf("Webhook - %v Deleted", slug), nil)
}

func (h *Handler) ListWebhooks(c *gin.Context) {
	pipeline := c.Param("pipeline")
	var filter webhooks.WebhookFilter
	if pipeline == "" {
		filter = nil
	} else {
		filter = webhooks.FilterByPipeline(pipeline)
	}

	data := h.webhook.Find(filter)

	h.response.Success(c, "All Webhooks", data)
}
