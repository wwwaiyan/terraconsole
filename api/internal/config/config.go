package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	EncryptionKey   string
	TerraformDir    string
	WorkingDir      string
	AllowedOrigins  string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://terraconsole:terraconsole@localhost:5432/terraconsole?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production-please"),
		EncryptionKey:  getEnv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef"),
		TerraformDir:   getEnv("TERRAFORM_DIR", "/opt/terraform/versions"),
		WorkingDir:     getEnv("WORKING_DIR", "/var/lib/terraconsole/workspaces"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
