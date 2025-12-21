package git

import (
	"context"
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

	// Check if repo already exists and is valid
	if s.repo.IsRepoValid(ctx, s.gitPath) {
		s.logger.LogInfo("Valid git repository found at: %s", s.repo.CloneDir)
		s.mu.Lock()
		s.ready = true
		s.mu.Unlock()
		return nil
	}

	// If directory exists but is not a valid repo, remove it
	// if _, err := os.Stat(s.repo.CloneDir); err == nil {
	// 	s.logger.Log("Removing invalid/incomplete clone directory: %s", s.repo.CloneDir)
	// 	if err := os.RemoveAll(s.repo.CloneDir); err != nil {
	// 		s.logger.LogNewError("Failed to remove invalid directory: %v", err)
	// 		return err
	// 	}
	// }

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
