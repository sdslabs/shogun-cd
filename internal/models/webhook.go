package models

import "time"

type Webhook struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Slug      string `gorm:"uniqueIndex;not null;type:varchar(50)"`
	Secret    string `gorm:"not null"`
	Pipeline  string `gorm:"index;not null"`
	Alias     string `gorm:"type:varchar(100)"`
	CreatedBy string `gorm:"type:varchar(100)"`
	IsActive  *bool  `gorm:"default:true"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
