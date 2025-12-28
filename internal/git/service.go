package git

import (
	"context"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
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
	// Locks the repo for exclusive access
	LockRepo()
	// Unlocks the repo
	UnlockRepo()
	// RLocks the repo for shared access
	RLockRepo()
	// RUnlocks the repo
	RUnlockRepo()
	// Get repo details
	GetRepoRoot() string
}

type Service struct {
	logger          utils.Logger
	repo            Repo
	pollingInterval int

	gitPath        string
	pullbusy       atomic.Bool
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
		pullEventsChan: make(chan []string, 10),
	}

	return gitSvc, nil
}
