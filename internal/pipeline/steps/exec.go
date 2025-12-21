package pipelineSteps

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

func (es *ExecStep) Execute(deps *StepDeps) error {
	return nil
}
