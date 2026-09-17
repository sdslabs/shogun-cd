package dto

type PipelineTrigger struct {
	Type  string   `json:"type"`
	Paths []string `json:"paths,omitempty"`
}

type PipelineStepSummary struct {
	Index       int    `json:"index"`
	Type        string `json:"type"`
	TriggerWhen string `json:"trigger_when,omitempty"`
	Target      string `json:"target,omitempty"`
}

type PipelineSummary struct {
	Name     string                `json:"name"`
	Enabled  bool                  `json:"enabled"`
	Triggers []PipelineTrigger     `json:"triggers"`
	Steps    []PipelineStepSummary `json:"steps"`
}

type TargetSummary struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Host string `json:"host"`
	User string `json:"user"`
	Port int    `json:"port"`
}
