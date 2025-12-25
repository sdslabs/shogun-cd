package main

import (
	"github.com/kunalvirwal/shogun-cd/internal/app"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

func main() {
	initServices()
	// api.StartAPIServer()
	// pipeline.LoadPipeline("./examples/pipeline.yaml")
}

func initServices() {

	// Initialize logger
	logger := utils.NewLogger(utils.DebugLevel, true)

	// Load configurations
	cfg, err := config.LoadConfigs(logger)
	if err != nil {
		logger.LogNewError("Invalid config: Stopping Shogun...")
		return
	}

	logger.SetLevel(cfg.Debug)

	// Initialize Git service
	gitService, err := git.NewGitService(logger, cfg.GitConfig)
	if err != nil {
		logger.LogNewError("Unable to initialize Git service: Stopping Shogun...")
		return
	}

	// Initialize Pipeline service
	pipelineService := pipeline.NewPipelineService(logger, gitService)

	// Initialize main application
	app := app.NewApp(pipelineService, gitService, logger)

	pipelineService.LoadPipeline("./cache/pipeline.yaml")
	_ = app

	<-make(chan struct{}) // Block forever

}
