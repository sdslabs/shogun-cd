package pipelineSteps

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/pkg/sftp"
)

// ExecSteps are only valid for target type server

type SyncStep struct {
	TriggerWhen string       `yaml:"trigger_when,omitempty"`
	Target      string       `yaml:"target"` // [TODO]: change this to pointer if needed
	Files       []FileUpdate `yaml:"files"`
}

type FileUpdate struct {
	Dst string `yaml:"dst"`
	Src string `yaml:"src"`
}

func (*SyncStep) Type() string {
	return SyncType
}

func (ss *SyncStep) Trigger() string {
	return ss.TriggerWhen
}

func (ss *SyncStep) TargetInstance() string {
	return ss.Target
}

func (ss *SyncStep) Execute(ctx context.Context, deps *StepDeps) (string, error) {
	deps.GitService.LockRepo()
	defer deps.GitService.UnlockRepo()

	var output string
	// Validate target
	targetInstance, exists := deps.Targets[ss.Target]
	if !exists {
		return deps.Logger.LogShogunError(output, "Target not found: %s", ss.Target)
	}
	if targetInstance.Metadata.Type != target.ServerType {
		return deps.Logger.LogShogunError(output, "Sync step can only be executed on server type targets. Target %s is of type %s", ss.Target, targetInstance.Metadata.Type)
	}

	client, err := deps.SSHManager.GetClient(ss.Target)
	// No client or dead client, create a new one
	if err != nil {
		sshkey, err := deps.SecretService.FetchSecret(ctx, targetInstance.Spec.AccessSecret)
		if err != nil {
			return deps.Logger.LogShogunError(output, "Failed to fetch secret for target %s: %v", ss.Target, err)
		}
		client, err = deps.SSHManager.NewClient(targetInstance.Metadata.Name, targetInstance.Spec.Host, targetInstance.Spec.Port, targetInstance.Spec.User, []byte(sshkey))
		if err != nil {
			return deps.Logger.LogShogunError(output, "Failed to create SSH client for target %s: %v", ss.Target, err)
		}
	}

	sftpclient, err := sftp.NewClient(client)
	if err != nil {
		return deps.Logger.LogShogunError(output, "Failed to create SFTP client for target %s: %v", ss.Target, err)
	}
	defer sftpclient.Close()

	for _, file := range ss.Files {
		src := InterpolateVariables(file.Src, deps.HookValues)
		src, err = deps.SecretService.ResolveSecrets(ctx, src)
		if err != nil {
			return deps.Logger.LogShogunError(output, "Failed to resolve secrets for source path %s: %v", file.Src, err)
		}
		dst := InterpolateVariables(file.Dst, deps.HookValues)
		dst, err = deps.SecretService.ResolveSecrets(ctx, dst)
		if err != nil {
			return deps.Logger.LogShogunError(output, "Failed to resolve secrets for destination path %s: %v", file.Dst, err)
		}

		srcFile, err := os.Open(filepath.Join(deps.GitService.GetRepoRoot(), src))
		if err != nil {
			return deps.Logger.LogShogunError(output, "Failed to read file %s: %v", file.Src, err)
		}

		dir := path.Dir(dst)
		if err := sftpclient.MkdirAll(dir); err != nil {
			return deps.Logger.LogShogunError(output, "Failed to create remote directory %s: %v", dir, err)
		}
		tmpFile, err := sftpclient.Create(dst + ".tmp")
		if err != nil {
			return deps.Logger.LogShogunError(output, "Failed to create remote temp file %s: %v", file.Dst+".tmp", err)
		}
		_, err = io.Copy(tmpFile, srcFile)
		if err != nil {
			return deps.Logger.LogShogunError(output, "Failed to copy file %s to %s: %v", file.Src, file.Dst+".tmp", err)
		}
		srcFile.Close()
		tmpFile.Close()
		_ = sftpclient.Remove(dst)
		err = sftpclient.Rename(dst+".tmp", dst)
		if err != nil {
			return deps.Logger.LogShogunError(output, "Failed to rename remote file %s: %v", file.Dst+".tmp", err)
		}
		output = deps.Logger.LogShogunInfo(output, "Successfully synced file %s to %s", file.Src, file.Dst)
	}

	deps.Logger.Log("Executed sync step")
	return output, nil
}
