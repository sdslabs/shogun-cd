package pipelineSteps

import (
	"context"
	"strings"

	"github.com/kunalvirwal/shogun-cd/internal/target"
	"golang.org/x/crypto/ssh"
)

// ExecSteps are only valid for target type server

type ExecStep struct {
	TriggerWhen string   `yaml:"trigger_when,omitempty"`
	Target      string   `yaml:"target"` // [TODO]: change this to pointer if needed
	Commands    []string `yaml:"commands"`
}

func (*ExecStep) Type() string {
	return ExecType
}

func (es *ExecStep) Trigger() string {
	return es.TriggerWhen
}

func (es *ExecStep) TargetInstance() string {
	return es.Target
}

func (es *ExecStep) Execute(ctx context.Context, deps *StepDeps) (string, error) {
	var output string
	// Validate target
	targetInstance, exists := deps.Targets[es.Target]
	if !exists {
		return deps.Logger.LogShogunError(output, "Target not found: %s", es.Target)
	}
	if targetInstance.Metadata.Type != target.ServerType {
		return deps.Logger.LogShogunError(output, "Exec step can only be executed on server type targets. Target %s is of type %s", es.Target, targetInstance.Metadata.Type)
	}

	client, err := deps.SSHManager.GetClient(es.Target)
	// No client or dead client, create a new one
	if err != nil {
		sshkey, err := deps.SecretService.FetchSecret(targetInstance.Spec.AccessSecret)
		if err != nil {
			return deps.Logger.LogShogunError(output, "Failed to fetch secret for target %s: %v", es.Target, err)
		}
		client, err = deps.SSHManager.NewClient(targetInstance.Metadata.Name, targetInstance.Spec.Host, targetInstance.Spec.Port, targetInstance.Spec.User, []byte(sshkey))
		if err != nil {
			return deps.Logger.LogShogunError(output, "Failed to create SSH client for target %s: %v", es.Target, err)
		}
	}

	// create command script
	var script strings.Builder

	script.WriteString("set -eux\n")
	for _, cmd := range es.Commands {

		cmd = InterpolateVariables(cmd, deps.HookValues)
		cmd = deps.SecretService.ResolveSecrets(cmd)

		script.WriteString(cmd)
		script.WriteByte('\n')

	}

	session, err := client.NewSession()
	if err != nil {
		return deps.Logger.LogShogunError(output, "Failed to create SSH session for target %s: %v", es.Target, err)
	}

	// [TODO]: Use these for streaming output
	// stdout, err := session.StdoutPipe()
	// if err != nil {
	// 	return fmt.Errorf("Failed to get stdout pipe for target %s: %v", es.Target, err)
	// }
	// stderr, err := session.StderrPipe()
	// if err != nil {
	// 	return fmt.Errorf("Failed to get stderr pipe for target %s: %v", es.Target, err)
	// }
	done := make(chan error, 1)
	var out []byte
	var execErr error

	go func() {
		defer close(done)
		defer session.Close()
		out, execErr = session.CombinedOutput(
			"/bin/sh -s <<'SHOGUN_EOF'\n" +
				script.String() +
				"SHOGUN_EOF\n",
		)
	}()

	select {
	// context done case
	case <-ctx.Done():
		return deps.Logger.LogShogunError(output, "Command execution stopped by context cancellation")

	// exec done case
	case <-done:
		if execErr != nil {
			if exitErr, ok := execErr.(*ssh.ExitError); ok {
				return deps.Logger.LogShogunError(output, "Command execution failed on target %s: %v, output: \n %s", es.Target, exitErr, string(out))
			}

			return deps.Logger.LogShogunError(output, "SSH session error on target %s: %v", es.Target, err)
		}
	}

	output = deps.Logger.LogShogunInfo(output, "Command output: \n%s", string(out))
	deps.Logger.Log("Executed exec step")
	return output, nil
}
