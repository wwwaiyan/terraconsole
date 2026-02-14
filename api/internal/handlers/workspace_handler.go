package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/terraconsole/api/internal/middleware"
	"github.com/terraconsole/api/internal/models"
	"github.com/terraconsole/api/internal/services"
	"gorm.io/gorm"
)

type WorkspaceHandler struct {
	db        *gorm.DB
	encryptor *services.EncryptionService
}

func NewWorkspaceHandler(db *gorm.DB, enc *services.EncryptionService) *WorkspaceHandler {
	return &WorkspaceHandler{db: db, encryptor: enc}
}

func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")

	var workspaces []models.Workspace
	h.db.Where("project_id = ?", projectID).Order("name ASC").Find(&workspaces)
	writeJSON(w, http.StatusOK, workspaces)
}

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")

	var req struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		TerraformVersion string `json:"terraform_version"`
		WorkingDirectory string `json:"working_directory"`
		AutoApply        bool   `json:"auto_apply"`
		VCSRepoURL       string `json:"vcs_repo_url"`
		VCSBranch        string `json:"vcs_branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Name is required"})
		return
	}

	tfVersion := req.TerraformVersion
	if tfVersion == "" {
		tfVersion = "latest"
	}

	workspace := models.Workspace{
		Name:             req.Name,
		Description:      req.Description,
		ProjectID:        projectID,
		TerraformVersion: tfVersion,
		WorkingDirectory: req.WorkingDirectory,
		AutoApply:        req.AutoApply,
		VCSRepoURL:       req.VCSRepoURL,
		VCSBranch:        req.VCSBranch,
	}

	if workspace.WorkingDirectory == "" {
		workspace.WorkingDirectory = "."
	}
	if workspace.VCSBranch == "" {
		workspace.VCSBranch = "main"
	}

	if err := h.db.Create(&workspace).Error; err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Workspace name already exists in this project"})
		return
	}

	writeJSON(w, http.StatusCreated, workspace)
}

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	var workspace models.Workspace
	if err := h.db.First(&workspace, "id = ?", wsID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Workspace not found"})
		return
	}

	writeJSON(w, http.StatusOK, workspace)
}

func (h *WorkspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	var req struct {
		Name             *string `json:"name"`
		Description      *string `json:"description"`
		TerraformVersion *string `json:"terraform_version"`
		WorkingDirectory *string `json:"working_directory"`
		AutoApply        *bool   `json:"auto_apply"`
		VCSRepoURL       *string `json:"vcs_repo_url"`
		VCSBranch        *string `json:"vcs_branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.TerraformVersion != nil {
		updates["terraform_version"] = *req.TerraformVersion
	}
	if req.WorkingDirectory != nil {
		updates["working_directory"] = *req.WorkingDirectory
	}
	if req.AutoApply != nil {
		updates["auto_apply"] = *req.AutoApply
	}
	if req.VCSRepoURL != nil {
		updates["vcs_repo_url"] = *req.VCSRepoURL
	}
	if req.VCSBranch != nil {
		updates["vcs_branch"] = *req.VCSBranch
	}

	h.db.Model(&models.Workspace{}).Where("id = ?", wsID).Updates(updates)

	var workspace models.Workspace
	h.db.First(&workspace, "id = ?", wsID)
	writeJSON(w, http.StatusOK, workspace)
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")
	h.db.Where("id = ?", wsID).Delete(&models.Workspace{})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Workspace deleted"})
}

func (h *WorkspaceHandler) Lock(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")
	user := middleware.GetUser(r)

	var workspace models.Workspace
	if err := h.db.First(&workspace, "id = ?", wsID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Workspace not found"})
		return
	}

	if workspace.Locked {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Workspace is already locked"})
		return
	}

	h.db.Model(&workspace).Updates(map[string]interface{}{
		"locked":    true,
		"locked_by": user.ID,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Workspace locked"})
}

func (h *WorkspaceHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	h.db.Model(&models.Workspace{}).Where("id = ?", wsID).Updates(map[string]interface{}{
		"locked":    false,
		"locked_by": nil,
		"locked_at": nil,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Workspace unlocked"})
}

// Variable handlers
func (h *WorkspaceHandler) ListVariables(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	var variables []models.Variable
	h.db.Where("workspace_id = ?", wsID).Order("key ASC").Find(&variables)

	// Mask sensitive values
	for i, v := range variables {
		if v.Sensitive {
			variables[i].Value = "***SENSITIVE***"
		}
	}

	writeJSON(w, http.StatusOK, variables)
}

func (h *WorkspaceHandler) CreateVariable(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	var req struct {
		Key         string                  `json:"key"`
		Value       string                  `json:"value"`
		Description string                  `json:"description"`
		Category    models.VariableCategory `json:"category"`
		HCL         bool                    `json:"hcl"`
		Sensitive   bool                    `json:"sensitive"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	value := req.Value
	if req.Sensitive {
		encrypted, err := h.encryptor.Encrypt(value)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to encrypt value"})
			return
		}
		value = encrypted
	}

	variable := models.Variable{
		WorkspaceID: wsID,
		Key:         req.Key,
		Value:       value,
		Description: req.Description,
		Category:    req.Category,
		HCL:         req.HCL,
		Sensitive:   req.Sensitive,
	}

	if err := h.db.Create(&variable).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create variable"})
		return
	}

	if variable.Sensitive {
		variable.Value = "***SENSITIVE***"
	}

	writeJSON(w, http.StatusCreated, variable)
}

func (h *WorkspaceHandler) UpdateVariable(w http.ResponseWriter, r *http.Request) {
	varID := chi.URLParam(r, "variableId")

	var existing models.Variable
	if err := h.db.First(&existing, "id = ?", varID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Variable not found"})
		return
	}

	var req struct {
		Value       *string `json:"value"`
		Description *string `json:"description"`
		HCL         *bool   `json:"hcl"`
		Sensitive   *bool   `json:"sensitive"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.HCL != nil {
		updates["hcl"] = *req.HCL
	}
	if req.Sensitive != nil {
		updates["sensitive"] = *req.Sensitive
	}
	if req.Value != nil {
		value := *req.Value
		isSensitive := existing.Sensitive
		if req.Sensitive != nil {
			isSensitive = *req.Sensitive
		}
		if isSensitive {
			encrypted, err := h.encryptor.Encrypt(value)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to encrypt"})
				return
			}
			value = encrypted
		}
		updates["value"] = value
	}

	h.db.Model(&models.Variable{}).Where("id = ?", varID).Updates(updates)

	var variable models.Variable
	h.db.First(&variable, "id = ?", varID)
	if variable.Sensitive {
		variable.Value = "***SENSITIVE***"
	}
	writeJSON(w, http.StatusOK, variable)
}

func (h *WorkspaceHandler) DeleteVariable(w http.ResponseWriter, r *http.Request) {
	varID := chi.URLParam(r, "variableId")
	h.db.Where("id = ?", varID).Delete(&models.Variable{})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Variable deleted"})
}
