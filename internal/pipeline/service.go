package pipeline

import (
	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type PipelineService interface {
	// Loads a pipeline from the specified yaml file path
	LoadPipeline(path string)
	// Executes the given pipeline with the specified trigger
	ExecutePipeline(pipeline *Pipeline, trigger TriggerKind) bool
}

type Service struct {
	logger     utils.Logger
	gitService git.GitService
}

func NewPipelineService(logger utils.Logger, gitService git.GitService) *Service {
	return &Service{
		logger:     logger,
		gitService: gitService,
	}
}
