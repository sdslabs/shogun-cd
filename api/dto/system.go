package dto

import "time"

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
	LastRun  *PipelineRunSummary   `json:"last_run"`
}

// PipelineRunSummary is the small execution snapshot needed by the pipeline
// index. Full step details remain available through the pipeline run routes.
type PipelineRunSummary struct {
	ID          uint       `json:"id"`
	TriggerKind string     `json:"trigger_kind"`
	Status      string     `json:"status"`
	Success     *bool      `json:"success"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
}

type TargetSummary struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Host string `json:"host"`
	User string `json:"user"`
	Port int    `json:"port"`
}
