package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/terraconsole/api/internal/middleware"
	"github.com/terraconsole/api/internal/models"
	"gorm.io/gorm"
)

type RunHandler struct {
	db *gorm.DB
}

func NewRunHandler(db *gorm.DB) *RunHandler {
	return &RunHandler{db: db}
}

func (h *RunHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")

	var runs []models.Run
	h.db.Preload("Creator").
		Where("workspace_id = ?", wsID).
		Order("created_at DESC").
		Limit(50).
		Find(&runs)

	writeJSON(w, http.StatusOK, runs)
}

func (h *RunHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceId")
	user := middleware.GetUser(r)

	var req struct {
		Operation models.RunOperation `json:"operation"`
		Message   string              `json:"message"`
		AutoApply bool                `json:"auto_apply"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	// Get workspace to check lock and auto_apply
	var workspace models.Workspace
	if err := h.db.First(&workspace, "id = ?", wsID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Workspace not found"})
		return
	}

	if workspace.Locked {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Workspace is locked"})
		return
	}

	// Check for pending runs
	var pendingCount int64
	h.db.Model(&models.Run{}).Where("workspace_id = ? AND status IN ?", wsID,
		[]models.RunStatus{models.RunStatusPending, models.RunStatusPlanning, models.RunStatusApplying}).
		Count(&pendingCount)
	if pendingCount > 0 {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Workspace already has an active run"})
		return
	}

	autoApply := req.AutoApply || workspace.AutoApply
	isDestroy := req.Operation == models.RunOperationDestroy

	run := models.Run{
		WorkspaceID:      wsID,
		Status:           models.RunStatusPending,
		Operation:        req.Operation,
		Message:          req.Message,
		IsDestroy:        isDestroy,
		AutoApply:        autoApply,
		TerraformVersion: workspace.TerraformVersion,
		CreatedBy:        user.ID,
	}

	if err := h.db.Create(&run).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create run"})
		return
	}

	// Load creator
	h.db.Preload("Creator").First(&run, "id = ?", run.ID)

	writeJSON(w, http.StatusCreated, run)
}

func (h *RunHandler) Get(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")

	var run models.Run
	if err := h.db.Preload("Creator").First(&run, "id = ?", runID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Run not found"})
		return
	}

	writeJSON(w, http.StatusOK, run)
}

func (h *RunHandler) GetPlanLog(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")

	var run models.Run
	if err := h.db.Select("id, plan_log").First(&run, "id = ?", runID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Run not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"log": run.PlanLog})
}

func (h *RunHandler) GetApplyLog(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")

	var run models.Run
	if err := h.db.Select("id, apply_log").First(&run, "id = ?", runID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Run not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"log": run.ApplyLog})
}

func (h *RunHandler) Approve(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")

	var run models.Run
	if err := h.db.First(&run, "id = ?", runID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Run not found"})
		return
	}

	if run.Status != models.RunStatusNeedsConfirm {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Run is not awaiting confirmation"})
		return
	}

	h.db.Model(&run).Update("status", models.RunStatusApplying)

	writeJSON(w, http.StatusOK, map[string]string{"message": "Run approved, applying..."})
}

func (h *RunHandler) Discard(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")

	var run models.Run
	if err := h.db.First(&run, "id = ?", runID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Run not found"})
		return
	}

	if run.Status != models.RunStatusNeedsConfirm && run.Status != models.RunStatusPlanned {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Run cannot be discarded in current state"})
		return
	}

	now := time.Now()
	h.db.Model(&run).Updates(map[string]interface{}{
		"status":       models.RunStatusDiscarded,
		"completed_at": &now,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "Run discarded"})
}

func (h *RunHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runId")

	var run models.Run
	if err := h.db.First(&run, "id = ?", runID).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Run not found"})
		return
	}

	if run.Status != models.RunStatusPending && run.Status != models.RunStatusPlanning {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Run cannot be cancelled in current state"})
		return
	}

	now := time.Now()
	h.db.Model(&run).Updates(map[string]interface{}{
		"status":       models.RunStatusCancelled,
		"completed_at": &now,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "Run cancelled"})
}
