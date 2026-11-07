package models

import "time"

// CronRunContextKey carries the dispatch record ID into pipeline persistence.
// The pipeline store creates the run and its cron link in one transaction.
type CronRunContextKey struct{}

// Cron definitions live in Shogun; their pipeline reference is resolved from Git.
type Cron struct {
	ID        uint              `gorm:"primaryKey" json:"id"`
	Name      string            `gorm:"not null;type:varchar(100)" json:"name"`
	Pipeline  string            `gorm:"not null;type:varchar(50)" json:"pipeline"`
	Schedule  string            `gorm:"not null;type:varchar(200)" json:"schedule"`
	Payload   map[string]string `gorm:"serializer:json;type:jsonb;not null" json:"payload,omitempty"`
	Enabled   bool              `gorm:"not null" json:"enabled"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty"`
}

// CronRun records the single dispatch attempt, including failures before a pipeline run exists.
type CronRun struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	CronID        uint              `gorm:"not null;index" json:"cron_id"`
	Pipeline      string            `gorm:"not null" json:"pipeline"`
	Payload       map[string]string `gorm:"serializer:json;type:jsonb;not null" json:"-"`
	ScheduledAt   time.Time         `gorm:"not null;index" json:"scheduled_at"`
	CreatedAt     time.Time         `json:"created_at"`
	Status        string            `gorm:"not null" json:"status"`
	Error         string            `gorm:"type:text" json:"error,omitempty"`
	PipelineRunID *uint             `json:"pipeline_run_id,omitempty"`
	PipelineRun   *PipelineRun      `gorm:"foreignKey:PipelineRunID" json:"pipeline_run,omitempty"`
}
