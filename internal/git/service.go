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
	CloneRepo(ctx context.Context, repoURL, branch, destPath string) error
	// Polls the remote repository for changes if clone succeeded outherwise retries clone
	StartPoller(ctx context.Context)
	// Pulls the latest changes from the remote repository
	Pull(ctx context.Context) error
	// Blocks till repo is ready to be used
	WaitReady() error
}

type Service struct {
	logger          utils.Logger
	repo            Repo
	ready           bool
	pollingInterval int

	gitPath string
	mu      sync.Mutex
	pullSF  singleflight.Group
}

type Repo struct {
	URL      string
	Branch   string
	CloneDir string
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
		ready:           false,
		pollingInterval: gitConfig.PollingInterval,
		repo: Repo{
			URL:      gitConfig.Repo,
			Branch:   gitConfig.Branch,
			CloneDir: gitConfig.CloneDir,
		},
		gitPath: gitPath,
		mu:      sync.Mutex{},
		pullSF:  singleflight.Group{},
	}

	ctx := context.Background()
	go func() {
		gitSvc.CloneRepo(ctx, gitConfig.Repo, gitConfig.Branch, gitConfig.CloneDir)
		gitSvc.StartPoller(ctx)
		// [TODO] Start polling if clone else retry clone
	}()

	return gitSvc, nil
}
