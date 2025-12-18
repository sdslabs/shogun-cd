package pipeline

import pipelineSteps "github.com/kunalvirwal/shogun-cd/internal/pipeline/steps"

func (p *Service) ExecutePipeline(pipeline *Pipeline, trigger TriggerKind) bool {

	// [TODO] Move this check outside to caller
	if !pipeline.Metadata.Enabled {
		p.logger.LogInfo("Pipeline %s is disabled; skipping execution", pipeline.Metadata.Name)
		return false
	}

	p.logger.LogInfo("Executing pipeline: %s", pipeline.Metadata.Name)
	for i, sw := range pipeline.Spec.Steps {

		p.logger.Log("Executing step %d", i+1)
		var step pipelineSteps.Step

		switch sw.Step.Type() {
		case pipelineSteps.MutateType:
			step = sw.Step.(*pipelineSteps.MutateStep)
		case pipelineSteps.ApplyType:
			step = sw.Step.(*pipelineSteps.ApplyStep)
		case pipelineSteps.ExecType:
			step = sw.Step.(*pipelineSteps.ExecStep)
		case pipelineSteps.SyncType:
			step = sw.Step.(*pipelineSteps.SyncStep)
		default:
			p.logger.LogNewError("unknown step type in step %d", i+1)
			return false
		}

		// If no trigger is specified for the step, or if the step's trigger matches the pipeline trigger, execute the step
		if step.Trigger() == "" || step.Trigger() == string(trigger) {
			step.Execute()
		}
	}
	return true
}
