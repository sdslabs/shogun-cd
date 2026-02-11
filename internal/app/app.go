package app

import (
	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/secrets"
	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type App struct {
	Logger          utils.Logger
	GitService      git.GitService
	PipelineService pipeline.PipelineService
	TargetService   target.TargetService
	SecretService   secrets.SecretService
}

func NewApp(logger utils.Logger, gitService git.GitService, pipelineService pipeline.PipelineService, targetService target.TargetService, secretService secrets.SecretService) *App {
	return &App{
		Logger:          logger,
		GitService:      gitService,
		PipelineService: pipelineService,
		TargetService:   targetService,
		SecretService:   secretService,
	}
}
