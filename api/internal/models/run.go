package models

import (
	"time"
)

type RunStatus string

const (
	RunStatusPending       RunStatus = "pending"
	RunStatusPlanning      RunStatus = "planning"
	RunStatusPlanned       RunStatus = "planned"
	RunStatusNeedsConfirm  RunStatus = "needs_confirmation"
	RunStatusApplying      RunStatus = "applying"
	RunStatusApplied       RunStatus = "applied"
	RunStatusErrored       RunStatus = "errored"
	RunStatusCancelled     RunStatus = "cancelled"
	RunStatusDiscarded     RunStatus = "discarded"
	RunStatusPlanOnly      RunStatus = "planned_and_finished"
)

type RunOperation string

const (
	RunOperationPlan    RunOperation = "plan"
	RunOperationApply   RunOperation = "plan_and_apply"
	RunOperationDestroy RunOperation = "destroy"
	RunOperationRefresh RunOperation = "refresh"
)

type Run struct {
	ID               string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WorkspaceID      string       `json:"workspace_id" gorm:"type:uuid;not null;index"`
	Workspace        Workspace    `json:"-" gorm:"foreignKey:WorkspaceID"`
	Status           RunStatus    `json:"status" gorm:"type:varchar(30);not null;default:'pending'"`
	Operation        RunOperation `json:"operation" gorm:"type:varchar(20);not null"`
	Message          string       `json:"message"`
	IsDestroy        bool         `json:"is_destroy" gorm:"default:false"`
	AutoApply        bool         `json:"auto_apply" gorm:"default:false"`
	TerraformVersion string       `json:"terraform_version"`
	CreatedBy        string       `json:"created_by" gorm:"type:uuid;not null"`
	Creator          User         `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	PlanLog          string       `json:"-" gorm:"type:text"`
	PlanJSON         string       `json:"-" gorm:"type:text"`
	ApplyLog         string       `json:"-" gorm:"type:text"`
	ResourcesAdded   int          `json:"resources_added" gorm:"default:0"`
	ResourcesChanged int          `json:"resources_changed" gorm:"default:0"`
	ResourcesDeleted int          `json:"resources_deleted" gorm:"default:0"`
	StartedAt        *time.Time   `json:"started_at"`
	PlanCompletedAt  *time.Time   `json:"plan_completed_at"`
	AppliedAt        *time.Time   `json:"applied_at"`
	CompletedAt      *time.Time   `json:"completed_at"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

type StateVersion struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WorkspaceID  string    `json:"workspace_id" gorm:"type:uuid;not null;index"`
	RunID        *string   `json:"run_id" gorm:"type:uuid"`
	Serial       int       `json:"serial" gorm:"not null"`
	Lineage      string    `json:"lineage"`
	State        []byte    `json:"-" gorm:"type:bytea"`
	StateHash    string    `json:"state_hash"`
	Outputs      string    `json:"outputs" gorm:"type:text"`
	ResourceCount int      `json:"resource_count" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by" gorm:"type:uuid"`
}

type AuditLog struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID string    `json:"organization_id" gorm:"type:uuid;not null;index"`
	UserID         string    `json:"user_id" gorm:"type:uuid;not null"`
	User           User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Action         string    `json:"action" gorm:"not null"`
	ResourceType   string    `json:"resource_type" gorm:"not null"`
	ResourceID     string    `json:"resource_id"`
	ResourceName   string    `json:"resource_name"`
	Details        string    `json:"details" gorm:"type:text"`
	IPAddress      string    `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
}

type Notification struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Title     string    `json:"title" gorm:"not null"`
	Message   string    `json:"message"`
	Type      string    `json:"type" gorm:"type:varchar(30)"`
	Read      bool      `json:"read" gorm:"default:false"`
	LinkURL   string    `json:"link_url"`
	CreatedAt time.Time `json:"created_at"`
}
