package git

import (
	"context"
	"os/exec"
	"sync"

	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"golang.org/x/sync/singleflight"
)

type GitService interface {
	// Clones a git repository to the specified destination path
	CloneRepo(ctx context.Context) error
	// Polls the remote repository for changes if clone succeeded outherwise retries clone
	CloneAndStartPoller(ctx context.Context)
	// Pulls the latest changes from the remote repository
	Pull(ctx context.Context) error
	// Returns a channel that emits pull events with changed file paths
	GetPullEvents() <-chan []string
}

type Service struct {
	logger          utils.Logger
	repo            Repo
	pollingInterval int

	gitPath        string
	pullSF         singleflight.Group
	pullEventsChan chan []string
}

type Repo struct {
	URL      string
	Branch   string
	CloneDir string
	mu       sync.RWMutex
}

func NewGitService(logger utils.Logger, gitConfig config.Git) (GitService, error) {

	gitPath, err := exec.LookPath("git")
	if err != nil {
		logger.LogNewError("Git executable not found in PATH:", err)
		return nil, err
	}
	logger.Log("Git executable found at: %s", gitPath)

	gitSvc := &Service{
		logger:          logger,
		pollingInterval: gitConfig.PollingInterval,
		repo: Repo{
			URL:      gitConfig.Repo,
			Branch:   gitConfig.Branch,
			CloneDir: gitConfig.CloneDir,
		},
		gitPath:        gitPath,
		pullSF:         singleflight.Group{},
		pullEventsChan: make(chan []string, 10),
	}

	return gitSvc, nil
}
