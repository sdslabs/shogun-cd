package pipelineSteps

type ApplyStep struct {
	TriggerWhen string   `yaml:"trigger_when,omitempty"`
	Target      string   `yaml:"target"` // [TODO]: change this to pointer if needed
	Files       []string `yaml:"files"`
}

func (*ApplyStep) Type() string {
	return ApplyType
}

func (as *ApplyStep) Trigger() string {
	return as.TriggerWhen
}

func (as *ApplyStep) TargetInstance() string {
	return as.Target
}

func (as *ApplyStep) Execute() {}
