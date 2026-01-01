package git

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	GHusername = ""
	GHemail    = ""
)

// CloneRepo clones the git repository to the specified directory
// If the repository already exists, it validates the remote URL and branch or switches to the correct branch
// If the clone fails, it returns an error
func (s *Service) CloneRepo(ctx context.Context) error {

	// Check if repo exists and validate remote URL and branch
	if s.repo.IsRepoValid(ctx, s.gitPath) {

		s.logger.Log("Repository found at: %s, validating remote and branch...", s.repo.CloneDir)
		currentRemote, err := s.repo.ExecGitCommand(ctx, s.gitPath, "remote", "get-url", "origin")
		if err != nil {
			s.logger.Log("Removing directory due to remote check failure: %v", err)
			s.ClearCacheDir()
		} else {
			remoteURL := strings.TrimSpace(string(currentRemote))
			if remoteURL != s.repo.URL {
				s.logger.Log("Remote URL mismatch. Expected: %s, Got: %s", s.repo.URL, remoteURL)
				s.ClearCacheDir()
			} else {
				currentBranch, err := s.repo.ExecGitCommand(ctx, s.gitPath, "rev-parse", "--abbrev-ref", "HEAD")
				if err != nil {
					s.logger.LogNewError("Failed to get current branch: %v", err)
					s.ClearCacheDir()
				} else {

					branch := strings.TrimSpace(string(currentBranch))
					if branch != s.repo.Branch {
						s.logger.Log("Branch mismatch. Expected: %s, Got: %s. Checking out correct branch...", s.repo.Branch, branch)
						if _, err := s.repo.ExecGitCommand(ctx, s.gitPath, "fetch", "origin"); err != nil {
							s.logger.LogNewError("Failed to fetch from origin: %v", err)
							s.ClearCacheDir()
							return err
						}
						if _, err := s.repo.ExecGitCommand(ctx, s.gitPath, "checkout", "-B", s.repo.Branch, "origin/"+s.repo.Branch); err != nil {
							s.logger.LogNewError("Failed to checkout branch: %v", err)
							s.ClearCacheDir()
							return err
						}
						s.logger.LogInfo("Switched to branch: %s", s.repo.Branch)
					}
					_, err := s.repo.ExecGitCommand(ctx, s.gitPath, "reset", "--hard", "origin/"+s.repo.Branch)
					if err != nil {
						s.logger.LogNewError("Failed to pull from origin: %v", err)
						s.ClearCacheDir()
						return err
					}
					// An older version exists now pull will take care of updating it
					return nil
				}
			}
		}
	}

	s.logger.Log("Github account configured")

	// Fresh clone
	s.logger.LogInfo("Cloning git repository: %s branch: %s to dir: %s", s.repo.URL, s.repo.Branch, s.repo.CloneDir)
	output, err := s.repo.ExecGitCommand(ctx, s.gitPath, "clone", "-b", s.repo.Branch, s.repo.URL, s.repo.CloneDir)
	if err != nil {
		s.logger.LogNewError("Git clone failed: %v", err)
		if len(output) > 0 {
			s.logger.Log("Git output: %s", string(output))
		}
		return err
	}

	if len(output) > 0 {
		s.logger.Log("Git clone output: %s", string(output))
	}
	s.logger.LogInfo("Repository cloned successfully")

	err = s.setGHaccount()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) setGHaccount() error {

	_, err := s.repo.ExecGitCommand(context.Background(), s.gitPath, "config", "user.name", GHusername)
	if err != nil {
		s.logger.LogNewError("Failed to set git username: %v", err)
		return err
	}

	_, err = s.repo.ExecGitCommand(context.Background(), s.gitPath, "config", "user.email", GHemail)
	if err != nil {
		s.logger.LogNewError("Failed to set git email: %v", err)
		return err
	}

	s.logger.LogInfo("Git user configured")
	return nil
}

func (s *Service) ClearCacheDir() error {
	if err := os.RemoveAll(s.repo.CloneDir); err != nil {
		s.logger.LogNewError("Failed to remove directory: %v", err)
		return err
	}
	return nil
}

