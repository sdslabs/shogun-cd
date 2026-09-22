package controllers

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/internal/orchestrator"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/store"
)

var triggerValueKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (h *Handler) StartPipelineRun(c *gin.Context) {
	var req dto.StartPipelineRunInput
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, "Bad Input", err)
		return
	}

	if err := validateTriggerValues(req.Values); err != nil {
		h.response.BadRequest(c, "Invalid trigger values", err)
		return
	}

	pipelineName := c.Param("pipeline")
	runID, err := h.orch.RunPipeline(c.Request.Context(), pipelineName, pipeline.UITriggerKind, req.Values)
	if err != nil {
		switch {
		case errors.Is(err, orchestrator.ErrPipelineNotFound):
			h.response.NotFound(c, err.Error(), err)
		case errors.Is(err, orchestrator.ErrPipelineDisabled), errors.Is(err, orchestrator.ErrTriggerNotConfigured):
			h.response.Conflict(c, err.Error(), err)
		default:
			h.response.ServerError(c, err)
		}
		return
	}

	h.response.Accepted(c, fmt.Sprintf("Pipeline run started - %s", pipelineName), dto.StartPipelineRunData{RunID: runID})
}

func validateTriggerValues(values map[string]string) error {
	for key := range values {
		if !triggerValueKeyPattern.MatchString(key) {
			return fmt.Errorf("invalid trigger value key %q", key)
		}
	}

	return nil
}

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
