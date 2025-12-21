package pipelineSteps

type MutateStep struct {
	TriggerWhen string `yaml:"trigger_when,omitempty"`
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
	return nil
}
