package models

import (
	"time"

	"github.com/google/uuid"
)

// NOTE : the values of constants must be consistent to the "column" in gorm tags.
const (
	// table name
	UserTableName = "users"
	// columns
	UserColID       = "id"
	UserColEmail    = "email"
	UserColPassword = "password"
	UserColRole     = "role"
	UserColIsActive = "is_active"
	// automatically managed
	UserColCreatedAt = "created_at"
	UserColUpdatedAt = "updated_at"
)

type User struct {
	ID       uuid.UUID `gorm:"column:id; type:uuid; default:gen_random_uuid(); primaryKey" json:"id"`
	Email    string    `gorm:"column:email; uniqueIndex; not null" json:"email"`
	Password string    `gorm:"column:password; not null" json:"-"`
	Role     string    `gorm:"column:role; not null" json:"role"`
	IsActive *bool     `gorm:"column:is_active; default:true" json:"is_active"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return UserTableName
}
