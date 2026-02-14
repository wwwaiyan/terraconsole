package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ExecutionMode string

const (
	ExecutionModeLocal  ExecutionMode = "local"
	ExecutionModeAgent  ExecutionMode = "agent"
)

type Workspace struct {
	ID               string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name             string         `json:"name" gorm:"not null;uniqueIndex:idx_workspace_project"`
	Description      string         `json:"description"`
	ProjectID        string         `json:"project_id" gorm:"type:uuid;not null;uniqueIndex:idx_workspace_project"`
	Project          Project        `json:"-" gorm:"foreignKey:ProjectID"`
	TerraformVersion string         `json:"terraform_version" gorm:"default:'latest'"`
	WorkingDirectory string         `json:"working_directory" gorm:"default:'.'"`
	AutoApply        bool           `json:"auto_apply" gorm:"default:false"`
	ExecutionMode    ExecutionMode  `json:"execution_mode" gorm:"type:varchar(20);default:'local'"`
	Locked           bool           `json:"locked" gorm:"default:false"`
	LockedBy         *string        `json:"locked_by" gorm:"type:uuid"`
	LockedAt         *time.Time     `json:"locked_at"`
	VCSRepoURL       string         `json:"vcs_repo_url"`
	VCSBranch        string         `json:"vcs_branch" gorm:"default:'main'"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
	Variables        []Variable     `json:"variables,omitempty" gorm:"foreignKey:WorkspaceID"`
	Runs             []Run          `json:"runs,omitempty" gorm:"foreignKey:WorkspaceID"`
	CurrentStateID   *string        `json:"current_state_id" gorm:"type:uuid"`
}

type VariableCategory string

const (
	VariableCategoryTerraform VariableCategory = "terraform"
	VariableCategoryEnv       VariableCategory = "env"
)

type Variable struct {
	ID          string           `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WorkspaceID string           `json:"workspace_id" gorm:"type:uuid;not null;index"`
	Key         string           `json:"key" gorm:"not null"`
	Value       string           `json:"value"`
	Description string           `json:"description"`
	Category    VariableCategory `json:"category" gorm:"type:varchar(20);not null;default:'terraform'"`
	HCL         bool             `json:"hcl" gorm:"default:false"`
	Sensitive   bool             `json:"sensitive" gorm:"default:false"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type VariableSet struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name           string    `json:"name" gorm:"not null"`
	Description    string    `json:"description"`
	OrganizationID string    `json:"organization_id" gorm:"type:uuid;not null"`
	Global         bool      `json:"global" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Variables      []VariableSetVariable `json:"variables,omitempty" gorm:"foreignKey:VariableSetID"`
}

type VariableSetVariable struct {
	ID            string           `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	VariableSetID string           `json:"variable_set_id" gorm:"type:uuid;not null;index"`
	Key           string           `json:"key" gorm:"not null"`
	Value         string           `json:"value"`
	Description   string           `json:"description"`
	Category      VariableCategory `json:"category" gorm:"type:varchar(20);not null;default:'terraform'"`
	HCL           bool             `json:"hcl" gorm:"default:false"`
	Sensitive     bool             `json:"sensitive" gorm:"default:false"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

type VariableSetWorkspace struct {
	VariableSetID string `json:"variable_set_id" gorm:"primaryKey;type:uuid"`
	WorkspaceID   string `json:"workspace_id" gorm:"primaryKey;type:uuid"`
}

// JSON helper type
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
