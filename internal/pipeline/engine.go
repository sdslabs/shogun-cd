package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/kunalvirwal/shogun-cd/internal/models"
	pipelineSteps "github.com/kunalvirwal/shogun-cd/internal/pipeline/steps"
	"github.com/kunalvirwal/shogun-cd/internal/sshclient"
	"github.com/kunalvirwal/shogun-cd/internal/target"
)

func (p *Service) ExecutePipeline(runID uint, pipeline *Pipeline, trigger TriggerKind, targets map[string]*target.Target, hookValues map[string]string) (success bool) {
	ctx := context.Background()
	defer func() {
		if err := p.finishPipelineRun(ctx, runID, success); err != nil {
			p.logger.LogNewError("Failed to persist final state for pipeline run %d: %v", runID, err)
		}
	}()

	// [TODO] Move this check outside to caller
	if !pipeline.Metadata.Enabled {
		p.logger.LogInfo("Pipeline %s is disabled; skipping execution", pipeline.Metadata.Name)
		return false
	}

	if hookValues == nil {
		hookValues = make(map[string]string)
	}

	found := false
	for _, t := range pipeline.Spec.Triggers {
		if t.Type == string(trigger) {
			found = true
			break
		}
	}

	if !found {
		p.logger.LogNewError("Pipeline %s does not have trigger of type %s; cannot execute", pipeline.Metadata.Name, string(trigger))
		return false
	}

	deps := &pipelineSteps.StepDeps{
		PipelineName:  pipeline.Metadata.Name,
		Logger:        p.logger,
		GitService:    p.gitService,
		SecretService: p.secretService,
		Targets:       targets,
		HookValues:    hookValues,
		SSHManager:    sshclient.NewSSHManager(),
	}

	// SSH Client cleanup after pipeline execution only if SSHManager was initialized
	defer deps.SSHManager.CleanupSSHClients()

	success = true
	output := p.logger.LogShogunInfo("", "Starting execution of pipeline: %s", pipeline.Metadata.Name)
	p.logger.LogInfo("Executing pipeline: %s", pipeline.Metadata.Name)
	for i, sw := range pipeline.Spec.Steps {

		step := sw.Step

		// If trigger is specified for the step, and the step's trigger doesn't match the pipeline trigger, skip the step
		if step.Trigger() != "" && step.Trigger() != string(trigger) {
			output = p.logger.LogShogunInfo(output, "Skipping step %d as its trigger '%s' does not match pipeline trigger '%s'", i+1, step.Trigger(), string(trigger))
			continue
		}

		output = p.logger.LogShogunInfo(output, "Executing step %d of type %s", i+1, step.Type())

		stepData := &models.PipelineRunStep{
			RunID:     runID,
			StepIndex: i,
			StepType:  step.Type(),
			Status:    models.StepStatusInProgress,
			StartedAt: time.Now(),
		}
		if err := p.store.CreatePipelineRunStep(ctx, stepData); err != nil {
			p.logger.LogNewError("Failed to persist start of step %d for pipeline run %d: %v", i+1, runID, err)
			success = false
			break
		}

		out, err := step.Execute(ctx, deps)
		output += out

		stepData.Logs = p.logger.LogShogunInfo(stepData.Logs, out)

		// p.logger.Log("Step %d output:\n %s", i+1, out)
		if err != nil {
			p.logger.LogNewError("Step %d failed: %v", i+1, err)
			output, _ = deps.Logger.LogShogunError(output, "Step %d failed: %v\n", i+1, err)
			success = false

			stepData.Status = models.StepStatusFailed
			stepData.Logs, _ = p.logger.LogShogunError(stepData.Logs, "Step %d failed: %v\n", i+1, err)
			now := time.Now()
			stepData.FinishedAt = &now
			if storeErr := p.store.UpdatePipelineRunStep(ctx, stepData); storeErr != nil {
				p.logger.LogNewError("Failed to persist failure of step %d for pipeline run %d: %v", i+1, runID, storeErr)
			}

			break
		}
		output = p.logger.LogShogunInfo(output, "Step %d executed successfully\n", i+1)

		stepData.Status = models.StepStatusSucceeded
		stepData.Logs = p.logger.LogShogunInfo(stepData.Logs, "Step %d executed successfully\n", i+1)
		now := time.Now()
		stepData.FinishedAt = &now
		if err := p.store.UpdatePipelineRunStep(ctx, stepData); err != nil {
			p.logger.LogNewError("Failed to persist success of step %d for pipeline run %d: %v", i+1, runID, err)
			success = false
			break
		}

		p.logger.Log("Step %d executed successfully", i+1)
	}

	fmt.Println(output)

	return success
}
