package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
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
	// CreateAndAddDeployKey creates an SSH key pair to be added as a deploy key and and adds it to git service
	CreateAndAddDeployKey(keyDir string) error
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
	URL        string
	Branch     string
	CloneDir   string
	DeployKeys *KeyPair
	mu         sync.RWMutex // Mutex for synchronizing exclusive access to the repository
}

type KeyPair struct {
	PrivatePath string
	PublicPath  string
}

func NewGitService(logger utils.Logger, gitConfig config.Git) (GitService, error) {

	gitPath, err := exec.LookPath("git")
	if err != nil {
		logger.LogNewError("Git executable not found in PATH:", err)
		return nil, err
	}
	logger.Log("Git executable found at: %s", gitPath)

	if gitConfig.CloneDir == "" {
		gitConfig.CloneDir = ".data/cache/"
	}

	gitConfig.Repo = strings.TrimSpace(gitConfig.Repo)

	useKeys := gitConfig.CreateDeployKey
	gitURL := strings.HasPrefix(gitConfig.Repo, "git@") || strings.HasPrefix(gitConfig.Repo, "ssh://")
	httpURL := strings.HasPrefix(gitConfig.Repo, "http://") || strings.HasPrefix(gitConfig.Repo, "https://")
	// if deploy keys enabled, must be SSH
	if useKeys && !gitURL {
		logger.LogNewError("To allow Yaml mutations, deploy key creation must be enabled with SSH based repo URLs")
		return nil, fmt.Errorf("Deploy Keys and SSH based repo URLs should be always used together")
	}
	// neither ssh nor HTTP
	if !gitURL && !httpURL {
		logger.LogNewError("Unsupported git repo URL format: %s", gitConfig.Repo)
		return nil, fmt.Errorf("Unsupported git repo URL format")
	}

	gitSvc := &Service{
		logger:          logger,
		pollingInterval: gitConfig.PollingInterval,
		repo: Repo{
			URL:        gitConfig.Repo,
			Branch:     gitConfig.Branch,
			CloneDir:   gitConfig.CloneDir,
			DeployKeys: nil,
		},
		gitPath:        gitPath,
		pullEventsChan: make(chan []string, 10),
	}

	if useKeys {
		if err := gitSvc.CreateAndAddDeployKey(gitConfig.KeyDir); err != nil {
			return nil, err
		}
	}

	return gitSvc, nil
}
