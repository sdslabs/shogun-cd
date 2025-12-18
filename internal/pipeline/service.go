package pipeline

import "github.com/kunalvirwal/shogun-cd/internal/utils"

type PipelineService interface {
	LoadPipeline(path string)
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
