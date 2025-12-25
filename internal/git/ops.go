package git

import (
	"context"
	"os"
	"strings"
	"time"
)

func (s *Service) CloneRepo(ctx context.Context, repoURL, branch, destPath string) error {
	s.mu.Lock()
	if s.ready {
		s.logger.Log("Repository already cloned and ready.")
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

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
							return err
						}
					if _, err := s.repo.ExecGitCommand(ctx, s.gitPath, "checkout", "-B", s.repo.Branch, "origin/"+s.repo.Branch); err != nil {
							s.logger.LogNewError("Failed to checkout branch: %v", err)
							return err
						}
						s.logger.LogInfo("Switched to branch: %s", s.repo.Branch)
					}
				}

				s.mu.Lock()
				s.ready = true
				s.mu.Unlock()
				return nil
			}
		}
	}

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

	s.mu.Lock()
	s.ready = true
	s.mu.Unlock()

	return nil
}

func (s *Service) ClearCacheDir() error {
	if err := os.RemoveAll(s.repo.CloneDir); err != nil {
		s.logger.LogNewError("Failed to remove directory: %v", err)
		return err
	}
	return nil
}

func (s *Service) StartPoller(ctx context.Context) {
	s.mu.Lock()
	ready := s.ready
	s.mu.Unlock()
	if !ready {
		for {
			s.logger.Log("Clone Failed: Retrying Clone...")
			err := s.CloneRepo(ctx, s.repo.URL, s.repo.Branch, s.repo.CloneDir)
			if err == nil {
				break
			}
		}
	}

	s.logger.LogInfo("Starting git poller for repository: %s", s.repo.URL)

	ticker := time.NewTicker(time.Duration(s.pollingInterval) * time.Second) // [TODO] Make polling interval configurable

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
	return nil
}

func (s *Service) WaitReady() error {
	return nil
}
