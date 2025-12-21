package app

import (
	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type App struct {
	Logger          utils.Logger
	GitService      git.GitService
	PipelineService pipeline.PipelineService
}

func NewApp(pipelineService pipeline.PipelineService, gitService git.GitService, logger utils.Logger) *App {
	return &App{
		Logger:          logger,
		GitService:      gitService,
		PipelineService: pipelineService,
	}
}
