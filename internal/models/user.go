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
	UserColPassword = "password_hash"
	UserColRole     = "role"
	UserColIsActive = "is_active"
	// automatically managed
	UserColCreatedAt = "created_at"
	UserColUpdatedAt = "updated_at"

	// user roles
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	ID           uuid.UUID `gorm:"column:id; type:uuid; default:gen_random_uuid(); primaryKey" json:"id"`
	Email        string    `gorm:"column:email; uniqueIndex; not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash; not null" json:"-"`
	Role         string    `gorm:"column:role; not null; default:user" json:"role"`
	IsActive     *bool     `gorm:"column:is_active; default:true" json:"is_active"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return UserTableName
}
