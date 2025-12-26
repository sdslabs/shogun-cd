package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/response"
)

func (h *Handler) Login(c *gin.Context) {

	//[TODO] login handler

	response.Success(c, fmt.Sprintf("logged in as - %v", nil), nil)
}
