package models

import (
	"time"

	"github.com/google/uuid"
)

// NOTE : the values of constants must be consistent to the "column" in gorm tags.
const (
	// table name
	SecretTableName = "secrets"
	// columns
	SecretColID    = "id"
	SecretColName  = "name"
	SecretColValue = "value"
	// automatically managed
	SecretColCreatedAt = "created_at"
	SecretColUpdatedAt = "updated_at"
)

type Secret struct {
	ID    uuid.UUID `gorm:"column:id; type:uuid; default:gen_random_uuid(); primaryKey" json:"-"`
	Name  string    `gorm:"column:name; uniqueIndex; not null"`
	Value string    `gorm:"column:value; not null; type:text" json:"-"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Secret) TableName() string {
	return SecretTableName
}
