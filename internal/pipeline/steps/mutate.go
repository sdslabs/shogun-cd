package pipelineSteps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

// Only POSIX compliant
func (ms *MutateStep) Execute(ctx context.Context, deps *StepDeps) error {
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

		change.Value = InterpolateVariables(change.Value, deps.HookValues)
		change.UpdateField = InterpolateVariables(change.UpdateField, deps.HookValues)
		fieldPath := strings.Split(change.UpdateField, ".")
		err = mutateYAMLField(&root, fieldPath, change.Value)
		if err != nil {
			return err
		}

		dir := filepath.Dir(file)
		tmp, err := os.CreateTemp(dir, "mutate-*.yaml")
		if err != nil {
			return err
		}

		enc := yaml.NewEncoder(tmp)
		enc.SetIndent(2)
		err = enc.Encode(&root)
		if err != nil {
			return err
		}
		enc.Close()
		tmp.Close()

		// This is only Posix compliant as the original file pre-exists an renaming on windows doesn't overwrite it
		err = os.Rename(tmp.Name(), file)
		if err != nil {
			return err
		}

	}

	deps.Logger.Log("Yaml successfully mutated")
	deps.GitService.CommitAndPushChanges(ctx, "[TEST] Replaced values via pipeline \"%s\"", deps.PipelineName)

	return nil
}

// svc.web.volumes[0].name

func mutateYAMLField(node *yaml.Node, fieldPath []string, value string) error {
	var n *yaml.Node
	if node.Kind == yaml.DocumentNode {
		n = node.Content[0]
	} else {
		n = node
	}
	for i := 0; i < len(fieldPath); {
		found := false
		field, index := parseIndexed(fieldPath[i])
		if n.Kind == yaml.MappingNode {
			for j := 0; j < len(n.Content); j += 2 {
				// key matches
				if n.Content[j].Value == field {
					if i == len(fieldPath)-1 {
						if index != -1 {
							if n.Content[j+1].Kind != yaml.SequenceNode {
								return fmt.Errorf("Field %v is not a sequence", fieldPath[i])
							}
							if index >= len(n.Content[j+1].Content) {
								return fmt.Errorf("Index %d out of bounds for field %v", index, fieldPath[i])
							}
							if n.Content[j+1].Content[index].Kind != yaml.ScalarNode {
								return fmt.Errorf("Field %v is not a scalar", fieldPath[i])
							}
							// Update value
							fmt.Println("modifying", n.Content[j+1].Content[index].Value, "to", value)
							n.Content[j+1].Content[index].Value = value
							return nil
						}
						// Update value
						if n.Content[j+1].Kind != yaml.ScalarNode {
							return fmt.Errorf("Field %v is not a scalar", fieldPath[i])
						}
						n.Content[j+1].Value = value
						return nil
					} else {
						if index != -1 {
							if n.Content[j+1].Kind != yaml.SequenceNode {
								return fmt.Errorf("Field %v is not a sequence", fieldPath[i])
							}
							if index >= len(n.Content[j+1].Content) {
								return fmt.Errorf("Index %d out of bounds for field %v", index, fieldPath[i])
							}
							// Traverse deeper
							n = n.Content[j+1].Content[index]
							found = true
							break
						}
						if n.Content[j+1].Kind != yaml.MappingNode {
							return fmt.Errorf("Field %v is not a mapping", fieldPath[i])
						}
						// Traverse deeper
						n = n.Content[j+1]
						found = true
						break
					}
				}
			}
		} else if n.Kind == yaml.ScalarNode && i == len(fieldPath)-1 {
			n.Value = value
			return nil
		} else if n.Kind == yaml.SequenceNode {
			if index == -1 || index >= len(n.Content) {
				return fmt.Errorf("Invalid index in field %v", fieldPath[i])
			}
			n = n.Content[index]
			found = true
		}

		if !found {
			return fmt.Errorf("Field %v not found", fieldPath[i])
		}
		i++
	}
	return nil
}

func parseIndexed(s string) (string, int) {
	open := strings.IndexByte(s, '[')
	close := strings.IndexByte(s, ']')
	if open == -1 || close == -1 || close <= open+1 {
		return s, -1
	}
	index, err := strconv.Atoi(s[open+1 : close])
	if err != nil {
		return s, -1
	}
	return s[:open], index
}
