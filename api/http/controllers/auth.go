package controllers

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/api/http/apiutils"
	"github.com/kunalvirwal/shogun-cd/internal/users"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(c *gin.Context) {
	var req dto.LoginInput

	if err := c.ShouldBind(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	ctx := c.Request.Context()
	err := h.user.Create(ctx, &users.CreateParams{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, users.ErrEmailTaken):
			h.response.BadRequest(c, users.ErrEmailTaken.Error(), err)
			return

		default:
			h.response.ServerError(c, err)
			return
		}
	}

	h.response.Success(c, fmt.Sprintf("registered new user - %v", req.Email), nil)
}

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginInput

	if err := c.ShouldBind(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	user, err := h.user.FindOne(c.Request.Context(), &users.FilterParams{
		Email: req.Email,
	})
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			h.response.Unauthorized(c, "Invalid email or password", err)
			return
		default:
			h.response.ServerError(c, err)
			return
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			h.response.Unauthorized(c, "Invalid email or password", err)
			return
		default:
			h.response.ServerError(c, err)
			return
		}
	}
	if !user.IsActive {
		h.response.Forbidden(c, "This Account is Suspended", users.ErrAccountSuspended)
		return
	}

	role := user.Role
	secret := h.config.ApiConfig.JWT.Secret
	exp := h.config.ApiConfig.JWT.ExpirationHours

	token, err := apiutils.GenerateToken(req.Email, role, secret, exp)
	if err != nil {
		h.response.ServerError(c, err)
		return
	}

	data := gin.H{
		"token":            token,
		"expiration_hours": exp,
	}

	h.response.Success(c, fmt.Sprintf("logged in as - %v", req.Email), data)
}
