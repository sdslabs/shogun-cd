package git

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type GitService interface {
	// Clones the repo and retries until it succeeds
	Clone(ctx context.Context)
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
	// Get repo URL
	GetRepoURL() string
	// CreateAndAddDeployKey creates an SSH key pair to be added as a deploy key and and adds it to git service
	CreateAndAddDeployKey(keyDir string) error
	// Get Polling Interval
	GetPollingInterval() int
	// CommitAndPushChanges commits and pushes changes to the remote repository with the specified commit message
	CommitAndPushChanges(ctx context.Context, commitMsg string, args ...any) error
}

const (
	defaultDataDir = "var/lib/shogun/"
	cloneSubDir    = "cache"
	keyDir         = "ssh"
)

type Service struct {
	logger          utils.Logger
	repo            Repo
	pollingInterval int

	gitPath        string
	pullEventsChan chan []string
}

type Repo struct {
	URL        string
	Branch     string
	CloneDir   string
	DeployKeys *KeyPair
	// RWMutex for synchronizing exclusive access to the repository
	mu sync.RWMutex
}

type KeyPair struct {
	PrivatePath string // Path to the private key file wrt the clone directory
	PublicPath  string // Path to the public key file wrt the clone directory
}

func NewGitService(logger utils.Logger, Config *config.Config) (GitService, error) {

	gitPath, err := exec.LookPath("git")
	if err != nil {
		logger.LogNewError("Git executable not found in PATH:", err)
		return nil, err
	}
	logger.Log("Git executable found at: %s", gitPath)

	if Config.DataDir == "" {
		Config.DataDir = defaultDataDir
	}

	Config.GitConfig.Repo = strings.TrimSpace(Config.GitConfig.Repo)

	useKeys := Config.GitConfig.CreateDeployKey
	gitURL := strings.HasPrefix(Config.GitConfig.Repo, "git@") || strings.HasPrefix(Config.GitConfig.Repo, "ssh://")
	httpURL := strings.HasPrefix(Config.GitConfig.Repo, "http://") || strings.HasPrefix(Config.GitConfig.Repo, "https://")
	// if deploy keys enabled, must be SSH
	if useKeys && !gitURL {
		logger.LogNewError("To allow Yaml mutations, deploy key creation must be enabled with SSH based repo URLs")
		return nil, fmt.Errorf("Deploy Keys and SSH based repo URLs should be always used together")
	}
	// neither ssh nor HTTP
	if !gitURL && !httpURL {
		logger.LogNewError("Unsupported git repo URL format: %s", Config.GitConfig.Repo)
		return nil, fmt.Errorf("Unsupported git repo URL format")
	}

	gitSvc := &Service{
		logger:          logger,
		pollingInterval: Config.GitConfig.PollingInterval,
		repo: Repo{
			URL:        Config.GitConfig.Repo,
			Branch:     Config.GitConfig.Branch,
			CloneDir:   filepath.Join(Config.DataDir, cloneSubDir),
			DeployKeys: nil,
		},
		gitPath:        gitPath,
		pullEventsChan: make(chan []string, 10),
	}

	if useKeys {
		if err := gitSvc.CreateAndAddDeployKey(filepath.Join(Config.DataDir, keyDir)); err != nil {
			return nil, err
		}
	}

	return gitSvc, nil
}
