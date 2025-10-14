package config

import (
	"os"
)

type Config struct {
	ServerAddress string
	DatabasePath  string
	JWTSecret     string
}

func Load() (*Config, error) {
	// Значения по умолчанию
	cfg := &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8081"),
		DatabasePath:  getEnv("DATABASE_PATH", "./auth.db"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
	}
	
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}