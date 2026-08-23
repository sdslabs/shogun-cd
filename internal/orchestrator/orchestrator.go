package orchestrator

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type Orchestrator interface {
	// Starts all the necessary services and routines
	Start()
	// Runs the indexer to load pipelines and targets
	RunIndexer()
	// Run repo Poller go routine
	StartPoller(ctx context.Context)
	// Executes a pipeline with the given name, trigger kind, and variables
	RunPipeline(ctx context.Context, pipelineName string, triggerKind pipeline.TriggerKind, variables map[string]string) (uint, error)
	// Checks if a pipeline with a given name exists
	PipelineExists(pipelineName string) bool
	// // Locks the piplines for running indexer and pull operations
	// LockPipelines()
	// // Unlocks the pipelines after running indexer and pull operations
	// UnlockPipelines()

	// Add mannual pipeline pause/resume
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
	// RWMutex to synchronize pipeline runs with Indexer and Pull operation
	mu sync.RWMutex
}

func New(logger utils.Logger, gitSvc git.GitService, pipelineSvc pipeline.PipelineService, targetSvc target.TargetService) Orchestrator {
	return &orchestrator{
		logger:          logger,
		gitService:      gitSvc,
		pipelineService: pipelineSvc,
		targetService:   targetSvc,
	}
}
