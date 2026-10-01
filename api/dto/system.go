package dto

import "time"

type PipelineTrigger struct {
	Type  string   `json:"type"`
	Paths []string `json:"paths,omitempty"`
}

type PipelineStepSummary struct {
	Index       int            `json:"index"`
	Type        string         `json:"type"`
	TriggerWhen string         `json:"trigger_when,omitempty"`
	Target      string         `json:"target,omitempty"`
	Config      map[string]any `json:"config"`
}

// PipelineSyncStepDTO is needed because sync files are structured src/dst
// objects, unlike apply files and exec commands which are plain string lists.
type PipelineSyncStepDTO struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

// PipelineMutateStepDTO is needed because each mutate change contains three
// named fields, so it requires a stable JSON object for the frontend.
type PipelineMutateStepDTO struct {
	File        string `json:"file"`
	UpdateField string `json:"update_field"`
	Value       string `json:"value"`
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
	Name         string `json:"name"`
	Type         string `json:"type"`
	Host         string `json:"host"`
	User         string `json:"user"`
	Port         int    `json:"port"`
	AccessSecret string `json:"access_secret"`
}
