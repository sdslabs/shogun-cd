package pipelineSteps

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/secrets"
	"github.com/kunalvirwal/shogun-cd/internal/sshclient"
	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

var varRegex = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*)\s*\}\}`)

type valueSource func(ctx context.Context, key string) (string, bool, error)

func mapSource(prefix string, src map[string]string) valueSource {
	return func(ctx context.Context, key string) (string, bool, error) {
		pre, token, _ := strings.Cut(key, ".")

		if pre != prefix {
			return "", false, nil
		}

		if val, exists := src[token]; exists {
			return val, exists, nil
		}

		return "", false, fmt.Errorf("no such variable exists under `%s` values", prefix)
	}
}

func secretSource(src secrets.SecretService) valueSource {
	return func(ctx context.Context, key string) (string, bool, error) {
		pre, token, _ := strings.Cut(key, ".")

		if pre != "secrets" {
			return "", false, nil
		}

		val, err := src.FetchSecret(ctx, token)

		if err != nil {
			return "", false, fmt.Errorf("no such variable exists under `secrets` values")
		}

		return val, true, nil
	}
}

// StepDeps represents dependencies required by pipeline steps.
type StepDeps struct {
	// [TODO] Add SSH
	PipelineName  string
	Logger        utils.Logger
	GitService    git.GitService
	SecretService secrets.SecretService
	Targets       map[string]*target.Target
	TriggerValues map[string]string
	RunParameters map[string]string
	SSHManager    sshclient.SSHManager
}

func (deps *StepDeps) ResolveString(ctx context.Context, input string) (string, error) {
	sources := []valueSource{
		mapSource("run", deps.RunParameters),
		mapSource("payload", deps.TriggerValues),
		secretSource(deps.SecretService),
	}

	var out strings.Builder
	last := 0

	for _, loc := range varRegex.FindAllStringSubmatchIndex(input, -1) {
		out.WriteString(input[last:loc[0]])

		// 2 and 3 stores the location for the first match group i.e. our key
		key := input[loc[2]:loc[3]]

		resolved := false
		for _, source := range sources {
			val, exists, err := source(ctx, key)
			if err != nil {
				return input, err
			}
			if exists {
				out.WriteString(val)
				resolved = true
				break
			}
		}

		// If not resolved, write back the original {{...}} value
		if !resolved {
			out.WriteString(input[loc[0]:loc[1]])
		}

		last = loc[1]
	}

	out.WriteString(input[last:])

	return out.String(), nil
}
