package git

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"
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

	s.logger.LogInfo("Starting git poller for repository: %s", s.repo.URL)

	ticker := time.NewTicker(time.Duration(s.pollingInterval) * time.Second)
	for range ticker.C {
		s.logger.Log("Polling for changes...")
		go func() {
			if err := s.Pull(ctx); err != nil {
				s.logger.LogNewError("Git pull failed: ", err)
			}
		}()
	}
}

func (s *Service) Pull(ctx context.Context) error {
	// Use singleflight to prevent concurrent pulls
	fileDiffs, err, _ := s.pullSF.Do("git-pull", func() (any, error) {

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

		s.repo.mu.Lock()
		defer s.repo.mu.Unlock()

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

	})

	fileDiffSlice, ok := fileDiffs.([]string)
	if !ok {
		fileDiffSlice = []string{}
	}

	if len(fileDiffSlice) > 0 {
		select {
		case s.pullEventsChan <- fileDiffSlice:
		default:
			s.logger.LogNewError("Pull events channel is full; skipping sending pull event")
		}
	}

	// [TODO] Add support for file diff triggered pipeline execution
	return err
}

// func (s *Service) WaitReady() error {
// 	return nil
// }

func (s *Service) GetPullEvents() <-chan []string {
	return s.pullEventsChan
}
