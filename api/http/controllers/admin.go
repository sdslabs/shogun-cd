package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/dto"
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

func (h *Handler) ListAllSecrets(c *gin.Context) {
	var req dto.SecretFilter

	if err := c.ShouldBind(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	// [TODO] granular error handling for different http status codes.
	data, err := h.secret.FindMany(c.Request.Context(), &req)
	if err != nil {
		h.response.ServerError(c, err)
		return
	}

	h.response.Success(c, "All Secrets (without value)", data)
}

func (h *Handler) CreateSecrets(c *gin.Context) {
	var req []dto.SecretInput

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	// [TODO] granular error handling for different http status codes.
	if err := h.secret.Create(c.Request.Context(), req); err != nil {
		h.response.ServerError(c, err)
		return
	}

	h.response.Created(c, "New Secret(s) Created", nil)
}

func (h *Handler) ForceCreateSecrets(c *gin.Context) {
	var req []dto.SecretInput

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	// [TODO] granular error handling for different http status codes.
	if err := h.secret.Upsert(c.Request.Context(), req); err != nil {
		h.response.ServerError(c, err)
		return
	}

	h.response.Created(c, "New Secret(s) Created", nil)
}

func (h *Handler) UpdateSecret(c *gin.Context) {
	var req dto.SecretInput

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	// [TODO] granular error handling for different http status codes.
	if err := h.secret.Update(c.Request.Context(), &req); err != nil {
		h.response.ServerError(c, err)
		return
	}

	h.response.Success(c, fmt.Sprintf("Secret Updated - %v", req.Name), nil)

}

func (h *Handler) DeleteSecret(c *gin.Context) {
	s := c.Param("secret")

	if err := h.secret.Delete(c.Request.Context(), s); err != nil {
		h.response.ServerError(c, err)
		return
	}

	h.response.Success(c, fmt.Sprintf("Secret Deleted - %v", s), nil)
}
