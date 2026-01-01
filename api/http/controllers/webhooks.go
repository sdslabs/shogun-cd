package controllers

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/request"
	"github.com/kunalvirwal/shogun-cd/webhooks"
)

const (
	HMACHeader = "X-Hub-Signature-256"
)

func (h *Handler) HandleWebhook(c *gin.Context) {
	slug := c.Param("slug")

	hook, err := h.webhook.Resolve(slug)
	if err != nil {
		h.response.NotFound(c, "Unable to Resolve Webhook", err)
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.response.ServerError(c, fmt.Errorf("Failed to Read Request Body: %v", err))
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	signature := c.GetHeader(HMACHeader)
	if signature == "" {
		h.response.BadRequest(c, "Bad Request", fmt.Errorf("HMAC header required"))
		return
	}

	if !h.webhook.VerifyPayload(bodyBytes, signature, hook.Secret) {
		h.response.Forbidden(c, "Unauthorised", fmt.Errorf("Failed to Authenticate the Payload (HMAC mismatch)"))
		return
	}

	h.response.Success(c, fmt.Sprintf("token received - %v", slug), hook)
}

func (h *Handler) CreateWebhook(c *gin.Context) {
	pipeline := c.Param("pipeline")
	email := request.GetUserEmail(c)

	data, err := h.webhook.Create(&webhooks.WebhookInput{
		Pipeline:  pipeline,
		Alias:     "", // usage upto discussion
		CreatedBy: email,
	})
	if err != nil {
		h.response.ServerError(c, err)
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

	h.response.Created(c, fmt.Sprintf("Webhook Created for Pipeline - %v", pipeline), response)
}

func (h *Handler) DeleteWebhook(c *gin.Context) {
	slug := c.Param("slug")

	err := h.webhook.Delete(slug)
	if err != nil {
		if err.Error() == webhooks.WebhookNotFound {
			h.response.NotFound(c, err.Error(), nil)
		} else {
			h.response.ServerError(c, err)
		}
		return
	}

	h.response.Success(c, fmt.Sprintf("Webhook - %v Deleted", slug), nil)
}

func (h *Handler) ListAllWebhooks(c *gin.Context) {
	data := h.webhook.Find(nil)

	h.response.Success(c, "All Webhooks", data)
}

func (h *Handler) ListPipelineWebhooks(c *gin.Context) {
	p := c.Param("pipeline")

	data := h.webhook.Find(webhooks.FilterByPipeline(p))

	h.response.Success(c, fmt.Sprintf("All Webhooks for Pipeline - %v", p), data)
}
