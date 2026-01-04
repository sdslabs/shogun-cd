package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/api/http/apiutils"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginInput

	if err := c.ShouldBind(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	//[TODO] multi user support through DB lookup
	validEmail := utils.SecureCompare(req.Email, h.config.ApiConfig.Admin.Email)
	validPassword := utils.SecureCompare(req.Password, h.config.ApiConfig.Admin.Password)
	if !validEmail || !validPassword {
		h.response.Unauthorized(c, "Invalid Email or Password", nil)
		return
	}

	//[TODO] DB call to get the user's role to be embedded in jwt.
	role := apiutils.RoleAdmin
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
