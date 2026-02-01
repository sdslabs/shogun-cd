package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/kunalvirwal/shogun-cd/api/http/apiutils"
	"gorm.io/gorm"
)

type User struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email       string     `gorm:"uniqueIndex;not null" json:"email"`
	Password    string     `gorm:"not null" json:"-"`
	Role        string     `gorm:"not null" json:"role"`
	IsActive    *bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at"` // [TODO] discuss viability of last login time
}

// hook to insert default role during user creation
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Role == "" {
		u.Role = string(apiutils.RoleUser)
	}
	return nil
}
