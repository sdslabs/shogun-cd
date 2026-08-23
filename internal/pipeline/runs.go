package pipeline

import (
	"context"
	"time"

	"github.com/kunalvirwal/shogun-cd/internal/models"
)

func (p *Service) CreatePipelineRun(ctx context.Context, pipelineName string, trigger TriggerKind) (uint, error) {
	run := &models.PipelineRun{
		Pipeline:    pipelineName,
		TriggerKind: string(trigger),
		Status:      models.PipelineRunStatusRunning,
		StartedAt:   time.Now(),
	}

	if err := p.store.CreatePipelineRun(ctx, run); err != nil {
		return 0, err
	}

	return run.ID, nil
}

func (p *Service) finishPipelineRun(ctx context.Context, runID uint, success bool) error {
	status := models.PipelineRunStatusFailed
	if success {
		status = models.PipelineRunStatusSucceeded
	}

	finishedAt := time.Now()
	return p.store.UpdatePipelineRun(ctx, &models.PipelineRun{
		ID:         runID,
		Status:     status,
		Success:    &success,
		FinishedAt: &finishedAt,
	})
}
