package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/terraconsole/api/internal/config"
	"github.com/terraconsole/api/internal/middleware"
	"github.com/terraconsole/api/internal/services"
	"gorm.io/gorm"
	"strings"
)

func NewRouter(cfg *config.Config, db *gorm.DB) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(cfg.AllowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	encryptor := services.NewEncryptionService(cfg.EncryptionKey)

	// Handlers
	authHandler := NewAuthHandler(db, cfg, encryptor)
	orgHandler := NewOrgHandler(db)
	projectHandler := NewProjectHandler(db)
	workspaceHandler := NewWorkspaceHandler(db, encryptor)
	runHandler := NewRunHandler(db)
	stateHandler := NewStateHandler(db, encryptor)
	tfVersionHandler := NewTFVersionHandler(cfg)

	// Health check
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"version": "1.0.0",
			"service": "terraconsole",
		})
	})

	// Auth routes (public)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/login", authHandler.Login)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg, db))

		// Auth
		r.Get("/api/auth/me", authHandler.GetMe)
		r.Post("/api/auth/refresh", authHandler.RefreshToken)
		r.Post("/api/auth/mfa/setup", authHandler.SetupMFA)
		r.Post("/api/auth/mfa/verify", authHandler.VerifyMFA)
		r.Post("/api/auth/mfa/disable", authHandler.DisableMFA)

		// Terraform versions
		r.Get("/api/terraform/versions", tfVersionHandler.ListVersions)
		r.Get("/api/terraform/versions/installed", tfVersionHandler.ListInstalledVersions)
		r.Post("/api/terraform/versions/{version}/install", tfVersionHandler.InstallVersion)

		// Organizations
		r.Route("/api/organizations", func(r chi.Router) {
			r.Get("/", orgHandler.List)
			r.Post("/", orgHandler.Create)
			r.Route("/{orgId}", func(r chi.Router) {
				r.Get("/", orgHandler.Get)
				r.Put("/", orgHandler.Update)
				r.Delete("/", orgHandler.Delete)

				// Members
				r.Get("/members", orgHandler.ListMembers)
				r.Post("/members", orgHandler.AddMember)
				r.Put("/members/{memberId}", orgHandler.UpdateMember)
				r.Delete("/members/{memberId}", orgHandler.RemoveMember)

				// Projects
				r.Get("/projects", projectHandler.List)
				r.Post("/projects", projectHandler.Create)
			})
		})

		// Projects
		r.Route("/api/projects/{projectId}", func(r chi.Router) {
			r.Get("/", projectHandler.Get)
			r.Put("/", projectHandler.Update)
			r.Delete("/", projectHandler.Delete)

			// Workspaces
			r.Get("/workspaces", workspaceHandler.List)
			r.Post("/workspaces", workspaceHandler.Create)
		})

		// Workspaces
		r.Route("/api/workspaces/{workspaceId}", func(r chi.Router) {
			r.Get("/", workspaceHandler.Get)
			r.Put("/", workspaceHandler.Update)
			r.Delete("/", workspaceHandler.Delete)
			r.Post("/lock", workspaceHandler.Lock)
			r.Post("/unlock", workspaceHandler.Unlock)

			// Variables
			r.Get("/variables", workspaceHandler.ListVariables)
			r.Post("/variables", workspaceHandler.CreateVariable)
			r.Put("/variables/{variableId}", workspaceHandler.UpdateVariable)
			r.Delete("/variables/{variableId}", workspaceHandler.DeleteVariable)

			// Runs
			r.Get("/runs", runHandler.List)
			r.Post("/runs", runHandler.Create)

			// State
			r.Get("/state", stateHandler.GetCurrentState)
			r.Get("/state-versions", stateHandler.ListStateVersions)
			r.Get("/state-versions/{versionId}", stateHandler.GetStateVersion)
			r.Get("/outputs", stateHandler.GetOutputs)
		})

		// Runs
		r.Route("/api/runs/{runId}", func(r chi.Router) {
			r.Get("/", runHandler.Get)
			r.Get("/plan-log", runHandler.GetPlanLog)
			r.Get("/apply-log", runHandler.GetApplyLog)
			r.Post("/approve", runHandler.Approve)
			r.Post("/discard", runHandler.Discard)
			r.Post("/cancel", runHandler.Cancel)
		})

		// State HTTP Backend (for terraform remote state)
		r.Route("/api/state/{workspaceId}", func(r chi.Router) {
			r.Get("/", stateHandler.HTTPBackendGet)
			r.Post("/", stateHandler.HTTPBackendPost)
			r.HandleFunc("/lock", stateHandler.HTTPBackendLock)
			r.HandleFunc("/unlock", stateHandler.HTTPBackendUnlock)
		})
	})

	return r
}
