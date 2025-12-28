package orchestrator

import (
	"sync/atomic"

	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type Orchestrator interface {
	Start()
	RunIndexer()
}

type PipelineMap map[string]*pipeline.Pipeline
type TargetMap map[string]*target.Target

type orchestrator struct {
	logger          utils.Logger
	gitService      git.GitService
	pipelineService pipeline.PipelineService
	targetService   target.TargetService

	Pipelines atomic.Pointer[PipelineMap]
	Targets   atomic.Pointer[TargetMap]
}

func New(logger utils.Logger, gitSvc git.GitService, pipelineSvc pipeline.PipelineService, targetSvc target.TargetService) Orchestrator {
	return &orchestrator{
		logger:          logger,
		gitService:      gitSvc,
		pipelineService: pipelineSvc,
		targetService:   targetSvc,
	}
}
