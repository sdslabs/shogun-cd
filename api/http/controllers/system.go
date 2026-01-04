package controllers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListAllTargets(c *gin.Context) {

	// [TODO] lists all targets

	h.response.Success(c, "All Targets", nil)
}

func (h *Handler) ListAllPipelines(c *gin.Context) {

	// [TODO] lists all pipelines

	h.response.Success(c, "All Pipelines", nil)
}
