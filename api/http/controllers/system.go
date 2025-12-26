package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/response"
)

func (h *Handler) ListAllTargets(c *gin.Context) {

	response.Success(c, "All Targets", nil)
}

func (h *Handler) ListAllPipelines(c *gin.Context) {

	response.Success(c, "All Pipelines", nil)
}
