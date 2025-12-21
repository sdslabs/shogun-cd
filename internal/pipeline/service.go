package pipeline

import "github.com/kunalvirwal/shogun-cd/internal/utils"

type PipelineService interface {
	// Loads a pipeline from the specified yaml file path
	LoadPipeline(path string)
	// Executes the given pipeline with the specified trigger
	ExecutePipeline(pipeline *Pipeline, trigger TriggerKind) bool
}

type Service struct {
	logger utils.Logger
}

func NewPipelineService(logger utils.Logger) *Service {
	return &Service{
		logger: logger,
	}
}
