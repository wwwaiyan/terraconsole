package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/terraconsole/api/internal/middleware"
	"github.com/terraconsole/api/internal/models"
	"github.com/terraconsole/api/internal/services"
	"gorm.io/gorm"
)

type StateHandler struct {
	db        *gorm.DB
	encryptor *services.EncryptionService
}

func NewStateHandler(db *gorm.DB, enc *services.EncryptionService) *StateHandler {
	return &StateHandler{db: db, encryptor: enc}
}

func (h *StateHandler) GetCurrentState(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	var state models.StateVersion
	if err := h.db.Where("workspace_id = ?", wsID).Order("serial DESC").First(&state).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "No state found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(state.State)
}

func (h *StateHandler) ListStateVersions(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	var versions []models.StateVersion
	h.db.Select("id, workspace_id, run_id, serial, lineage, state_hash, outputs, resource_count, created_at, created_by").
		Where("workspace_id = ?", wsID).
		Order("serial DESC").
		Limit(50).
		Find(&versions)

	writeJSON(w, http.StatusOK, versions)
}

func (h *StateHandler) GetStateVersion(w http.ResponseWriter, r *http.Request) {
	versionID := chi.URLParam(r, "versionId")

	var state models.StateVersion
	if err := h.db.First(&state, "id = ?", versionID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "State version not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(state.State)
}

func (h *StateHandler) GetOutputs(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	var state models.StateVersion
	if err := h.db.Select("id, outputs").Where("workspace_id = ?", wsID).Order("serial DESC").First(&state).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "No state found"})
		return
	}

	if state.Outputs == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{})
		return
	}

	var outputs interface{}
	json.Unmarshal([]byte(state.Outputs), &outputs)
	writeJSON(w, http.StatusOK, outputs)
}

// HTTP Backend for Terraform state
// These endpoints implement the Terraform HTTP backend protocol

func (h *StateHandler) HTTPBackendGet(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	var state models.StateVersion
	if err := h.db.Where("workspace_id = ?", wsID).Order("serial DESC").First(&state).Error; err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(state.State)
}

func (h *StateHandler) HTTPBackendPost(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")
	user := middleware.GetUser(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to read body"})
		return
	}

	// Parse to get serial and lineage
	var stateData struct {
		Serial  int    `json:"serial"`
		Lineage string `json:"lineage"`
		Outputs map[string]interface{} `json:"outputs"`
	}
	json.Unmarshal(body, &stateData)

	hash := fmt.Sprintf("%x", sha256.Sum256(body))

	outputsJSON, _ := json.Marshal(stateData.Outputs)

	state := models.StateVersion{
		WorkspaceID:   wsID,
		Serial:        stateData.Serial,
		Lineage:       stateData.Lineage,
		State:         body,
		StateHash:     hash,
		Outputs:       string(outputsJSON),
		ResourceCount: 0,
	}

	if user != nil {
		state.CreatedBy = user.ID
	}

	if err := h.db.Create(&state).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save state"})
		return
	}

	// Update workspace current state
	h.db.Model(&models.Workspace{}).Where("id = ?", wsID).Update("current_state_id", state.ID)

	w.WriteHeader(http.StatusOK)
}

func (h *StateHandler) HTTPBackendLock(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	var workspace models.Workspace
	if err := h.db.First(&workspace, "id = ?", wsID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Workspace not found"})
		return
	}

	if workspace.Locked {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Workspace is locked"})
		return
	}

	h.db.Model(&workspace).Update("locked", true)
	writeJSON(w, http.StatusOK, map[string]string{"status": "locked"})
}

func (h *StateHandler) HTTPBackendUnlock(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	h.db.Model(&models.Workspace{}).Where("id = ?", wsID).Update("locked", false)
	writeJSON(w, http.StatusOK, map[string]string{"status": "unlocked"})
}
