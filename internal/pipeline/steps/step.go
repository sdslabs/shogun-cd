package pipelineSteps

import (
	"context"
	"fmt"

	"go.yaml.in/yaml/v3"
)

const (
	MutateType = "mutate"
	SyncType   = "sync"
	ExecType   = "exec"
	ApplyType  = "apply"
)

type StepWrapper struct {
	Step Step
}
type Step interface {
	// Returns the type of the step as a string
	Type() string
	// Returns the trigger condition of the step
	Trigger() string
	// Runs the step execution logic
	Execute(ctx context.Context, deps *StepDeps) error
}

// UnmarshalYAML is an interface hook for custom unmarshaling of StepWrapper
// This function name must be exactly "UnmarshalYAML" to be recognized by the YAML package
func (sw *StepWrapper) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping node")
	}
	if len(node.Content) < 2 {
		return fmt.Errorf("invalid step format")
	}
	stepType := node.Content[0].Value
	var step Step
	switch stepType {
	case MutateType:
		step = &MutateStep{}
	case SyncType:
		step = &SyncStep{}
	case ExecType:
		step = &ExecStep{}
	case ApplyType:
		step = &ApplyStep{}
	default:
		return fmt.Errorf("unknown step type: %s", stepType)
	}
	data := node.Content[1]
	if err := data.Decode(step); err != nil {
		return fmt.Errorf("failed to decode step of type %s: %v", stepType, err)
	}
	sw.Step = step
	return nil
}

// MarshalYAML is an interface hook for custom marshaling of StepWrapper
// This function name must be exactly "MarshalYAML" to be recognized by the YAML package
// Currently MarshalYAML is not used anywhere in the codebase, but is added for the sake of completeness
func (sw StepWrapper) MarshalYAML() (interface{}, error) {
	if sw.Step == nil {
		return nil, fmt.Errorf("cannot marshal nil step")
	}

	// Return map with step type as key
	return map[string]interface{}{
		sw.Step.Type(): sw.Step,
	}, nil
}
