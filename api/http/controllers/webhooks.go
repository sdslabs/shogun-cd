package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (h *Handler) HandleWebhook(c *gin.Context) {
	token := c.Param("token")

	//[TODO] webhook handler

	h.response.Success(c, fmt.Sprintf("token received - %v", token), nil)
}

func (h *Handler) ListPipelineWebhooks(c *gin.Context) {
	p := c.Param("pipeline")

	h.response.Success(c, fmt.Sprintf("All Webhooks of Pipeline - %v", p), nil)
}

func (h *Handler) CreateWebhook(c *gin.Context) {
	p := c.Param("pipeline")

	h.response.Created(c, fmt.Sprintf("Webhook Created for Pipeline - %v", p), nil)
}

func (h *Handler) DeleteWebhook(c *gin.Context) {
	p := c.Param("pipeline")
	w := c.Param("webhook_id")

	h.response.Success(c, fmt.Sprintf("Webhook - %v Deleted for Pipeline - %v", w, p), nil)
}
