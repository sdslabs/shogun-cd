package main

import (
	"fmt"
	"os"

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
	if err := initServices(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initServices() error {

	// Initialize logger
	logger := utils.NewLogger(utils.DebugLevel, true)

	// Load configurations
	cfg, err := config.LoadConfigs(logger)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	logger.SetLevel(cfg.Debug)

	db, err := store.Connect(cfg, logger)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	store := store.NewStore(cfg, db)
	encryption, err := encryption.NewService(cfg, logger)
	if err != nil {
		return fmt.Errorf("initialize encryption service: %w", err)
	}

	// Initialize Secret Manager
	secretService := secrets.NewSecretService(encryption, store, logger)

	// Initialize User Service
	userService := users.NewUserService(store, logger)

	// Initialize Git service
	gitService, err := git.NewGitService(logger, cfg)
	if err != nil {
		return fmt.Errorf("initialize Git service: %w", err)
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
		return fmt.Errorf("seed admin account: %w", err)
	}

	// Initialize webhook service
	webhookService := webhooks.NewWebhookService(logger, orch, store, encryption)
	if err = webhookService.Load(); err != nil {
		logger.LogNewError("Unable to load Webhooks : %v", err.Error())
	}

	// Initialize api
	go api.StartAPIServer(logger, cfg, webhookService, userService, secretService, store, orch)

	<-make(chan struct{}) // Block forever
	return nil
}
