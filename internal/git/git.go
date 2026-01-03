package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (r *Repo) ExecGitCommand(ctx context.Context, gitPath string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, gitPath, args...)

	// Only set working directory for non-clone commands
	// git clone creates the directory, so we can't cd into it first
	if len(args) > 0 && args[0] != "clone" {
		cmd.Dir = r.CloneDir
	}

	if r.DeployKeys != nil {
		cmd.Env = append(os.Environ(),
			"GIT_SSH_COMMAND=ssh -i "+r.DeployKeys.PrivatePath+" -o IdentitiesOnly=yes -F /dev/null -o StrictHostKeyChecking=no",
		)
	} else {
		cmd.Env = append(os.Environ(),
			"GIT_SSH_COMMAND=ssh -o IdentitiesOnly=yes -o IdentityFile=/dev/null -F /dev/null -o StrictHostKeyChecking=no",
		)
	}

	// [TODO] Change this to streaming output
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	return out, err
}

// IsRepoValid checks if the clone directory exists and is a valid git repository with intact working tree
func (r *Repo) IsRepoValid(ctx context.Context, gitPath string) bool {
	// Check if directory exists
	if _, err := os.Stat(r.CloneDir); os.IsNotExist(err) {
		return false
	}

	// Check if .git directory exists
	gitDir := filepath.Join(r.CloneDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false
	}

	return true
}
