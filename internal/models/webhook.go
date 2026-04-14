package models

import "time"

// NOTE : the values of constants must be consistent to the "column" in gorm tags.
const (
	// table name
	WebhookTableName = "webhooks"
	// columns
	WebhookColID        = "id"
	WebhookColSlug      = "slug"
	WebhookColSecret    = "secret"
	WebhookColPipeline  = "pipeline"
	WebhookColAlias     = "alias"
	WebhookColCreatedBy = "created_by"
	WebhookColIsActive  = "is_active"
	// automatically managed
	WebhookColCreatedAt = "created_at"
	WebhookColUpdatedAt = "updated_at"
)

type Webhook struct {
	ID        string `gorm:"column:id; primaryKey; type:uuid; default:gen_random_uuid()"`
	Slug      string `gorm:"column:slug; uniqueIndex; not null; type:varchar(50)"`
	Secret    string `gorm:"column:secret; not null" json:"-"`
	Pipeline  string `gorm:"column:pipeline; index; not null; type:varchar(50)"`
	Alias     string `gorm:"column:alias; type:varchar(100)"`
	CreatedBy string `gorm:"column:created_by; type:varchar(100)"`
	IsActive  *bool  `gorm:"column:is_active; default:true"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Webhook) TableName() string {
	return WebhookTableName
}
