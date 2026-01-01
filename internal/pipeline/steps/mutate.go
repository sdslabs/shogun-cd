package pipelineSteps

import (
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

type MutateStep struct {
	TriggerWhen string   `yaml:"trigger_when,omitempty"`
	Changes     []Change `yaml:"changes"`
}

type Change struct {
	File        string `yaml:"file"`
	UpdateField string `yaml:"update_field"`
	Value       string `yaml:"value"`
}

func (*MutateStep) Type() string {
	return MutateType
}

func (ms *MutateStep) Trigger() string {
	return ms.TriggerWhen
}

func (ms *MutateStep) Execute(deps *StepDeps) error {
	deps.GitService.LockRepo()
	defer deps.GitService.UnlockRepo()
	for _, change := range ms.Changes {

		file := filepath.Join(deps.GitService.GetRepoRoot(), change.File)
		f, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		// YAML mutation logic

		var root yaml.Node
		err = yaml.Unmarshal(f, &root)
		if err != nil {
			return err
		}

	}
	return nil
}
