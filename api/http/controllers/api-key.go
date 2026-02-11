package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateAPIKey(c *gin.Context) {

	// [TODO] create api key and display it only ONCE (alias -> unique identifier)

	h.response.Created(c, "New API Key Created", nil)
}

func (h *Handler) ListAllAPIKeys(c *gin.Context) {

	// [TODO] display aliases and other metadata of api keys (NEVER the actual key)

	h.response.Success(c, "All API Keys", nil)
}

func (h *Handler) DeleteAPIKey(c *gin.Context) {
	key := c.Param("key")

	// [TODO] delete an api key through given alias

	h.response.Success(c, fmt.Sprintf("Key Deleted - %v", key), nil)
}
