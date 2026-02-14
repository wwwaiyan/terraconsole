package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/terraconsole/api/internal/middleware"
	"github.com/terraconsole/api/internal/models"
	"gorm.io/gorm"
)

type OrgHandler struct {
	db *gorm.DB
}

func NewOrgHandler(db *gorm.DB) *OrgHandler {
	return &OrgHandler{db: db}
}

func (h *OrgHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	var orgs []models.Organization
	h.db.Joins("JOIN org_members ON org_members.organization_id = organizations.id").
		Where("org_members.user_id = ?", user.ID).
		Find(&orgs)

	writeJSON(w, http.StatusOK, orgs)
}

func (h *OrgHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
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

	org := models.Organization{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Description: req.Description,
		OwnerID:     user.ID,
	}

	tx := h.db.Begin()

	if err := tx.Create(&org).Error; err != nil {
		tx.Rollback()
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Organization name already taken"})
		return
	}

	member := models.OrgMember{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           models.OrgRoleOwner,
	}
	if err := tx.Create(&member).Error; err != nil {
		tx.Rollback()
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create membership"})
		return
	}

	tx.Commit()
	writeJSON(w, http.StatusCreated, org)
}

func (h *OrgHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	user := middleware.GetUser(r)

	var org models.Organization
	if err := h.db.First(&org, "id = ?", orgID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Organization not found"})
		return
	}

	// Check membership
	var member models.OrgMember
	if err := h.db.Where("organization_id = ? AND user_id = ?", orgID, user.ID).First(&member).Error; err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Access denied"})
		return
	}

	writeJSON(w, http.StatusOK, org)
}

func (h *OrgHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	user := middleware.GetUser(r)

	if !h.hasRole(orgID, user.ID, models.OrgRoleOwner, models.OrgRoleAdmin) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Insufficient permissions"})
		return
	}

	var req struct {
		DisplayName *string `json:"display_name"`
		Email       *string `json:"email"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	h.db.Model(&models.Organization{}).Where("id = ?", orgID).Updates(updates)

	var org models.Organization
	h.db.First(&org, "id = ?", orgID)
	writeJSON(w, http.StatusOK, org)
}

func (h *OrgHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	user := middleware.GetUser(r)

	if !h.hasRole(orgID, user.ID, models.OrgRoleOwner) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Only the owner can delete the organization"})
		return
	}

	h.db.Where("id = ?", orgID).Delete(&models.Organization{})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Organization deleted"})
}

func (h *OrgHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")

	var members []models.OrgMember
	h.db.Preload("User").Where("organization_id = ?", orgID).Find(&members)
	writeJSON(w, http.StatusOK, members)
}

func (h *OrgHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	user := middleware.GetUser(r)

	if !h.hasRole(orgID, user.ID, models.OrgRoleOwner, models.OrgRoleAdmin) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Insufficient permissions"})
		return
	}

	var req struct {
		Email string          `json:"email"`
		Role  models.OrgRole  `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	var targetUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&targetUser).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	member := models.OrgMember{
		OrganizationID: orgID,
		UserID:         targetUser.ID,
		Role:           req.Role,
	}

	if err := h.db.Create(&member).Error; err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "User is already a member"})
		return
	}

	h.db.Preload("User").First(&member, "id = ?", member.ID)
	writeJSON(w, http.StatusCreated, member)
}

func (h *OrgHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	memberID := chi.URLParam(r, "memberId")
	user := middleware.GetUser(r)

	if !h.hasRole(orgID, user.ID, models.OrgRoleOwner, models.OrgRoleAdmin) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Insufficient permissions"})
		return
	}

	var req struct {
		Role models.OrgRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	h.db.Model(&models.OrgMember{}).Where("id = ? AND organization_id = ?", memberID, orgID).Update("role", req.Role)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Member role updated"})
}

func (h *OrgHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	memberID := chi.URLParam(r, "memberId")
	user := middleware.GetUser(r)

	if !h.hasRole(orgID, user.ID, models.OrgRoleOwner, models.OrgRoleAdmin) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Insufficient permissions"})
		return
	}

	h.db.Where("id = ? AND organization_id = ?", memberID, orgID).Delete(&models.OrgMember{})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Member removed"})
}

func (h *OrgHandler) hasRole(orgID, userID string, roles ...models.OrgRole) bool {
	var member models.OrgMember
	if err := h.db.Where("organization_id = ? AND user_id = ? AND role IN ?", orgID, userID, roles).First(&member).Error; err != nil {
		return false
	}
	return true
}
