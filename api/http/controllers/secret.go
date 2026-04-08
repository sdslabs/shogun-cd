package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/internal/secrets"
)

func (h *Handler) ListAllSecrets(c *gin.Context) {
	var req dto.SecretFilter

	if err := c.ShouldBind(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	data, err := h.secret.FindMany(c.Request.Context(), &secrets.FilterParams{
		Name: req.Name,
	})
	if err != nil {
		h.response.ServerError(c, err)
		return
	}

	h.response.Success(c, "All Secrets (names only)", data)
}

func (h *Handler) SetSecrets(c *gin.Context) {
	var req []dto.SecretInput

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	secretArr := make([]secrets.CreateParams, len(req))
	for i, secret := range req {
		secretArr[i] = secrets.CreateParams{
			Name:  secret.Name,
			Value: secret.Value,
		}
	}

	if err := h.secret.SetSecrets(c.Request.Context(), secretArr); err != nil {
		h.response.ServerError(c, err)
		return
	}

	h.response.Created(c, "New Secret(s) Created", nil)
}

func (h *Handler) DeleteSecret(c *gin.Context) {
	s := c.Param("secret")

	if err := h.secret.DeleteSecret(c.Request.Context(), s); err != nil {
		h.response.ServerError(c, err)
		return
	}

	h.response.Success(c, fmt.Sprintf("Secret Deleted - %v", s), nil)
}
