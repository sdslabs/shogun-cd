package main

import (
	api "github.com/kunalvirwal/shogun-cd/api/http"
	"github.com/kunalvirwal/shogun-cd/internal/app"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/orchestrator"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"github.com/kunalvirwal/shogun-cd/webhooks"
)

func main() {
	l, c, w := initServices()
	api.StartAPIServer(l, c, w)
	// pipeline.LoadPipeline("./examples/pipeline.yaml")
}

func initServices() (utils.Logger, *config.Config, webhooks.WebhookService) {

	// Initialize logger
	logger := utils.NewLogger(utils.DebugLevel, true)

	// Load configurations
	cfg, err := config.LoadConfigs(logger)
	if err != nil {
		logger.LogNewError("Invalid config: Stopping Shogun...")
		return logger, nil, nil
	}

	logger.SetLevel(cfg.Debug)

	// Initialize Git service
	gitService, err := git.NewGitService(logger, cfg.GitConfig)
	if err != nil {
		logger.LogNewError("Unable to initialize Git service: Stopping Shogun...")
		return logger, cfg, nil
	}

	// Initialize Pipeline service
	pipelineService := pipeline.NewPipelineService(logger, gitService)

	// Initialize Target service
	targetService := target.NewTargetService(logger, gitService)

	orch := orchestrator.New(logger, gitService, pipelineService, targetService)
	orch.Start()

	// Initialize main application
	app := app.NewApp(logger, gitService, pipelineService, targetService)

	_ = app

	// <-make(chan struct{}) // Block forever

	wh := webhooks.NewWebhookService(logger)

	return logger, cfg, wh

}