func (s *Service) CloneAndStartPoller(ctx context.Context) {
	s.repo.mu.Lock()
	for {
		err := s.CloneRepo(ctx)
		if err == nil {
			s.logger.Log("Git repository is cloned.")
			break
		}
		s.logger.Log("Clone Failed: Retrying Clone...")
	}
	s.repo.mu.Unlock()

	go func() {

		s.logger.LogInfo("Starting git poller for repository: %s", s.repo.URL)

		ticker := time.NewTicker(time.Duration(s.pollingInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				go func() {
					if err := s.Pull(ctx); err != nil {
						s.logger.LogNewError("Git pull failed: ", err)
					}
				}()
			case <-ctx.Done():
				s.logger.LogInfo("Git poller stopped.")
				return
			}
		}
	}()
}

// Pull pulls the latest changes from the remote repository
// If there are new changes, it sends the list of changed file paths to the pull events channel
// Pull is thread-safe and can be called concurrently
func (s *Service) Pull(ctx context.Context) error {
	// Use atomic.bool and CAS to ensure only one pull at a time
	if !s.pullbusy.CompareAndSwap(false, true) {
		s.logger.Log("Pull already in progress, skipping...")
		return nil
	}

	fileDiffs, err := func() ([]string, error) {

		s.repo.mu.Lock()
		defer s.repo.mu.Unlock()

		s.logger.Log("Fetching repo changes...")
		if _, err := s.repo.ExecGitCommand(ctx, s.gitPath, "fetch", "origin"); err != nil {
			s.logger.LogNewError("Failed to fetch from origin: %v", err)
			return nil, err
		}

		oldHEAD, err := s.repo.ExecGitCommand(ctx, s.gitPath, "rev-parse", "HEAD")
		if err != nil {
			s.logger.LogNewError("Failed to get current HEAD: %v", err)
			return nil, err
		}

		out, err := s.repo.ExecGitCommand(ctx, s.gitPath, "rev-list", "--count", "HEAD...origin/"+s.repo.Branch)
		if err != nil {
			s.logger.LogNewError("Failed to check for new commits: %v", err)
			return nil, err
		}

		diff, _ := strconv.Atoi(strings.TrimSpace(string(out)))
		if diff == 0 {
			// s.logger.Log("No new commits found.")
			return nil, nil
		}
		s.logger.Log("New commits found: %d. Pulling changes...", diff)

		output, err := s.repo.ExecGitCommand(ctx, s.gitPath, "reset", "--hard", "origin/"+s.repo.Branch)
		if err != nil {
			s.logger.LogNewError("Failed to pull from origin: %v", err)
			return nil, err
		}

		if len(output) > 0 {
			s.logger.Log("Pull output: %s", string(output))
		}
		s.logger.LogInfo("Pull completed successfully")

		fileDiffs := []string{}
		diffs, err := s.repo.ExecGitCommand(ctx, s.gitPath, "diff", "--name-only", strings.TrimSpace(string(oldHEAD)), "HEAD")
		if err != nil {
			s.logger.LogNewError("Failed to get file diffs: %v", err)
			return nil, err
		}
		s.logger.Log("Files changed:\n%s", string(diffs))
		fileDiffs = strings.Split(strings.TrimSpace(string(diffs)), "\n")

		return fileDiffs, nil

	}()

	s.pullbusy.Store(false)

	if len(fileDiffs) > 0 {
		select {
		case s.pullEventsChan <- fileDiffs:
		default:
			s.logger.LogNewError("Pull events channel is full; skipping sending pull event")
		}
	}

	return err
}

// GetPullEvents returns a channel that emits pull events with changed file paths
func (s *Service) GetPullEvents() <-chan []string {
	return s.pullEventsChan
}

// LockRepo locks the repository for exclusive access
func (s *Service) LockRepo() {
	s.repo.mu.Lock()
}

// UnlockRepo unlocks the repository
func (s *Service) UnlockRepo() {
	s.repo.mu.Unlock()
}

func (s *Service) RLockRepo() {
	s.repo.mu.RLock()
}

func (s *Service) RUnlockRepo() {
	s.repo.mu.RUnlock()
}

// GetRepoRoot returns the root directory of the repository
func (s *Service) GetRepoRoot() string {
	return s.repo.CloneDir
}
