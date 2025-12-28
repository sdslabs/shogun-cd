package pipeline

import (
	"strings"

	pipelineSteps "github.com/kunalvirwal/shogun-cd/internal/pipeline/steps"
	"go.yaml.in/yaml/v3"
)

func (p *Service) LoadPipeline(f []byte) *Pipeline {

	var pipeline Pipeline
	if err := yaml.Unmarshal(f, &pipeline); err != nil {
		p.logger.LogNewError("failed to unmarshal pipeline YAML: %v", err)
		return nil
	}

	// [TODO]: Validate the pipeline structure here if needed
	if !p.validatePipeline(&pipeline) {
		return nil
	}

	for i, sw := range pipeline.Spec.Steps {
		p.logger.Log("Step %d: Type=%s, Details=%+v", i+1, sw.Step.Type(), sw.Step)
	}
	p.logger.LogInfo("Pipeline %v loaded successfully", pipeline.Metadata.Name)

	return &pipeline
	// [TODO]: Further processing of the loaded pipeline, store into DB and all
}

func (p *Service) validatePipeline(pipeline *Pipeline) bool {
	// [TODO]: Implement validation logic
	if pipeline.ApiVersion != "shogun.dev/v1" {
		p.logger.LogNewError("invalid apiVersion: %s", pipeline.ApiVersion)
		return false
	}
	if pipeline.Kind != PipelineKind {
		p.logger.LogNewError("invalid kind: %s", pipeline.Kind)
		return false
	}
	// [TODO]: Check is name is unique
	if pipeline.Metadata.Name == "" {
		p.logger.LogNewError("metadata.name cannot be empty")
		return false
	}
	if len(pipeline.Spec.Triggers) == 0 {
		p.logger.LogNewError("at least one trigger must be specified")
		return false
	}
	for _, trigger := range pipeline.Spec.Triggers {
		if trigger.Type == string(WebhookTriggerKind) {
			continue
		}
		if trigger.Type == string(GitChangesTriggerKind) {
			if len(trigger.Paths) == 0 {
				p.logger.LogNewError("git_changes trigger must specify atleast one path")
				return false
			}
			continue
		}
		p.logger.LogNewError("invalid trigger type")
		return false
	}

	for i, sw := range pipeline.Spec.Steps {
		if sw.Step == nil {
			p.logger.LogNewError("step %d is nil", i+1)
			return false
		}

		switch sw.Step.Type() {
		case pipelineSteps.MutateType:
			step := sw.Step.(*pipelineSteps.MutateStep)
			if !p.validTrigger(step.TriggerWhen, i) {
				return false
			}
			for _, change := range step.Changes {
				if change.File == "" {
					p.logger.LogNewError("step %d mutate file cannot be empty", i+1)
					return false
				}
				if change.UpdateField == "" {
					p.logger.LogNewError("step %d mutate update_field cannot be empty", i+1)
					return false
				}
			}

		case pipelineSteps.SyncType:
			step := sw.Step.(*pipelineSteps.SyncStep)
			if !p.validTrigger(step.TriggerWhen, i) {
				return false
			}
			if step.Target == "" {
				p.logger.LogNewError("step %d exec target cannot be empty", i+1)
				return false
			}

		case pipelineSteps.ExecType:
			step := sw.Step.(*pipelineSteps.ExecStep)
			if !p.validTrigger(step.TriggerWhen, i) {
				return false
			}
			if step.Target == "" {
				p.logger.LogNewError("step %d exec target cannot be empty", i+1)
				return false
			}
			if len(step.Commands) == 0 {
				p.logger.LogNewError("step %d exec commands cannot be empty", i+1)
				return false
			}

		case pipelineSteps.ApplyType:
			step := sw.Step.(*pipelineSteps.ApplyStep)
			if !p.validTrigger(step.TriggerWhen, i) {
				return false
			}
			if step.Target == "" {
				p.logger.LogNewError("step %d exec target cannot be empty", i+1)
				return false
			}
			if len(step.Files) == 0 {
				p.logger.LogNewError("step %d apply files cannot be empty", i+1)
				return false
			}
			for j, file := range step.Files {
				if !strings.HasSuffix(file, ".yaml") && !strings.HasSuffix(file, ".yml") && !strings.HasSuffix(file, ".json") {
					p.logger.LogNewError("step %d apply has a file with invalid extension at index %d", i+1, j)
					return false
				}
			}

		default:
			p.logger.LogNewError("step %d has unknown type: %s", i+1, sw.Step.Type())
			return false
		}

	}
	return true
}

func (p *Service) validTrigger(trigger string, i int) bool {
	if trigger == "" || trigger == string(WebhookTriggerKind) || trigger == string(GitChangesTriggerKind) {
		return true
	}
	p.logger.LogNewError("step %d has invalid trigger_when: %s", i+1, trigger)
	return false
}
