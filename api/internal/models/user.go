package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	FullName  string         `json:"full_name"`
	AvatarURL string         `json:"avatar_url"`
	MFAEnabled bool          `json:"mfa_enabled" gorm:"default:false"`
	MFASecret  string        `json:"-"`
	IsActive   bool          `json:"is_active" gorm:"default:true"`
	LastLoginAt *time.Time   `json:"last_login_at"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

type APIToken struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;index"`
	User        User      `json:"-" gorm:"foreignKey:UserID"`
	Description string    `json:"description"`
	TokenHash   string    `json:"-" gorm:"uniqueIndex;not null"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}
