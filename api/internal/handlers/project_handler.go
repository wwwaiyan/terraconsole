package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/terraconsole/api/internal/middleware"
	"github.com/terraconsole/api/internal/models"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	db *gorm.DB
}

func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{db: db}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")

	var projects []models.Project
	h.db.Where("organization_id = ?", orgID).Order("name ASC").Find(&projects)
	writeJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	user := middleware.GetUser(r)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Name is required"})
		return
	}

	// Check org membership
	var member models.OrgMember
	if err := h.db.Where("organization_id = ? AND user_id = ?", orgID, user.ID).First(&member).Error; err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Access denied"})
		return
	}

	project := models.Project{
		Name:           req.Name,
		Description:    req.Description,
		OrganizationID: orgID,
	}

	if err := h.db.Create(&project).Error; err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Project name already exists in this organization"})
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")

	var project models.Project
	if err := h.db.First(&project, "id = ?", projectID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Project not found"})
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
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

	h.db.Model(&models.Project{}).Where("id = ?", projectID).Updates(updates)

	var project models.Project
	h.db.First(&project, "id = ?", projectID)
	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	h.db.Where("id = ?", projectID).Delete(&models.Project{})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Project deleted"})
}
