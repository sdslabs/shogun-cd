package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	api "github.com/kunalvirwal/shogun-cd/api/http"
	"github.com/kunalvirwal/shogun-cd/internal/app"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/orchestrator"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"github.com/kunalvirwal/shogun-cd/internal/webhooks"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	initServices(ctx)
	// pipeline.LoadPipeline("./examples/pipeline.yaml")

	<-ctx.Done() // Block forever
}

func initServices(ctx context.Context) {

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
	gitService, err := git.NewGitService(logger, cfg)
	if err != nil {
		logger.LogNewError("Unable to initialize Git service: Stopping Shogun...")
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

	db, err := store.Connect(cfg, logger)
	if err != nil {
		logger.LogNewError(err.Error())
		return
	}

	store := store.NewStore(db)
	if err := seedAdmin(ctx, store, cfg); err != nil {
		logger.LogNewError("Unable to Seed Admin account : %v", err.Error())
		return
	}

	// Initialize webhook service
	webhook := webhooks.NewWebhookService(logger, orch, store)
	if err = webhook.Load(ctx); err != nil {
		logger.LogNewError("Unable to load Webhooks : %v", err.Error())
	}

	// Initialize api
	go api.StartAPIServer(ctx, logger, cfg, webhook, store)

}
