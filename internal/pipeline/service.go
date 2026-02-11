package pipeline

import (
	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/secrets"
	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type PipelineService interface {
	// Loads a pipeline from the specified yaml file path
	LoadPipeline(f []byte) *Pipeline
	// Executes the given pipeline with the specified trigger
	ExecutePipeline(pipeline *Pipeline, trigger TriggerKind, targets map[string]*target.Target, variables map[string]string) bool
}

type Service struct {
	logger        utils.Logger
	gitService    git.GitService
	secretService secrets.SecretService
}

func NewPipelineService(logger utils.Logger, gitService git.GitService, secretService secrets.SecretService) *Service {
	return &Service{
		logger:        logger,
		gitService:    gitService,
		secretService: secretService,
	}
}
