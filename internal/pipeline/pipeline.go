package pipeline

import pipelineSteps "github.com/kunalvirwal/shogun-cd/internal/pipeline/steps"

type Kind string
type TriggerKind string

const (
	PipelineKind          Kind        = "Pipeline"
	WebhookTriggerKind    TriggerKind = "ci_webhook"
	GitChangesTriggerKind TriggerKind = "git_changes"
)

type Pipeline struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       Kind     `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Metadata struct {
	Name    string `yaml:"name"`
	Enabled bool   `yaml:"enabled"`
}

type Spec struct {
	Triggers []Trigger                   `yaml:"triggers"`
	Steps    []pipelineSteps.StepWrapper `yaml:"steps"`
}

type Trigger struct {
	Type  string   `yaml:"type"`
	Paths []string `yaml:"paths,omitempty"`
}
