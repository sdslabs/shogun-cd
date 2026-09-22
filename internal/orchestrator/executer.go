package orchestrator

import (
	"context"
	"fmt"

	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
)

// RunPipeline executes a pipeline with the given name, trigger kind, and variables
// It returns the persisted run ID, or an error if the run could not be started.
func (o *orchestrator) RunPipeline(ctx context.Context, pipelineName string, triggerKind pipeline.TriggerKind, triggerValues map[string]string) (uint, error) {

	pipelines := o.Pipelines.Load()
	if pipelines == nil {
		return 0, fmt.Errorf("%w: %s", ErrPipelineNotFound, pipelineName)
	}
	pipelineInstance, exists := (*pipelines)[pipelineName]
	if !exists {
		o.logger.Log("Pipeline not found: " + pipelineName)
		return 0, fmt.Errorf("%w: %s", ErrPipelineNotFound, pipelineName)
	}
	if !pipelineInstance.Metadata.Enabled {
		return 0, fmt.Errorf("%w: %s", ErrPipelineDisabled, pipelineName)
	}
	if !pipelineInstance.SupportsTrigger(triggerKind) {
		return 0, fmt.Errorf("%w: %s", ErrTriggerNotConfigured, triggerKind)
	}

	runID, err := o.pipelineService.CreatePipelineRun(ctx, pipelineInstance.Metadata.Name, triggerKind)
	if err != nil {
		return 0, fmt.Errorf("failed to create pipeline run: %w", err)
	}

	go func() {
		o.mu.RLock()
		defer o.mu.RUnlock()

		targets := o.Targets.Load()
		success := o.pipelineService.ExecutePipeline(runID, pipelineInstance, triggerKind, *targets, triggerValues)

		if !success {
			o.logger.Log("Pipeline execution failed: " + pipelineName)
		} else {
			o.logger.Log("Pipeline executed successfully: " + pipelineName)
		}
	}()

	return runID, nil
}

// LockPipelines locks the orchestrator's pipeline mutex for indexer and pull operations
func (o *orchestrator) LockPipelines() {
	o.mu.Lock()
}

// UnlockPipelines unlocks the orchestrator's pipeline mutex after indexer and pull operations
func (o *orchestrator) UnlockPipelines() {
	o.mu.Unlock()
}

// PipelineExists checks if a pipeline with the given name exists
func (o *orchestrator) PipelineExists(pipelineName string) bool {
	pipelines := o.Pipelines.Load()
	_, exists := (*pipelines)[pipelineName]
	return exists
}
