package models

import (
	"time"

	"gorm.io/gorm"
)

type Organization struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	DisplayName string         `json:"display_name"`
	Email       string         `json:"email"`
	Description string         `json:"description"`
	AvatarURL   string         `json:"avatar_url"`
	OwnerID     string         `json:"owner_id" gorm:"type:uuid;not null"`
	Owner       User           `json:"-" gorm:"foreignKey:OwnerID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Members     []OrgMember    `json:"members,omitempty" gorm:"foreignKey:OrganizationID"`
	Projects    []Project      `json:"projects,omitempty" gorm:"foreignKey:OrganizationID"`
}

type OrgRole string

const (
	OrgRoleOwner  OrgRole = "owner"
	OrgRoleAdmin  OrgRole = "admin"
	OrgRoleMember OrgRole = "member"
	OrgRoleViewer OrgRole = "viewer"
)

type OrgMember struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID string    `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex:idx_org_user"`
	UserID         string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_org_user"`
	User           User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Role           OrgRole   `json:"role" gorm:"type:varchar(20);not null;default:'member'"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
