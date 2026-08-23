package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/dto"
)

func (h *Handler) ListAllTargets(c *gin.Context) {
	targets := h.orch.ListTargets()
	data := make([]dto.TargetSummary, 0, len(targets))
	for _, resource := range targets {
		data = append(data, dto.TargetSummary{
			Name: resource.Metadata.Name,
			Type: resource.Metadata.Type,
			Host: resource.Spec.Host,
			User: resource.Spec.User,
			Port: resource.Spec.Port,
		})
	}

	h.response.Success(c, "All Targets", data)
}

func (h *Handler) ListAllPipelines(c *gin.Context) {
	pipelines := h.orch.ListPipelines()
	pipelineNames := make([]string, 0, len(pipelines))
	for _, resource := range pipelines {
		pipelineNames = append(pipelineNames, resource.Metadata.Name)
	}

	latestRuns, err := h.store.Pipeline.FetchLatestPipelineRuns(c.Request.Context(), pipelineNames)
	if err != nil {
		h.response.ServerError(c, err)
		return
	}

	data := make([]dto.PipelineSummary, 0, len(pipelines))
	for _, resource := range pipelines {
		triggers := make([]dto.PipelineTrigger, 0, len(resource.Spec.Triggers))
		for _, trigger := range resource.Spec.Triggers {
			triggers = append(triggers, dto.PipelineTrigger{
				Type:  trigger.Type,
				Paths: append([]string(nil), trigger.Paths...),
			})
		}

		steps := make([]dto.PipelineStepSummary, 0, len(resource.Spec.Steps))
		for i, wrapper := range resource.Spec.Steps {
			step := wrapper.Step
			summary := dto.PipelineStepSummary{
				Index:       i,
				Type:        step.Type(),
				TriggerWhen: step.Trigger(),
			}
			if targeted, ok := step.(interface{ TargetInstance() string }); ok {
				summary.Target = targeted.TargetInstance()
			}
			steps = append(steps, summary)
		}

		summary := dto.PipelineSummary{
			Name:     resource.Metadata.Name,
			Enabled:  resource.Metadata.Enabled,
			Triggers: triggers,
			Steps:    steps,
		}

		latestRun := latestRuns[resource.Metadata.Name]
		if latestRun != nil {
			summary.LastRun = &dto.PipelineRunSummary{
				ID:          latestRun.ID,
				TriggerKind: latestRun.TriggerKind,
				Status:      string(latestRun.Status),
				Success:     latestRun.Success,
				StartedAt:   latestRun.StartedAt,
				FinishedAt:  latestRun.FinishedAt,
			}
		}

		data = append(data, summary)
	}

	h.response.Success(c, "All Pipelines", data)
}
