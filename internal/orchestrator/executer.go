package orchestrator

import (
	"fmt"

	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
)

// RunPipeline executes a pipeline with the given name, trigger kind, and variables
// It returns an error if the pipeline is not found
func (o *orchestrator) RunPipeline(pipelineName string, triggerKind pipeline.TriggerKind, variables map[string]string) error {

	// Fast fail
	if !o.PipelineExists(pipelineName) {
		o.logger.Log("Pipeline not found: " + pipelineName)
		return fmt.Errorf("pipeline not found: %s", pipelineName)
	}

	// If fast fail passes then pipline is queued for execution
	// There is a possibility of TOCTTOU race condition but it is acceptable by the use of 2nd check below
	go func() {
		pipelines := o.Pipelines.Load()
		pipeline, exists := (*pipelines)[pipelineName]
		if !exists {
			o.logger.LogNewError("Pipeline %s removed before execution could start", pipelineName)
			return
		}

		o.mu.RLock()
		success := o.pipelineService.ExecutePipeline(pipeline, triggerKind, variables)
		o.mu.RUnlock()

		if !success {
			o.logger.Log("Pipeline execution failed: " + pipelineName)
		} else {
			o.logger.Log("Pipeline executed successfully: " + pipelineName)
		}
	}()

	return nil
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
