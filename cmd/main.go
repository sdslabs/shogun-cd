package main

import (
	api "github.com/kunalvirwal/shogun-cd/api/http"
	"github.com/kunalvirwal/shogun-cd/internal/app"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/encryption"
	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/orchestrator"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/secrets"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/kunalvirwal/shogun-cd/internal/users"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"github.com/kunalvirwal/shogun-cd/internal/webhooks"
)

func main() {
	initServices()
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

	db, err := store.Connect(cfg, logger)
	if err != nil {
		logger.LogNewError(err.Error())
		return
	}
	store := store.NewStore(cfg, db)
	encryption, err := encryption.NewService(cfg, logger)
	if err != nil {
		logger.LogNewError("Unable to Initialise Encryption service : %v", err.Error())
		return
	}

	// Initialize Secret Manager
	secretService := secrets.NewSecretService(encryption, store, logger)

	// Initialize User Service
	userService := users.NewUserService(store, logger)

	// Initialize Git service
	gitService, err := git.NewGitService(logger, cfg)
	if err != nil {
		logger.LogNewError("Unable to initialize Git service: Stopping Shogun...")
	}

	// Initialize Pipeline service
	pipelineService := pipeline.NewPipelineService(logger, gitService, secretService, store.Pipeline)

	// Initialize Target service
	targetService := target.NewTargetService(logger, gitService)

	orch := orchestrator.New(logger, gitService, pipelineService, targetService)
	orch.Start()

	// Initialize main application
	app := app.NewApp(logger, gitService, pipelineService, targetService, secretService)

	_ = app

	if err := seedAdmin(userService, cfg); err != nil {
		logger.LogNewError("Unable to Seed Admin account : %v", err.Error())
		return
	}

	// Initialize webhook service
	webhookService := webhooks.NewWebhookService(logger, orch, store, encryption)
	if err = webhookService.Load(); err != nil {
		logger.LogNewError("Unable to load Webhooks : %v", err.Error())
	}

	// Initialize api
	go api.StartAPIServer(logger, cfg, webhookService, userService, secretService, store)

	<-make(chan struct{}) // Block forever
}
