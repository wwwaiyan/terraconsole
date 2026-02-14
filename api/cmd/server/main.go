package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/terraconsole/api/internal/config"
	"github.com/terraconsole/api/internal/database"
	"github.com/terraconsole/api/internal/handlers"
)

func main() {
	cfg := config.Load()

	// Ensure directories exist
	os.MkdirAll(cfg.TerraformDir, 0755)
	os.MkdirAll(cfg.WorkingDir, 0755)

	// Connect to database
	db := database.Connect(cfg)
	database.Migrate(db)

	// Create router
	router := handlers.NewRouter(cfg, db)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("TerraConsole API server starting on %s", addr)
	log.Printf("Terraform versions directory: %s", cfg.TerraformDir)
	log.Printf("Working directory: %s", cfg.WorkingDir)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
