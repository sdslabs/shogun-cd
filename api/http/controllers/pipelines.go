package controllers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/internal/store"
)

func (h *Handler) ListPipelineRuns(c *gin.Context) {
	p := c.Param("pipeline")
	r := c.Param("run_id")

	var runID uint
	if r != "" {
		parsedID, err := strconv.ParseUint(r, 10, 32)
		if err != nil {
			h.response.BadRequest(c, "Invalid RunID format (expected uint)", err)
			return
		}
		runID = uint(parsedID)
	}

	data, err := h.store.Pipeline.FetchPipelineRunDetails(c.Request.Context(), p, runID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			h.response.NotFound(c, err.Error(), err)
		default:
			h.response.ServerError(c, err)
		}
		return
	}

	h.response.Success(c, fmt.Sprintf("Pipeline Run(s) Details - %v", p), data)
}

func (h *Handler) StreamPipelineRunLogs(c *gin.Context) {
	p := c.Param("pipeline")
	r := c.Param("run_id")

	// [TODO] stream pipeline logs

	h.response.Success(c, fmt.Sprintf("Run - %v Logs for Pipeline - %v", r, p), nil)
}
