package models

import (
	"time"
)

// NOTE : the values of constants must be consistent to the "column" in gorm tags.
const (
	// table name
	PipelineRunTableName = "pipeline_runs"
	// columns
	PipelineRunColID          = "id"
	PipelineRunColPipeline    = "pipeline"
	PipelineRunColTriggerKind = "trigger_kind"
	PipelineRunColSuccess     = "success"
	PipelineRunColStartedAt   = "started_at"
	PipelineRunColFinishedAt  = "finished_at"
)

type PipelineRun struct {
	ID          uint              `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Pipeline    string            `gorm:"column:pipeline; index; not null; type:varchar(50)" json:"pipeline"`
	TriggerKind string            `gorm:"column:trigger_kind; not null; type:varchar(32)" json:"trigger_kind"`
	Success     *bool             `gorm:"column:success" json:"success"`
	StartedAt   time.Time         `gorm:"column:started_at; not null; index" json:"started_at"`
	FinishedAt  *time.Time        `gorm:"column:finished_at; index" json:"finished_at,omitempty"`
	Steps       []PipelineRunStep `gorm:"foreignKey:RunID;references:ID" json:"steps"`
}

func (PipelineRun) TableName() string {
	return PipelineRunTableName
}

// NOTE : the values of constants must be consistent to the "column" in gorm tags.
const (
	// table name
	PipelineRunStepTableName = "pipeline_run_steps"
	// columns
	PipelineRunStepColRunID      = "run_id"
	PipelineRunStepColStepIndex  = "step_index"
	PipelineRunStepColStepType   = "step_type"
	PipelineRunStepColStatus     = "status"
	PipelineRunStepColLogs       = "logs"
	PipelineRunStepColStartedAt  = "started_at"
	PipelineRunStepColFinishedAt = "finished_at"
)

type StepStatus string

const (
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusSucceeded  StepStatus = "succeeded"
	StepStatusFailed     StepStatus = "failed"
	StepStatusSkipped    StepStatus = "skipped"
)

func (s StepStatus) IsValid() bool {
	switch s {
	case StepStatusInProgress,
		StepStatusSucceeded,
		StepStatusFailed,
		StepStatusSkipped:
		return true
	default:
		return false
	}
}

type PipelineRunStep struct {
	RunID      uint       `gorm:"column:run_id;not null;primaryKey" json:"-"`
	StepIndex  int        `gorm:"column:step_index; not null; primaryKey" json:"step_index"`
	StepType   string     `gorm:"column:step_type; type:varchar(32); not null" json:"step_type"`
	Status     StepStatus `gorm:"column:status; type: varchar(32); not null; default:failed" json:"status"`
	Logs       string     `gorm:"column:logs; type:text" json:"logs"`
	StartedAt  time.Time  `gorm:"column:started_at; index" json:"started_at"`
	FinishedAt *time.Time `gorm:"column:finished_at; index" json:"finished_at"`
}

func (PipelineRunStep) TableName() string {
	return PipelineRunStepTableName
}
