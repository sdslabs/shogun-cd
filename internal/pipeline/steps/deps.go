package pipelineSteps

import "github.com/kunalvirwal/shogun-cd/internal/utils"

// StepDeps represents dependencies required by pipeline steps.
type StepDeps struct {
	// [TODO] Git and SSH
	Logger utils.Logger
}
