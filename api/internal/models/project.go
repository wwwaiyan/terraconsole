package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID             string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name           string         `json:"name" gorm:"not null;uniqueIndex:idx_project_org"`
	Description    string         `json:"description"`
	OrganizationID string         `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex:idx_project_org"`
	Organization   Organization   `json:"-" gorm:"foreignKey:OrganizationID"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	Workspaces     []Workspace    `json:"workspaces,omitempty" gorm:"foreignKey:ProjectID"`
}
