package database

import (
	"log"

	"github.com/terraconsole/api/internal/config"
	"github.com/terraconsole/api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	log.Println("Connected to database successfully")
	return db
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.APIToken{},
		&models.Organization{},
		&models.OrgMember{},
		&models.Project{},
		&models.Workspace{},
		&models.Variable{},
		&models.VariableSet{},
		&models.VariableSetVariable{},
		&models.VariableSetWorkspace{},
		&models.Run{},
		&models.StateVersion{},
		&models.AuditLog{},
		&models.Notification{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")
}
