package git

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	GHusername = "ShogunCD-bot"
	GHemail    = "shogun.cd.dev@gmail.com"
)

// Clones the repo and retries until it succeeds
func (s *Service) Clone(ctx context.Context) {
	// Lock repo because CloneRepo assumes exclusive access
	s.repo.mu.Lock()
	for {
		err := s.cloneRepo(ctx)
		if err == nil {
			s.logger.Log("Git repository is cloned.")
			break
		}
		s.logger.Log("Clone Failed: Retrying Clone...")
	}
	err := s.setGHaccount()
	if err != nil {
		s.logger.LogNewError("Failed to set Github account: %v", err)
	}
	s.repo.mu.Unlock()
}

// cloneRepo clones the git repository to the specified directory
// If the repository already exists, it validates the remote URL and branch or switches to the correct branch
// If the clone fails, it returns an error
// cloneRepo assumes that the repo mutex is already locked for exclusive access before calling this method
func (s *Service) cloneRepo(ctx context.Context) error {

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

	s.logger.LogInfo("Git user configured: %s <%s>", GHusername, GHemail)
	return nil
}

func (s *Service) ClearCacheDir() error {
	if err := os.RemoveAll(s.repo.CloneDir); err != nil {
		s.logger.LogNewError("Failed to remove directory: %v", err)
		return err
	}
	return nil
}

// Pull pulls the latest changes from the remote repository
// If there are new changes, it sends the list of changed file paths to the pull events channel
// Pull is blocking and should be only called using the poller goroutine
func (s *Service) Pull(ctx context.Context) error {

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

// UnlockRepo unlocks the repository from exclusive access
func (s *Service) UnlockRepo() {
	s.repo.mu.Unlock()
}

// RLockRepo locks the repository for shared access
func (s *Service) RLockRepo() {
	s.repo.mu.RLock()
}

// RUnlockRepo unlocks the repository from shared access
func (s *Service) RUnlockRepo() {
	s.repo.mu.RUnlock()
}

// GetRepoRoot returns the root directory of the repository
func (s *Service) GetRepoRoot() string {
	return s.repo.CloneDir
}

// GetRepoURL returns the URL of the repository
func (s *Service) GetRepoURL() string {
	return s.repo.URL
}

// GetPollingInterval returns the polling interval in seconds
func (s *Service) GetPollingInterval() int {
	return s.pollingInterval
}

// CommitAndPushChanges commits and pushes changes to the remote repository with the specified commit message
// It assumes that the repository is already locked for exclusive access before calling this method
func (s *Service) CommitAndPushChanges(ctx context.Context, commitMsg string, args ...any) (string, error) {
	output := ""
	msg := fmt.Sprintf(commitMsg, args...)
	out, err := s.repo.ExecGitCommand(ctx, s.gitPath, "add", ".")
	if len(out) > 0 {
		// s.logger.LogCustom(utils.Green, "Git-logs", "\n"+string(out))
		output = s.logger.LogShogunGit(output, string(out))
	}
	if err != nil {
		s.logger.LogNewError("Failed to stage changes: %v", err)
		return output, err
	}

	out, err = s.repo.ExecGitCommand(ctx, s.gitPath, "commit", "-m", msg)
	if len(out) > 0 {
		// s.logger.LogCustom(utils.Green, "Git-logs", "\n"+string(out))
		output = s.logger.LogShogunGit(output, string(out))
	}
	if err != nil {
		s.logger.LogNewError("Failed to commit changes: %v", err)
		return output, err
	}

	out, err = s.repo.ExecGitCommand(ctx, s.gitPath, "push", "origin", s.repo.Branch)
	if len(out) > 0 {
		// s.logger.LogCustom(utils.Green, "Git-logs", "\n"+string(out))
		output = s.logger.LogShogunGit(output, string(out))
	}
	if err != nil {
		s.logger.LogNewError("Failed to push changes: %v", err)
		return output, err
	}
	// s.logger.LogInfo("Changes committed and pushed successfully")
	output = s.logger.LogShogunInfo(output, "Changes committed and pushed successfully")
	return output, nil
}
