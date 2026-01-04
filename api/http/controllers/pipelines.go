package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListPipelineRuns(c *gin.Context) {
	p := c.Param("pipeline")

	// [TODO] list pipeline runs

	h.response.Success(c, fmt.Sprintf("All Runs for Pipeline - %v", p), nil)
}

func (h *Handler) StreamPipelineRunLogs(c *gin.Context) {
	p := c.Param("pipeline")
	r := c.Param("run_id")

	// [TODO] stream pipeline logs

	h.response.Success(c, fmt.Sprintf("Run - %v Logs for Pipeline - %v", r, p), nil)
}
func (h *Handler) ListPipelineRunSteps(c *gin.Context) {
	p := c.Param("pipeline")
	r := c.Param("run_id")

	// [TODO] list the steps of a particular run

	h.response.Success(c, fmt.Sprintf("Run - %v Steps for Pipeline - %v", r, p), nil)
}
