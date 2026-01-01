package pipelineSteps

import (
	"regexp"
	"strings"

	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

// StepDeps represents dependencies required by pipeline steps.
type StepDeps struct {
	// [TODO] Git and SSH
	Logger     utils.Logger
	GitService git.GitService
	HookValues map[string]string
}

// InterpolateVariables replaces {{...}} patterns in the input string with values from the variables map
func InterpolateVariables(input string, variables map[string]string) string {
	re := regexp.MustCompile(`(?s)\{\{(.*?)\}\}`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract the key from {{KEY}}
		key := strings.TrimSpace(match[2 : len(match)-2])

		// Look up the value in the variables map
		if value, exists := variables[key]; exists {
			return value
		}

		// If not found, return the original placeholder
		return match
	})
}
