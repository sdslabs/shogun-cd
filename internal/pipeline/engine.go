package pipeline

import pipelineSteps "github.com/kunalvirwal/shogun-cd/internal/pipeline/steps"

func (p *Service) ExecutePipeline(pipeline *Pipeline, trigger TriggerKind) bool {

	// [TODO] Move this check outside to caller
	if !pipeline.Metadata.Enabled {
		p.logger.LogInfo("Pipeline %s is disabled; skipping execution", pipeline.Metadata.Name)
		return false
	}

	deps := &pipelineSteps.StepDeps{
		Logger:     p.logger,
		GitService: p.gitService,
	}

	p.logger.LogInfo("Executing pipeline: %s", pipeline.Metadata.Name)
	for i, sw := range pipeline.Spec.Steps {

		step := sw.Step
		p.logger.Log("Executing step %d of type %s", i+1, step.Type())

		// If trigger is specified for the step, and the step's trigger doesn't match the pipeline trigger, skip the step
		if step.Trigger() != "" && step.Trigger() != string(trigger) {
			continue
		}

		err := step.Execute(deps)
		if err != nil {
			p.logger.LogNewError("Step %d failed: %v", i+1, err)
			return false
		}
		p.logger.Log("Step %d executed successfully", i+1)

	}
	return true
}
