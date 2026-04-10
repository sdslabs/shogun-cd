package models

import (
	"time"

	"github.com/google/uuid"
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
	ID          uuid.UUID  `gorm:"column:id; type:uuid; default:gen_random_uuid(); primaryKey"`
	Pipeline    string     `gorm:"column:pipeline; index; not null; type:varchar(50)"`
	TriggerKind string     `gorm:"column:trigger_kind; not null; type:varchar(32)"`
	Success     *bool      `gorm:"column:success"`
	StartedAt   time.Time  `gorm:"column:started_at; not null; index"`
	FinishedAt  *time.Time `gorm:"column:finished_at; index"`
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
	RunID       uuid.UUID   `gorm:"column:run_id; type:uuid; not null; primaryKey"`
	StepIndex   int         `gorm:"column:step_index; not null; primaryKey"`
	StepType    string      `gorm:"column:step_type; type:varchar(32); not null"`
	Status      StepStatus  `gorm:"column:status; type: varchar(32); not null; default:failed"`
	Logs        string      `gorm:"column:logs; type:text"`
	StartedAt   time.Time   `gorm:"column:started_at; index"`
	FinishedAt  *time.Time  `gorm:"column:finished_at; index"`
	PipelineRun PipelineRun `gorm:"foreignKey:RunID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (PipelineRunStep) TableName() string {
	return PipelineRunStepTableName
}

// combined bundle consisting of a complete run's details
type PipelineRunDetails struct {
	Run   PipelineRun
	Steps []PipelineRunStep
}
