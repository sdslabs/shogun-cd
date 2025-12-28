package orchestrator

import (
	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type Orchestrator interface {
	Start()
	RunIndexer()
}

type orchestrator struct {
	logger          utils.Logger
	gitService      git.GitService
	pipelineService pipeline.PipelineService
	targetService   target.TargetService
}

func New(logger utils.Logger, gitSvc git.GitService, pipelineSvc pipeline.PipelineService, targetSvc target.TargetService) Orchestrator {
	return &orchestrator{
		logger:          logger,
		gitService:      gitSvc,
		pipelineService: pipelineSvc,
		targetService:   targetSvc,
	}
}
