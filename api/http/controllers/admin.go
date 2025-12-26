package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/response"
)

func (h *Handler) CreateAPIKey(c *gin.Context) {

	response.Created(c, "New API Key Created", nil)
}

func (h *Handler) ListAllAPIKeys(c *gin.Context) {

	response.Success(c, "All API Keys", nil)
}

func (h *Handler) DeleteAPIKey(c *gin.Context) {
	key := c.Param("key")

	response.Success(c, fmt.Sprintf("Key Deleted - %v", key), nil)
}

func (h *Handler) ListAllSecrets(c *gin.Context) {

	response.Success(c, "All Secrets (without value)", nil)
}

func (h *Handler) CreateSecret(c *gin.Context) {

	response.Created(c, "New Secret Created", nil)
}

func (h *Handler) UpdateSecret(c *gin.Context) {
	s := c.Param("secret")

	response.Success(c, fmt.Sprintf("Secret Updated - %v", s), nil)

}

func (h *Handler) DeleteSecret(c *gin.Context) {
	s := c.Param("secret")

	response.Success(c, fmt.Sprintf("Secret Deleted - %v", s), nil)
}
